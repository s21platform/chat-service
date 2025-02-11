package service

import (
	"github.com/google/uuid"
	"github.com/s21platform/chat-service/internal/model"
)

type DBRepo interface {
	GetRecentMessages(chatUUID string) (*[]model.Message, error)
	DeleteMessage(messageUUID uuid.UUID, scope model.DeletionScope) (bool, error)
	EditMessage(messageID string, newContent string) (*model.EditedMessage, error)
}
