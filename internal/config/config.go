package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type key string

const (
	KeyUUID   = key("uuid")
	KeyLogger = key("logger")
)

type Config struct {
	Service  Service
	Postgres Postgres
	Logger   Logger
	Platform Platform
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

type Logger struct {
	Host string `env:"LOGGER_SERVICE_HOST"`
	Port string `env:"LOGGER_SERVICE_PORT"`
}

type Platform struct {
	Env string `env:"ENV"`
}

func MustLoad() *Config {
	cfg := &Config{}
	err := cleanenv.ReadEnv(cfg)

	if err != nil {
		log.Fatalf("Can not read env variables: %s", err)
	}

	return cfg
}
