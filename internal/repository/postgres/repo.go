package postgres

import (
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
)

//const (
//	//TODO: убрать после добавления kafka-consumer-avatar
//	defaultAvatar = "https://storage.yandexcloud.net/space21/avatars/default/logo-discord.jpeg"
//	typePrivate   = "private"
//)

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

func (r *Repository) CreatePrivateChat(initiator *model.ChatMemberParams, companion *model.ChatMemberParams) (string, error) {
	var chatUUID string

	//TODO: сделать через Squirrel
	sqlStr := "INSERT INTO chats DEFAULT VALUES RETURNING uuid"

	err := r.connection.QueryRow(sqlStr).Scan(&chatUUID)
	if err != nil {
		return "", fmt.Errorf("failed to create chat in db: %v", err)
	}

	query := sq.Insert("chats_user").
		Columns("chat_uuid", "user_uuid", "username", "avatar_link").
		Values(chatUUID, initiator.UserID, initiator.Nickname, initiator.AvatarLink).
		Values(chatUUID, companion.UserID, companion.Nickname, companion.AvatarLink).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build chat_members insert query: %v", err)
	}

	_, err = r.connection.Exec(sqlStr, args...)
	if err != nil {
		return "", fmt.Errorf("failed to insert chat members in db: %v", err)
	}

	return chatUUID, nil
}

func (r *Repository) GetPrivateChats(userUUID string) (*model.ChatInfoList, error) {
	var chats model.ChatInfoList

	query := sq.Select(
		"COALESCE(m.content, '') AS content",
		"(SELECT username FROM chats_user WHERE chat_uuid = c.uuid AND user_uuid = $1) AS chat_name",
		"(SELECT avatar_link FROM chats_user WHERE chat_uuid = c.uuid AND user_uuid = $1) AS avatar_link",
		"COALESCE((SELECT MAX(sent_at) FROM messages WHERE chat_uuid = c.uuid), c.created_at) AS created_at",
		"c.uuid",
	).
		From("chats_user cu").
		Join("chats c ON c.uuid = cu.chat_uuid").
		LeftJoin("messages m ON c.uuid = m.chat_uuid AND m.sent_at = (SELECT MAX(sent_at) FROM messages WHERE chat_uuid = c.uuid)").
		Where(sq.Eq{"cu.user_uuid": userUUID}).
		PlaceholderFormat(sq.Dollar)

	// Генерируем финальный SQL и список аргументов
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetPrivateChats query: %v", err)
	}

	err = r.connection.Select(&chats, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get private chats from db: %v", err)
	}

	return &chats, nil
}

func (r *Repository) GetGroupChats(userUUID string) (*model.ChatInfoList, error) {
	var chats model.ChatInfoList

	query := sq.Select(
		"COALESCE(gm.content, '') AS content",
		"gc.chat_name",
		"gc.avatar_link",
		"COALESCE((SELECT MAX(sent_at) FROM messages WHERE chat_uuid = gc.uuid), gc.created_at) AS created_at",
		"gc.uuid",
	).
		From("group_chats_user gcu").
		Join("group_chats gc ON gc.uuid = gcu.chat_uuid").
		LeftJoin("group_messages gm ON gc.uuid = gm.chat_uuid AND gm.sent_at = (SELECT MAX(sent_at) FROM group_messages WHERE chat_uuid = gc.uuid)").
		Where(sq.Eq{"gcu.user_uuid": userUUID}).
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

func (r *Repository) GetPrivateRecentMessages(chatUUID string, userUUID string) (*model.MessageList, error) {
	var messages model.MessageList

	query := sq.Select(
		"sender_uuid",
		"content",
		"sent_at",
		"COALESCE(updated_at, sent_at) AS updated_at",
		"root_uuid",
		"parent_uuid",
	).
		From("messages").
		Where(sq.Eq{"chat_uuid": chatUUID}).
		Where(sq.Or{
			sq.Eq{"delete_format": nil},
			sq.NotEq{"delete_format": "all"},
			sq.And{
				sq.Eq{"delete_format": "self"},
				sq.NotEq{"deleted_by": userUUID},
			},
		}).
		OrderBy("sent_at DESC").
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
