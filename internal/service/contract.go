//go:generate mockgen -destination=mock_contract_test.go -package=${GOPACKAGE} -source=contract.go
package service

import (
	"context"

	"github.com/s21platform/chat-service/internal/model"
)

type DBRepo interface {
	CreatePrivateChat(initiator, companion *model.ChatMemberParams) (string, error)
	GetChats(UUID string) (*model.ChatInfoList, error)
	GetPrivateRecentMessages(chatUUID, userUUID string) (*model.MessageList, error)

	EditMessage(messageID, newContent string) (*model.EditedMessage, error)
	GetPrivateDeletionInfo(messageID string) (*model.DeletionInfo, error)
	DeletePrivateMessage(userUUID, messageID, mode string) (bool, error)
}

type UserClient interface {
	GetUserInfoByUUID(ctx context.Context, userUUID string) (*model.UserInfo, error)
}
