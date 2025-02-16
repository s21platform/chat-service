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

func (r *Repository) GetChats(UUID string) (*model.ChatInfoList, error) {
	var chats model.ChatInfoList

	query := `
		SELECT
			COALESCE(m.content, '') AS content,
			c.chat_name,
			c.avatar_link,
			m.created_at,
			c.uuid
		FROM chat_members cm
			JOIN public.chats c on c.id = cm.chat_id
			LEFT JOIN public.messages m ON c.last_message_id = m.id
		WHERE cm.user_uuid = $1
		ORDER BY m.created_at DESC`
	err := r.connection.Select(&chats, query, UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chats from db: %v", err)
	}

	return &chats, nil
}

func (r *Repository) GetRecentMessages(chatUUID string) (*[]model.Message, error) {
	var messages []model.Message

	query := `
		SELECT sender_uuid, content, created_at FROM messages
		WHERE chat_uuid = $1
		ORDER BY created_at DESC
		LIMIT 15`
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
		RETURNING id, content`
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
	WHERE id = $2`
	_, err := r.connection.Exec(query, mode, messageID)
	if err != nil {
		return false, fmt.Errorf("failed to delete message in db: %v", err)
	}

	return true, nil
}
