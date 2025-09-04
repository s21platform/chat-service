package postgres

import (
	"context"
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

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

func (r *Repository) GetStreamRecentMessages(ctx context.Context, streamID string, offset string, limit int32) (*model.MessageList, error) {
	queryBuilder := sq.Select(
		"id",
		"stream_id",
		"sender_id",
		"type",
		"content",
		"root_id",
		"parent_id",
		"sent_at",
		"updated_at",
	).
		From("messages").
		Where(sq.Eq{"stream_id": streamID}).
		Where(sq.Eq{"deleted_at": nil}).
		OrderBy("sent_at DESC")

	if offset != "" {
		queryBuilder = queryBuilder.Where(sq.LtOrEq{"sent_at": offset})
	}

	if limit > 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit))
	} else {
		queryBuilder = queryBuilder.Limit(50) // дефолтный лимит
	}

	query, args, err := queryBuilder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %v", err)
	}

	var messages model.MessageList
	err = r.Chk(ctx).SelectContext(ctx, &messages, query, args...)
	if err != nil {
		return nil, err
	}

	return &messages, nil
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

	_, err = r.Chk(ctx).ExecContext(ctx, query, args...)
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

	_, err = r.Chk(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) CreateStream(ctx context.Context, streamType, metadata, createdBy string) (string, error) {
	query, args, err := sq.Insert("streams").
		Columns("type", "metadata", "created_by").
		Values(streamType, metadata, createdBy).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build sql query: %v", err)
	}

	var streamID string
	err = r.Chk(ctx).GetContext(ctx, &streamID, query, args...)
	if err != nil {
		return "", err
	}

	return streamID, nil
}

func (r *Repository) AddNewUser(ctx context.Context, userInfo *model.StreamMemberParams) error {
	query, args, err := sq.Insert("users").
		Columns("id", "nickname", "avatar_url").
		Values(userInfo.UserID, userInfo.Nickname, userInfo.AvatarURL).
		Suffix("ON CONFLICT (id) DO NOTHING").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %v", err)
	}

	_, err = r.Chk(ctx).ExecContext(ctx, query, args...)

	return err
}

func (r *Repository) AddStreamMembers(ctx context.Context, streamID string, members []model.StreamMember) error {
	if len(members) == 0 {
		return nil
	}

	query := sq.Insert("stream_members").
		Columns("stream_id", "user_id", "metadata").
		PlaceholderFormat(sq.Dollar)

	for _, member := range members {
		query = query.Values(streamID, member.UserID, member.Metadata)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %v", err)
	}

	_, err = r.Chk(ctx).ExecContext(ctx, sql, args...)
	return err
}

func (r *Repository) SaveMessage(ctx context.Context, message *model.Message) error {
	query := sq.Insert("messages").
		Columns("id", "stream_id", "sender_id", "type", "content", "root_id", "parent_id").
		Values(message.ID, message.StreamID, message.SenderID, message.Type, message.Content, message.RootID, message.ParentID).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %v", err)
	}

	_, err = r.Chk(ctx).ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to save message: %v", err)
	}

	return nil
}

func (r *Repository) IsStreamMember(ctx context.Context, streamID, userID string) (bool, error) {
	query, args, err := sq.
		Select("COUNT(*) > 0").
		From("stream_members").
		Where(sq.And{
			sq.Eq{"stream_id": streamID},
			sq.Eq{"user_id": userID},
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build sql query: %v", err)
	}

	var isMember bool
	err = r.Chk(ctx).GetContext(ctx, &isMember, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to check stream membership: %v", err)
	}

	return isMember, nil
}

func (r *Repository) AddUserSubscriptions(ctx context.Context, subscriptions []model.UserSubscription) error {
	query := sq.Insert("user_subscriptions").
		Columns("user_id", "channel").
		Suffix("ON CONFLICT (user_id, channel) DO NOTHING").
		PlaceholderFormat(sq.Dollar)

	for _, sub := range subscriptions {
		query = query.Values(sub.UserID, sub.Channel)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %v", err)
	}

	_, err = r.Chk(ctx).ExecContext(ctx, sql, args...)

	return err
}

func (r *Repository) GetPrivateStreams(ctx context.Context, requesterID string) (*model.PrivateStreamPreviewList, error) {
	query := sq.Select(
		"s.id as stream_id",
		"u_companion.nickname as stream_name",
		"u_companion.avatar_url",
		"("+func() string {
			sql, _, _ := sq.Select("content").
				From("messages m2").
				Where("m2.stream_id = s.id").
				Where(sq.Eq{"m2.deleted_at": nil}).
				OrderBy("m2.sent_at DESC").
				Limit(1).ToSql()
			return sql
		}()+") as last_message_content",
		"("+func() string {
			sql, _, _ := sq.Select("sent_at").
				From("messages m2").
				Where("m2.stream_id = s.id").
				Where(sq.Eq{"m2.deleted_at": nil}).
				OrderBy("m2.sent_at DESC").
				Limit(1).ToSql()
			return sql
		}()+") as last_message_timestamp",
	).
		From("streams s").
		Join("stream_members sm1 ON s.id = sm1.stream_id").
		Join("stream_members sm2 ON s.id = sm2.stream_id").
		Join("users u_companion ON sm2.user_id = u_companion.id").
		Where(sq.And{
			sq.Eq{"sm1.user_id": requesterID},
			sq.NotEq{"sm2.user_id": requesterID},
			sq.Eq{"sm1.left_at": nil},
			sq.Eq{"sm2.left_at": nil},
		}).
		OrderBy("s.created_at DESC").
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %v", err)
	}

	var streams model.PrivateStreamPreviewList
	err = r.Chk(ctx).SelectContext(ctx, &streams, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get streams: %v", err)
	}

	return &streams, nil
}

func (r *Repository) GetUserActiveStreams(ctx context.Context, userID string) ([]string, error) {
	queryBuilder := sq.Select("stream_id").
		From("stream_members").
		Where(sq.Eq{
			"user_id": userID,
			"left_at": nil,
		}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %v", err)
	}

	var streamIDs []string
	err = r.Chk(ctx).SelectContext(ctx, &streamIDs, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user active streams: %v", err)
	}

	return streamIDs, nil
}
