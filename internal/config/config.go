package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Service     Service
	Postgres    Postgres
	Metrics     Metrics
	Logger      Logger
	Platform    Platform
	UserService UserService
	Kafka       Kafka
}

type Service struct {
	Port string `env:"CHAT_SERVICE_PORT"`
	Name string `env:"CHAT_SERVICE_NAME"`
}

type Postgres struct {
	User     string `env:"CHAT_SERVICE_POSTGRES_USER"`
	Password string `env:"CHAT_SERVICE_POSTGRES_PASSWORD"`
	Database string `env:"CHAT_SERVICE_POSTGRES_DB"`
	Host     string `env:"CHAT_SERVICE_POSTGRES_HOST"`
	Port     string `env:"CHAT_SERVICE_POSTGRES_PORT"`
}

type Metrics struct {
	Host string `env:"GRAFANA_HOST"`
	Port int    `env:"GRAFANA_PORT"`
}

type Logger struct {
	Port string `env:"LOGGER_SERVICE_PORT"`
	Host string `env:"LOGGER_SERVICE_HOST"`
}

type Platform struct {
	Env string `env:"ENV"`
}

type UserService struct {
	Host string `env:"USER_SERVICE_HOST"`
	Port string `env:"USER_SERVICE_PORT"`
}

type Kafka struct {
	Host           string `env:"KAFKA_HOST"`
	Port           string `env:"KAFKA_PORT"`
	UpdateNickname string `env:"KAFKA_UPDATE_NICKNAME"`
}

func MustLoad() *Config {
	cfg := &Config{}

	// 1) читаем .env-файл (урок: .env ← cleanenv)
	if err := cleanenv.ReadConfig(".env", cfg); err != nil {
		log.Fatalf("failed to read .env file: %s", err)
	}

	// 2) потом пере‑перезаписываем из реального окружения, если нужно
	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("failed to read env variables: %s", err)
	}

	return cfg
}
