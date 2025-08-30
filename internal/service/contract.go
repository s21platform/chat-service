//go:generate mockgen -destination=mock_contract_test.go -package=${GOPACKAGE} -source=contract.go
package service

import (
	"context"

	"github.com/s21platform/chat-service/internal/model"
	"github.com/s21platform/chat-service/pkg/chat"
)

type DBRepo interface {
	CreateStream(ctx context.Context, streamType, metadata, createdBy string) (string, error)
	AddStreamMember(ctx context.Context, streamID, userID, metadata string) error
	AddStreamMembers(ctx context.Context, streamID string, members []model.StreamMember) error
	AddNewUser(ctx context.Context, userInfo *model.StreamMemberParams) error
	AddUserSubscription(ctx context.Context, userID, channel string) error
	AddUserSubscriptions(ctx context.Context, subscriptions []model.UserSubscription) error
	SaveMessage(ctx context.Context, message *model.Message) error
	IsStreamMember(ctx context.Context, streamID, userID string) (bool, error)
	GetPrivateStreams(ctx context.Context, requesterID string) (*model.PrivateStreamPreviewList, error)
	GetStreamRecentMessages(ctx context.Context, streamID string, offset string, limit int32) (*model.MessageList, error)
	GetUserActiveStreams(ctx context.Context, userID string) ([]string, error)

	WithTx(ctx context.Context, cb func(ctx context.Context) error) error
}

type UserClient interface {
	GetUserInfoByUUID(ctx context.Context, userUUID string) (*model.StreamMemberParams, error)
}

type CetrifugeClient interface {
	Publish(ctx context.Context, channel string, data model.Message) error
}

type Validator interface {
	ValidateStreamByType(in *chat.CreateStreamIn, creatorID string) error
	ValidateSendMessage(in *chat.SendMessageIn) error
}

type JWTGenerator interface {
	GenerateConnectToken(userID string) (string, int64, error)
	GenerateSubscribeToken(userID, streamID string) (string, int64, error)
	ValidateConnectToken(tokenString string) (*model.CentrifugoConnectClaims, error)
	ValidateSubscribeToken(tokenString string) (*model.CentrifugoSubscribeClaims, error)
}
