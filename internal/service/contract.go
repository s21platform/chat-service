package service

import (
	"github.com/s21platform/chat-service/internal/model"
)

type DBRepo interface {
	CreatePrivateChat(params *model.PrivateChatSetup) (string, error)
	GetChats(UUID string) (*model.ChatInfoList, error)
	GetRecentMessages(chatUUID string) (*model.MessageList, error)
	DeleteMessage(messageID string, mode string) (bool, error)
	EditMessage(messageID string, newContent string) (*model.EditedMessage, error)
}
