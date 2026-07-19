# Habits Bot — Техническая документация

## Карта проекта

```
habits-bot/
├── cmd/bot/main.go              # Точка входа: конфиг → логгер → БД → бот → graceful shutdown
├── internal/
│   ├── config/config.go         # Конфигурация: чтение .env, MustLoad(), DSN
│   ├── domain/
│   │   ├── models.go            # Доменные модели: User, Habit
│   │   └── repository.go        # Интерфейсы: UserRepository, HabitRepository
│   ├── storage/
│   │   ├── storage.go           # Подключение к БД, goose-миграции, сборка репозиториев
│   │   ├── migrations/          # SQL-миграции (вшиваются в бинарник через embed)
│   │   │   ├── 001_users.sql
│   │   │   └── 002_habits.sql
│   │   └── postgres/
│   │       ├── user.go          # UserRepository реализация для PostgreSQL
│   │       └── habit.go         # HabitRepository реализация для PostgreSQL
│   └── bot/
│       ├── bot.go               # Bot struct, New(), цикл long polling (Run)
│       ├── telegram.go          # Типы Telegram API (Update, Message...), getUpdates()
│       ├── handler.go           # handleUpdate, ensureUser, routeMessage, parseCommands
│       ├── messages.go          # Все текстовые константы ответов бота
│       ├── habits.go            # Обработчики CRUD привычек + /help
│       ├── keyboards.go         # Inline-клавиатуры (habitList, action, confirmDelete)
│       └── sender.go            # sendMessage, editMessage, reply, MessageDTO
├── pkg/logger/logger.go         # Логгер: dev (TextHandler) / prod (JSONHandler)
├── docker-compose.yml           # PostgreSQL 17 + Redis 7 + RabbitMQ 3
├── Makefile                     # make up / make down / make build
├── .env.example                 # Образец переменных окружения
└── go.mod
```

---

## Как работает бот (Long Polling)

```
main()                     Bot.Run()                         Telegram API
  │                           │                                   │
  │  1. Загружаем конфиг      │                                   │
  │  2. Создаём логгер        │                                   │
  │  3. bot.New(token, log)   │                                   │
  │                           │                                   │
  │  bot.Run(ctx) ──────────→ │                                   │
  │                           │  GET /getUpdates?offset=0&t=10 →  │
  │                           │ ←─ {"ok":true,"result":[...]} ─── │
  │                           │                                   │
  │                           │  Для каждого Update:              │
  │                           │    handleUpdate(upd)              │
  │                           │      ├─ команда? → /start и т.д.  │
  │                           │      └─ sendMessage(chatID, text) │
  │                           │         POST /sendMessage ──────→ │
  │                           │                                   │
  │                           │  GET /getUpdates?offset=101...→   │
  │    (ждём Ctrl+C)          │                                   │
  │                           │                                   │
  │  ctx.Done() ←──────────── │  return ctx.Err()                 │
  │  "shutting down"          │                                   │
```

---

## Telegram API: формат JSON и маппинг в Go

### Что присылает Telegram (GET /getUpdates)

```json
{
    "ok": true,
    "result": [
        {
            "update_id": 100,
            "message": {
                "message_id": 42,
                "text": "/start",
                "chat": {
                    "id": 123456789
                },
                "from": {
                    "id": 123456789,
                    "username": "beloslav13"
                }
            }
        }
    ]
}
```

### Go-типы, в которые это парсится

| JSON-поле | Тип Go | Структура | Пояснение |
|-----------|--------|-----------|-----------|
| `ok` | `bool` | `Response.Ok` | Признак успешного ответа |
| `result` | `[]Update` | `Response.Result` | Массив обновлений |
| `result[].update_id` | `int` | `Update.UpdateID` | Уникальный ID обновления |
| `result[].message` | `*Message` | `Update.Message` | Сообщение (nil для callback) |
| `result[].callback_query` | `*CallbackQuery` | `Update.CallbackQuery` | Нажатие кнопки (nil для сообщений) |
| `message.text` | `string` | `Message.Text` | Текст сообщения |
| `message.chat.id` | `int` | `Chat.ID` | ID чата (куда отвечать) |
| `message.from.username` | `string` | `User.Username` | @username отправителя |

### Что мы отправляем (POST /sendMessage)

```json
{
    "chat_id": 123456789,
    "text": "Привет! Я трекер привычек."
}
```

| JSON-поле | Тип Go | Структура | Пояснение |
|-----------|--------|-----------|-----------|
| `chat_id` | `int` | `MessageDTO.ChatID` | Куда отправить |
| `text` | `string` | `MessageDTO.Text` | Текст сообщения |

---

## Конфигурация (config.go)

```
Config {
    Env      "local"        ← ENV (управляет режимом логгера)
    Bot {
        Token  "123:abc"   ← BOT_TOKEN (обязательное поле)
    }
    Database {             ← DB_*
        Host, Port, User, Password, Name, SSLMode
    }
    Redis {                ← REDIS_* (пока не используется)
        Host, Port, Password
    }
    RabbitMQ {             ← RABBITMQ_* (пока не используется)
        Host, Port, User, Password
    }
}
```

- `MustLoad()`: `godotenv.Load()` (загрузка `.env`) → `cleanenv.ReadEnv()` (заполнение структур)
- `DatabaseConfig.DSN()`: собирает строку подключения `postgres://user:pass@host:port/db?sslmode=...`

---

## Архитектура хранения (Repository + goose-миграции)

### Слои

```
┌─────────────────────────────┐
│  main.go / bot / handler    │  ← бизнес-логика (не знает SQL)
├─────────────────────────────┤
│  domain/repository.go       │  ← интерфейсы: UserRepository, HabitRepository
├─────────────────────────────┤
│  storage/postgres/          │  ← реализации для PostgreSQL (UserRepo, HabitRepo)
├─────────────────────────────┤
│  storage/storage.go         │  ← sql.Open + goose.Up + сборка в Storage{}
└─────────────────────────────┘
```

### Domain-модели (`internal/domain/models.go`)

| Модель | Таблица | Поля |
|--------|---------|------|
| `User` | `users` | ID, TelegramID, FirstName, LastName*, Username*, Language*, IsPremium, CreatedAt, UpdatedAt |
| `Habit` | `habits` | ID, UserID (FK→users), Name, Description*, CreatedAt, UpdatedAt |

*nullable поля → `*string` в Go (nil = NULL в БД)

### Repository (паттерн)

- Интерфейсы лежат в `domain/` — скрывают SQL от бизнес-логики
- Реализации в `storage/postgres/` — конкретные SQL-запросы
- При смене БД (MySQL, Mongo) меняется только `postgres/` → `mysql/`, бизнес-логика не трогается

### Миграции (goose)

Миграции лежат в `internal/storage/migrations/` и **вшиваются в бинарник** через `//go:embed`. Не нужно таскать SQL-файлы рядом с программой.

```sql
-- +goose Up     ← что выполнить при накатывании
CREATE TABLE users (...);

-- +goose Down   ← что выполнить при откате
DROP TABLE users;
```

При старте `storage.New()` вызывает `goose.Up()` — применяет все неприменённые миграции. Таблица `goose_db_version` в БД хранит историю. Повторный запуск не ломает уже применённые миграции.

### Жизненный цикл в main.go

```
main()
  ├─ config.MustLoad()                        ← читаем .env
  ├─ logger.New(cfg.Env)                      ← создаём логгер
  ├─ storage.New(cfg.DataBase.DSN())          ← подключаемся к БД + миграции
  │    ├─ sql.Open("postgres", dsn)            ← регистрируем драйвер
  │    ├─ db.Ping()                            ← проверяем соединение
  │    ├─ goose.Up(db, "migrations")           ← накатываем миграции
  │    └─ Storage{User: ..., Habit: ...}       ← собираем репозитории
  ├─ bot.New(token, log)                      ← создаём бота
  ├─ bot.Run(ctx)                             ← запускаем long polling
  └─ store.Close()                            ← закрываем соединение с БД (defer)
```

---

## Логгер (logger.go)

| Режим (ENV) | Handler | Уровень | Формат |
|-------------|---------|---------|--------|
| `local`, `dev` | TextHandler | Debug | `time=... level=DEBUG source=bot.go:40 msg="..."` |
| всё остальное | JSONHandler | Info | `{"time":"...","level":"INFO","msg":"..."}` |

- `log.Info("starting bot", "env", cfg.Env)` → в dev: `time=... level=INFO msg="starting bot" env=local`
- `log.Error("getUpdates error", "error", err)` → добавляет поле `error`

---

## Поток обработки сообщения

```
Telegram присылает Update
          │
          ▼
    getUpdates()
      ├─ HTTP GET
      ├─ io.ReadAll → []byte
      ├─ json.Unmarshal → Response
      └─ возвращает response.Result ([]Update)
          │
          ▼
    Run() — итерирует по []Update
      offset = upd.UpdateID + 1
      handleUpdate(ctx, upd)
          │
          ▼
    handleUpdate()
      ├─ CallbackQuery != nil? → routeCallback() (обработка нажатий кнопок)
      ├─ Message == nil? → return
      ├─ ensureUser() — находит или создаёт пользователя в БД
      ├─ parseCommands() — извлекает команду и аргументы
      └─ routeMessage(userID, chatID, text)
           ├─ /start      → приветствие + /help
           ├─ /help       → список команд
           ├─ /habits     → habitList: выводит список привычек
           ├─ /newhabit N → habitCreate: создаёт привычку
           ├─ /edithabit ID N → habitUpdate: изменяет название
           ├─ /deletehabit ID → habitDelete: удаляет (после проверки владельца)
           └─ остальное   → msgUnknownCommand

Все CRUD-операции проверяют:
- Валидацию названия (не пустое, ≤ 256 символов)
- Существование привычки (ErrHabitNotFound)
- Принадлежность пользователю (userID != h.UserID)
          │
          ▼
    reply(chatID, text)
      └─ sendMessage(chatID, text)
           ├─ MessageDTO → json.Marshal
           ├─ POST /sendMessage (application/json)
           └─ io.Copy(io.Discard) — читаем и выбрасываем тело ответа
```

---

## Inline-клавиатуры и Callback-и

### Формат callback_data

```
entity_action_id_userID

habit_view_5_1              ← выбор привычки (показать действия)
habit_edit_5_1             ← запрос на изменение (показать подсказку)
habit_delete_5_1           ← запрос на удаление (показать подтверждение)
habit_confirmdelete_5_1    ← подтвердить удаление
habit_new_0_1              ← запрос на создание (показать подсказку)
habits_list_0_1            ← вернуться к списку привычек
```

### Как это выглядит (запрос/ответ)

**Отправляем (/sendMessage):**
```json
{
    "chat_id": 123,
    "text": "Твои привычки:",
    "reply_markup": {
        "inline_keyboard": [
            [{"text": "🏃 спорт", "callback_data": "habit_view_5_1"}],
            [{"text": "📖 чтение", "callback_data": "habit_view_8_1"}],
            [{"text": "➕ Создать", "callback_data": "habit_new_0_1"}]
        ]
    }
}
```

**Telegram присылает (нажата кнопка):**
```json
{
    "callback_query": {
        "id": "abc123",
        "from": {"id": 123, "username": "user"},
        "message": {"message_id": 42, "chat": {"id": 123}},
        "data": "habit_confirmdelete_5_1"
    }
}
```

**Редактируем сообщение (/editMessageText):**
```json
{
    "chat_id": 123,
    "message_id": 42,
    "text": "Привычка удалена ✅",
    "reply_markup": {"inline_keyboard": [[...]]}
}
```

### Клавиатуры

| Клавиатура | Где | Кнопки |
|-----------|-----|--------|
| `habitListKeyboard` | /habits, список | Привычки (по 1 в строке) + ➕ Создать |
| `habitActionKeyboard` | Выбор привычки | ✏️ Изменить, 🗑 Удалить + ⬅️ Назад |
| `confirmDeleteKeyboard` | Подтверждение удаления | ✅ Да, удалить + ❌ Нет |

---

## Планы по развитию (roadmap)

| Этап | Что добавляем | Готово |
|------|---------------|:------:|
| 1 | Каркас проекта, конфиг, логгер | ✅ |
| 2 | Long polling, команды бота | ✅ |
| 3 | PostgreSQL, миграции, модель User + Habit | ✅ |
| 4 | CRUD привычек через бота | ✅ |
| 5 | Inline-клавиатуры, callback_query | 🟡 (клавиатуры ✅, стейт-машина ⬜) |
| 6 | REST API | ⬜ |
| 7 | RabbitMQ — отложенные задачи (напоминания) | ⬜ |
| 8 | Redis — кеширование | ⬜ |
| 9 | Социальные челленджи | ⬜ |
