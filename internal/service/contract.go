package service

import "chat-service/internal/model"

type DBRepo interface {
	GetChat(chatUUID string) (*[]model.Message, error)
}
