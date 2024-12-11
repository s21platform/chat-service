package rpc

import (
	"context"
	"fmt"
	"time"

	chat "github.com/s21platform/chat-proto/chat-proto"
)

type Server struct {
	chat.UnimplementedChatServiceServer
	dbR DbRepo
}

func New(repo DbRepo) *Server {
	return &Server{
		dbR: repo,
	}
}

func (s *Server) GetChat(ctx context.Context, in *chat.GetChatIn) (*chat.GetChatOut, error) {
	data, err := s.dbR.GetChat(in.Uuid)
	if err != nil {
		return nil, fmt.Errorf("s.dbR.GetChat: %v", err)
	}

	out := &chat.GetChatOut{
		Messages: make([]*chat.Message, len(*data)),
	}

	for i, message := range *data {
		out.Messages[i] = &chat.Message{
			Uuid: message.Uuid.String(),
			Content: message.Content,
			SentAt: message.SentAt.Format(time.RFC3339),
		}
	}

	return out, nil
}
