package service

import "github.com/s21platform/chat-service/internal/model"

type DBRepo interface {
	GetRecentMessages(chatUUID string) (*[]model.Message, error)
	EditMessage(MessageUUID string, Content string) (*model.EditedMessage, error)
}
