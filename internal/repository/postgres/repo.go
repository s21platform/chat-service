package postgres

import (
	"fmt"
	"log"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"

	"github.com/jmoiron/sqlx"
)

const (
	avatarLink = "https://storage.yandexcloud.net/space21/avatars/default/logo-discord.jpeg"
)

type Repository struct {
	connection *sqlx.DB
}

func New(cfg *config.Config) *Repository {
	conStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.Host, cfg.Postgres.Port)

	conn, err := sqlx.Connect("postgres", conStr)
	if err != nil {
		log.Fatal("error connect: ", err)
	}

	return &Repository{
		connection: conn,
	}
}

func (r *Repository) Close() {
	_ = r.connection.Close()
}

func (r *Repository) GetRecentMessages(chatUUID string) (*[]model.Message, error) {
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

func (r *Repository) EditMessage(messageID string, newContent string) (*model.EditedMessage, error) {
	var editedMessage model.EditedMessage

	query := `
		UPDATE messages 
		SET content = $1, edited_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id, content;
    `
	err := r.connection.Get(&editedMessage, query, newContent, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to edit message in db: %v", err)
	}

	return &editedMessage, nil
}

func (r *Repository) DeleteMessage(messageID string, mode string) (bool, error) {
	query := `
	UPDATE messages
	SET deleted_for = $1
	WHERE id = $2;
`
	_, err := r.connection.Exec(query, mode, messageID)
	if err != nil {
		return false, fmt.Errorf("failed to delete message in db: %v", err)
	}

	return true, nil
}

func (r *Repository) CreateChat(initiatorID, companionID string) (string, error) {
	var chatID int
	var chatUUID string
	query := `
	INSERT INTO chats (uuid, type, created_at, avatar_link)
	VALUES (gen_random_uuid(), 'private', CURRENT_TIMESTAMP, $1)
	RETURNING id, uuid;
`
	err := r.connection.QueryRow(query, avatarLink).Scan(&chatID, &chatUUID)
	if err != nil {
		return "", fmt.Errorf("failed to create chat in db: %v", err)
	}

	query = `
INSERT INTO chat_members (chat_id, user_uuid)
VALUES ($1, $2), ($1, $3);
`
	_, err = r.connection.Exec(query, chatID, initiatorID, companionID)
	if err != nil {
		return "", fmt.Errorf("failed to create chat in db: %v", err)
	}
	return chatUUID, nil
}
