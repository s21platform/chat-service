package postgres

import (
	"chat-service/internal/config"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	connection *sqlx.DB
}

func connect(cfg *config.Config) *Repository {
	conStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.Host, cfg.Postgres.Port)

	db, err := sqlx.Connect("postgres", conStr)
	if err != nil {
		log.Fatalf("failed to connect DB")
	}

	return &Repository{db}
}

func New(cfg *config.Config) *Repository {
	return connect(cfg)
}

func (r *Repository) Close() {
	r.connection.Close()
}
