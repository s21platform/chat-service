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

func (r *Repository) CreatePrivateChat() (string, error) {
	var chatUUID string

	query, args, err := sq.Insert("chats").
		Columns().
		Values().
		Suffix("RETURNING uuid").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return "", fmt.Errorf("failed to build insert query: %v", err)
	}

	err = r.connection.Get(&chatUUID, query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to create chat in db: %v", err)
	}

	return chatUUID, nil
}

func (r *Repository) AddPrivateChatMember(chatUUID string, member *model.ChatMemberParams) error {
	query, args, err := sq.Insert("chats_user").
		Columns("chat_uuid", "user_uuid", "username", "avatar_link").
		Values(chatUUID, member.UserUUID, member.Nickname, member.AvatarLink).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build chat_members insert query: %v", err)
	}

	_, err = r.connection.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert chat members in db: %v", err)
	}

	return nil
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

	query, args, err := sq.Select(
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
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build GetRecentMessages query: %v", err)
	}

	err = r.connection.Select(&messages, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages from db: %v", err)
	}

	return &messages, nil
}

func (r *Repository) GetPrivateDeletionInfo(messageID string) (*model.DeletionInfo, error) {
	var deletionInfo model.DeletionInfo

	query, args, err := sq.Select(
		"COALESCE(delete_format::text, '') AS delete_format",
		"COALESCE(deleted_by::text, '') AS deleted_by",
		"COALESCE(to_char(deleted_at, 'YYYY-MM-DD\"T\"HH24:MI:SSZ'), '') AS deleted_at").
		From("messages").
		Where(sq.Eq{"uuid": messageID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build GetPrivateDeletionInfo query: %v", err)
	}

	err = r.connection.Get(&deletionInfo, query, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get deletion info from db: %v", err)
	}

	return &deletionInfo, nil
}

func (r *Repository) IsChatMember(chatUUID, userUUID string) (bool, error) {
	var isMember bool

	query, args, err := sq.
		Select("COUNT(*) > 0").
		From("chats_user").
		Where(sq.And{
			sq.Eq{"chat_uuid": chatUUID},
			sq.Eq{"user_uuid": userUUID},
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("failed to build IsChatMember query: %v", err)
	}

	err = r.connection.Get(&isMember, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to check user membership in db: %v", err)
	}

	return isMember, nil
}

func (r *Repository) EditPrivateMessage(messageUUID string, newContent string) (*model.EditedMessage, error) {
	var editedPrivateMessage model.EditedMessage

	query, args, err := sq.Update("messages").
		Set("content", newContent).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"uuid": messageUUID}).
		Suffix("RETURNING uuid, content, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build EditPrivateMessage query: %v", err)
	}

	err = r.connection.Get(&editedPrivateMessage, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to edit private message in db: %v", err)
	}

	return &editedPrivateMessage, nil
}

func (r *Repository) GetPrivateMessage(messageUUID string) (*model.EditedMessage, error) {
	var editedMessage model.EditedMessage

	query, args, err := sq.Select(
		"uuid",
		"content",
		"updated_at").
		From("messages").
		Where(sq.Eq{"uuid": messageUUID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build GetPrivateMessage query: %v", err)
	}

	err = r.connection.Get(&editedMessage, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get message from db: %v", err)
	}

	return &editedMessage, nil
}

func (r *Repository) DeleteMessage(messageID string, mode string) (bool, error) {
	query, args, err := sq.Update("messages").
		Set("deleted_for", mode).
		Where(sq.Eq{"id": messageID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("failed to build DeleteMessage query: %v", err)
	}

	_, err = r.connection.Exec(query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to delete message in db: %v", err)
	}

	return true, nil
}
