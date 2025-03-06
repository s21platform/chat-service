//go:generate mockgen -destination=mock_contract_test.go -package=${GOPACKAGE} -source=contract.go
package service

import (
	"context"

	"github.com/s21platform/chat-service/internal/model"
)

type DBRepo interface {
	CreatePrivateChat(initiator *model.ChatMemberParams, companion *model.ChatMemberParams) (string, error)
	GetPrivateChats(userUUID string) (*model.ChatInfoList, error)
	GetGroupChats(userUUID string) (*model.ChatInfoList, error)
	GetPrivateRecentMessages(chatUUID string, userUUID string) (*model.MessageList, error)
	DeleteMessage(messageID string, mode string) (bool, error)
	EditMessage(messageID string, newContent string) (*model.EditedMessage, error)
}

type UserClient interface {
	GetUserInfoByUUID(ctx context.Context, userUUID string) (*model.UserInfo, error)
}
