package postgres

import (
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
)

const (
	//TODO: убрать после добавления kafka-consumer-avatar
	defaultAvatar = "https://storage.yandexcloud.net/space21/avatars/default/logo-discord.jpeg"
	typePrivate   = "private"
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

func (r *Repository) CreateChat(initiatorID, companionID string) (string, error) {
	var chatID int
	var chatUUID string

	query := sq.Insert("chats").
		Columns("uuid", "type", "avatar_link").
		Values(sq.Expr("gen_random_uuid()"), typePrivate, defaultAvatar).
		Suffix("RETURNING id, uuid").
		PlaceholderFormat(sq.Dollar) // Используем $1, $2...

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build chat insert query: %v", err)
	}

	err = r.connection.QueryRow(sqlStr, args...).Scan(&chatID, &chatUUID)
	if err != nil {
		return "", fmt.Errorf("failed to create chat in db: %v", err)
	}

	query = sq.Insert("chat_members").
		Columns("chat_id", "user_uuid").
		Values(chatID, initiatorID).
		Values(chatID, companionID).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err = query.ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build chat_members insert query: %v", err)
	}

	_, err = r.connection.Exec(sqlStr, args...)
	if err != nil {
		return "", fmt.Errorf("failed to insert chat members in db: %v", err)
	}

	return chatUUID, nil
}

func (r *Repository) GetChats(UUID string) (*model.ChatInfoList, error) {
	var chats model.ChatInfoList

	query := sq.Select(
		"COALESCE(m.content, '') AS content",
		"c.chat_name",
		"c.avatar_link",
		"m.created_at",
		"c.uuid",
	).
		From("chat_members cm").
		Join("public.chats c ON c.id = cm.chat_id").
		LeftJoin("public.messages m ON c.last_message_id = m.id").
		Where(sq.Eq{"cm.user_uuid": UUID}).
		OrderBy("m.created_at DESC").
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetChats query: %v", err)
	}

	err = r.connection.Select(&chats, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get chats from db: %v", err)
	}

	return &chats, nil
}

func (r *Repository) GetRecentMessages(chatUUID string) (*[]model.Message, error) {
	var messages []model.Message

	query := sq.Select(
		"sender_uuid",
		"content",
		"created_at",
	).
		From("messages").
		Where(sq.Eq{"chat_uuid": chatUUID}).
		OrderBy("created_at DESC").
		Limit(15).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetRecentMessages query: %v", err)
	}

	err = r.connection.Select(&messages, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages from db: %v", err)
	}

	return &messages, nil
}

func (r *Repository) EditMessage(messageID string, newContent string) (*model.EditedMessage, error) {
	var editedMessage model.EditedMessage

	query := sq.Update("messages").
		Set("content", newContent).
		Set("edited_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": messageID}).
		Suffix("RETURNING id, content").
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build EditMessage query: %v", err)
	}

	err = r.connection.Get(&editedMessage, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to edit message in db: %v", err)
	}

	return &editedMessage, nil
}

func (r *Repository) DeleteMessage(messageID string, mode string) (bool, error) {
	query := sq.Update("messages").
		Set("deleted_for", mode).
		Where(sq.Eq{"id": messageID}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build DeleteMessage query: %v", err)
	}

	_, err = r.connection.Exec(sqlStr, args...)
	if err != nil {
		return false, fmt.Errorf("failed to delete message in db: %v", err)
	}

	return true, nil
}
