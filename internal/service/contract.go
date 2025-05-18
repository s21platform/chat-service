//go:generate mockgen -destination=mock_contract_test.go -package=${GOPACKAGE} -source=contract.go
package service

import (
	"context"

	"github.com/s21platform/chat-service/internal/model"
)

type DBRepo interface {
	CreatePrivateChat(ctx context.Context) (string, error)
	AddPrivateChatMember(ctx context.Context, chatUUID string, member *model.ChatMemberParams) error
	GetPrivateChats(ctx context.Context, userUUID string) (*model.ChatInfoList, error)
	GetGroupChats(ctx context.Context, userUUID string) (*model.ChatInfoList, error)
	GetPrivateRecentMessages(ctx context.Context, chatUUID string, userUUID string) (*model.MessageList, error)
	DeletePrivateMessage(ctx context.Context, userUUID, messageID, mode string) (bool, error)
	GetPrivateDeletionInfo(ctx context.Context, messageID string) (*model.DeletionInfo, error)
	EditPrivateMessage(ctx context.Context, messageUUID string, newContent string) (*model.EditedMessage, error)
	IsChatMember(ctx context.Context, chatUUID, userUUID string) (bool, error)
	IsMessageOwner(ctx context.Context, chatUUID, messageUUID, userUUID string) (bool, error)
}

type UserClient interface {
	GetUserInfoByUUID(ctx context.Context, userUUID string) (*model.ChatMemberParams, error)
}
