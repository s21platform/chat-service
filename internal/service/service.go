package service

import (
	"github.com/s21platform/chat-service/pkg/chat"
)

type Server struct {
	chat.UnimplementedChatServiceServer
}

func New() *Server {
	return &Server{}
}
