# Habits Bot — Техническая документация

## Карта проекта

```
habits-bot/
├── cmd/bot/main.go              # Точка входа: инициализация → запуск → graceful shutdown
├── internal/
│   ├── config/config.go         # Конфигурация: чтение .env, MustLoad(), DSN
│   └── bot/
│       ├── bot.go               # Bot struct, New(), цикл long polling (Run)
│       ├── telegram.go          # Типы Telegram API (Update, Message...), getUpdates()
│       └── handler.go           # Роутинг команд (/start...), sendMessage(), reply()
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
    Database {             ← DB_* (пока не используется)
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

## Логгер (logger.go)

| Режим (ENV) | Handler | Уровень | Формат |
|-------------|---------|---------|--------|
| `local`, `dev` | TextHandler | Debug | `time=... level=DEBUG msg="..." key=value` |
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
      ├─ upd.Message == nil? → return (callback, не обрабатываем пока)
      ├─ Text == ""?         → return
      ├─ начинается с "/"?
      │   ├─ /start → reply(chatID, msgWelcome)
      │   └─ другие → reply(chatID, msgUnknownCmd)
      └─ обычный текст → reply(chatID, msgUseCommands)
          │
          ▼
    reply(chatID, text)
      └─ sendMessage(chatID, text)
           ├─ MessageDTO → json.Marshal
           ├─ POST /sendMessage (application/json)
           └─ io.Copy(io.Discard) — читаем и выбрасываем тело ответа
```

---

## Планы по развитию (roadmap)

| Этап | Что добавляем | Готово |
|------|---------------|:------:|
| 1 | Каркас проекта, конфиг, логгер | ✅ |
| 2 | Long polling, команды бота | ✅ |
| 3 | PostgreSQL, миграции, модель User + Habit | ⬜ |
| 4 | CRUD привычек через бота | ⬜ |
| 5 | Inline-клавиатуры, callback_query | ⬜ |
| 6 | REST API | ⬜ |
| 7 | RabbitMQ — отложенные задачи (напоминания) | ⬜ |
| 8 | Redis — кеширование | ⬜ |
| 9 | Социальные челленджи | ⬜ |
