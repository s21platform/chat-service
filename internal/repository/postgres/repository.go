package postgres

import (
	"chat-service/internal/config"
	"chat-service/internal/model"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	connection *sqlx.DB
}

func New(cfg *config.Config) *Repository {
	conStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.Host, cfg.Postgres.Port)

	db, err := sqlx.Connect("postgres", conStr)
	if err != nil {
		log.Fatalf("failed to connect DB")
	}

	return &Repository{db}
}

func (r *Repository) Close() {
	r.connection.Close()
}

func (r *Repository) GetChat(chatUUID string) (*[]model.Message, error) {
	var messages []model.Message

	query := `
		SELECT sender_uuid, content, created_at FROM messages
		WHERE chat_uuid = $1
		ORDER BY created_at DESC
		LIMIT 15; 
	`
	err := r.connection.Select(&messages, query)
	if err != nil {
		return nil, fmt.Errorf("r.connection.Select in GetChat: %v", err)
	}

	return &messages, nil
}
