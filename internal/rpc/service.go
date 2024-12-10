package rpc

import (
	chat "github.com/s21platform/chat-proto/chat-proto"
)

type Server struct {
	chat.UnimplementedChatServiceServer
}
