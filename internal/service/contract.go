//go:generate mockgen -destination=mock_contract_test.go -package=${GOPACKAGE} -source=contract.go
package service

import (
	"context"

	"github.com/s21platform/chat-service/internal/model"
)

type DBRepo interface {
	CreatePrivateChat() (string, error)
	AddPrivateChatMember(chatUUID string, member *model.ChatMemberParams) error
	GetPrivateChats(userUUID string) (*model.ChatInfoList, error)
	GetGroupChats(userUUID string) (*model.ChatInfoList, error)
	GetPrivateRecentMessages(chatUUID string, userUUID string) (*model.MessageList, error)
	DeleteMessage(messageID string, mode string) (bool, error)
	GetPrivateDeletionInfo(messageID string) (*model.DeletionInfo, error)
	EditPrivateMessage(messageUUID string, newContent string) (*model.EditedMessage, error)
	IsChatMember(chatUUID, userUUID string) (bool, error)
	GetPrivateMessage(messageUUID string) (*model.EditedMessage, error)
}

type UserClient interface {
	GetUserInfoByUUID(ctx context.Context, userUUID string) (*model.ChatMemberParams, error)
}
