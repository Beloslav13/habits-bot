package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Config содержит все настройки приложения, загружаемые из переменных окружения.
type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	Bot      BotConfig
	DataBase DatabaseConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
}

// MustLoad загружает конфигурацию из .env и переменных окружения.
// Вызывает panic, если обязательные поля (например, BOT_TOKEN) не заданы.
func MustLoad() *Config {
	cfg := &Config{}
	_ = godotenv.Load()

	if err := cleanenv.ReadEnv(cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return cfg
}

type BotConfig struct {
	Token string `env:"BOT_TOKEN" env-required:"true"`
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSLMODE"`
}

// DSN собирает строку подключения к PostgreSQL в формате URL.
func (cfg DatabaseConfig) DSN() string {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host,
		cfg.Port, cfg.Name, cfg.SSLMode)
	return dsn
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST"`
	Port     string `env:"REDIS_PORT"`
	Password string `env:"REDIS_PASSWORD"`
}

type RabbitMQConfig struct {
	Host     string `env:"RABBITMQ_HOST"`
	Port     string `env:"RABBITMQ_PORT"`
	User     string `env:"RABBITMQ_USER"`
	Password string `env:"RABBITMQ_PASSWORD"`
}
