package postgres

import (
	"fmt"
	"log"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"

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
		log.Fatalf("failed to connect to db: %v", err)
	}

	return &Repository{db}
}

func (r *Repository) Close() {
	_ = r.connection.Close()
}

func (r *Repository) GetChat(chatUUID string) (*[]model.Message, error) {
	var messages []model.Message

	query := `
		SELECT sender_uuid, content, created_at FROM messages
		WHERE chat_uuid = $1
		ORDER BY created_at DESC
		LIMIT 15; 
	`
	err := r.connection.Select(&messages, query, chatUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages from db: %v", err)
	}

	return &messages, nil
}
