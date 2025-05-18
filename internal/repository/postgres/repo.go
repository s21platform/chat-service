package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

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

func (r *Repository) CreatePrivateChat(ctx context.Context) (string, error) {
	query, args, err := sq.Insert("chats").
		Columns("created_at").
		Values(time.Now()).
		Suffix("RETURNING uuid").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build sql query: %v", err)
	}

	var chatUUID string
	err = r.connection.GetContext(ctx, &chatUUID, query, args...)
	if err != nil {
		return "", err
	}

	return chatUUID, nil
}

func (r *Repository) AddPrivateChatMember(ctx context.Context, chatUUID string, member *model.ChatMemberParams) error {
	query, args, err := sq.Insert("chats_user").
		Columns("chat_uuid", "user_uuid", "username", "avatar_link").
		Values(chatUUID, member.UserUUID, member.Nickname, member.AvatarLink).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %v", err)
	}

	_, err = r.connection.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetPrivateChats(ctx context.Context, userUUID string) (*model.ChatInfoList, error) {
	query, args, err := sq.Select(
		"COALESCE(m.content, '') AS content",
		"(SELECT username FROM chats_user WHERE chat_uuid = c.uuid AND user_uuid != $1) AS chat_name",
		"(SELECT avatar_link FROM chats_user WHERE chat_uuid = c.uuid AND user_uuid != $1) AS avatar_link",
		"(SELECT MAX(sent_at) FROM messages WHERE chat_uuid = c.uuid) AS created_at",
		"c.uuid",
	).
		From("chats_user cu").
		Join("chats c ON c.uuid = cu.chat_uuid").
		LeftJoin("messages m ON c.uuid = m.chat_uuid AND m.sent_at = (SELECT MAX(sent_at) FROM messages WHERE chat_uuid = c.uuid)").
		Where(sq.Eq{"cu.user_uuid": userUUID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %v", err)
	}

	var chats model.ChatInfoList
	err = r.connection.SelectContext(ctx, &chats, query, args...)
	if err != nil {
		return nil, err
	}

	return &chats, nil
}

func (r *Repository) GetGroupChats(ctx context.Context, userUUID string) (*model.ChatInfoList, error) {
	query, args, err := sq.Select(
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
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %v", err)
	}

	var chats model.ChatInfoList
	err = r.connection.SelectContext(ctx, &chats, query, args...)
	if err != nil {
		return nil, err
	}

	return &chats, nil
}

func (r *Repository) GetPrivateRecentMessages(ctx context.Context, chatUUID string, userUUID string) (*model.MessageList, error) {
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
		return nil, fmt.Errorf("failed to build sql query: %v", err)
	}

	var messages model.MessageList
	err = r.connection.SelectContext(ctx, &messages, query, args...)
	if err != nil {
		return nil, err
	}

	return &messages, nil
}

func (r *Repository) GetPrivateDeletionInfo(ctx context.Context, messageID string) (*model.DeletionInfo, error) {
	query, args, err := sq.Select(
		"COALESCE(delete_format::text, '') AS delete_format",
		"COALESCE(deleted_by::text, '') AS deleted_by",
		"COALESCE(to_char(deleted_at, 'YYYY-MM-DD\"T\"HH24:MI:SSZ'), '') AS deleted_at").
		From("messages").
		Where(sq.Eq{"uuid": messageID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %v", err)
	}

	var deletionInfo model.DeletionInfo
	err = r.connection.GetContext(ctx, &deletionInfo, query, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return &deletionInfo, nil
}

func (r *Repository) EditPrivateMessage(ctx context.Context, messageUUID string, newContent string) (*model.EditedMessage, error) {
	query, args, err := sq.Update("messages").
		Set("content", newContent).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"uuid": messageUUID}).
		Suffix("RETURNING uuid, content, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %v", err)
	}

	var editedPrivateMessage model.EditedMessage
	err = r.connection.GetContext(ctx, &editedPrivateMessage, query, args...)
	if err != nil {
		return nil, err
	}

	return &editedPrivateMessage, nil
}

func (r *Repository) DeletePrivateMessage(ctx context.Context, userUUID, messageID, mode string) (bool, error) {
	query, args, err := sq.Update("messages").
		Set("deleted_by", userUUID).
		Set("delete_format", mode).
		Set("deleted_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"uuid": messageID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build sql query: %v", err)
	}

	_, err = r.connection.ExecContext(ctx, query, args...)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) IsChatMember(ctx context.Context, chatUUID, userUUID string) (bool, error) {
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
		return false, fmt.Errorf("failed to build sql query: %v", err)
	}

	var isMember bool
	err = r.connection.GetContext(ctx, &isMember, query, args...)
	if err != nil {
		return false, err
	}

	return isMember, nil
}

func (r *Repository) IsMessageOwner(ctx context.Context, chatUUID, messageUUID, userUUID string) (bool, error) {
	query, args, err := sq.
		Select("COUNT(*) > 0").
		From("messages").
		Where(sq.And{
			sq.Eq{"uuid": messageUUID},
			sq.Eq{"chat_uuid": chatUUID},
			sq.Eq{"sender_uuid": userUUID},
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build sql query: %v", err)
	}

	var isOwner bool
	err = r.connection.GetContext(ctx, &isOwner, query, args...)
	if err != nil {
		return false, err
	}

	return isOwner, nil
}

func (r *Repository) UpdateUserNickname(ctx context.Context, userUUID, newNickname string) error {
	query, args, err := sq.Update("chats_user").
		Set("username", newNickname).
		Where(sq.Eq{"user_uuid": userUUID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %v", err)
	}

	_, err = r.connection.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateUserAvatar(ctx context.Context, userUUID, avatarLink string) error {
	query, args, err := sq.Update("chats_user").
		Set("avatar_link", avatarLink).
		Where(sq.Eq{"user_uuid": userUUID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %v", err)
	}

	_, err = r.connection.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
