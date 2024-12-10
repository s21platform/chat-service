package rpc

import "chat-service/internal/model"

type DbRepo interface {
	GetChat(chatUUID string) (*[]model.Message, error)
}
