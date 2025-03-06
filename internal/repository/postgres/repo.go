package postgres

import (
	"database/sql"
	"errors"
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

func (r *Repository) IsChatMember(chatUUID, userUUID string) (bool, error) {
	query := sq.
		Select("COUNT(*) > 0").
		From("chats_user").
		Where(sq.And{
			sq.Eq{"chat_uuid": chatUUID},
			sq.Eq{"user_uuid": userUUID},
		}).
		PlaceholderFormat(sq.Dollar)

	var isMember bool
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build IsChatMember query: %v", err)
	}

	err = r.connection.Get(&isMember, sqlStr, args...)
	if err != nil {
		return false, fmt.Errorf("failed to check user membership in db: %v", err)
	}

	return isMember, nil
}

func (r *Repository) GetPrivateDeletionInfo(messageID string) (*model.DeletionInfo, error) {
	var deletionInfo model.DeletionInfo

	query := sq.Select(
		"COALESCE(delete_format::text, '') AS delete_format",
		"COALESCE(deleted_by::text, '') AS deleted_by",
		"COALESCE(to_char(deleted_at, 'YYYY-MM-DD\"T\"HH24:MI:SSZ'), '') AS deleted_at").
		From("messages").
		Where(sq.Eq{"uuid": messageID}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetPrivateDeletionInfo query: %v", err)
	}

	err = r.connection.Get(&deletionInfo, sqlStr, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get deletion info from db: %v", err)
	}

	return &deletionInfo, nil
}

func (r *Repository) DeletePrivateMessage(userUUID, messageID, mode string) (bool, error) {
	query := sq.Update("messages").
		Set("deleted_by", userUUID).
		Set("delete_format", mode).
		Set("deleted_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"uuid": messageID}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build DeletePrivateMessage query: %v", err)
	}

	_, err = r.connection.Exec(sqlStr, args...)
	if err != nil {
		return false, fmt.Errorf("failed to delete message in db: %v", err)
	}

	return true, nil
}
