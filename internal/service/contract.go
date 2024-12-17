package service

import "github.com/s21platform/chat-service/internal/model"

type DBRepo interface {
	GetChat(chatUUID string) (*[]model.Message, error)
}
