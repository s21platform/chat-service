package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type key string

const KeyUUID key = key("uuid")

type Config struct {
	Service  Service
	Postgres Postgres
}

type Service struct {
	Port string `env:"CHAT_SERVICE_PORT"`
}

type Postgres struct {
	User     string `env:"CHAT_SERVICE_POSTGRES_USER"`
	Password string `env:"CHAT_SERVICE_POSTGRES_PASSWORD"`
	Database string `env:"CHAT_SERVICE_POSTGRES_DB"`
	Host     string `env:"CHAT_SERVICE_POSTGRES_HOST"`
	Port     string `env:"CHAT_SERVICE_POSTGRES_PORT"`
}

func MustLoad() *Config {
	cfg := &Config{}
	err := cleanenv.ReadEnv(cfg)

	if err != nil {
		log.Fatalf("Can not read env variables: %s", err)
	}

	return cfg
}
