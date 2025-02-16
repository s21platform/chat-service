package service

import (
	"context"
	"fmt"
	"time"

	"github.com/s21platform/chat-service/internal/model"

	chat "github.com/s21platform/chat-proto/chat-proto"
	"github.com/s21platform/chat-service/internal/config"
	logger_lib "github.com/s21platform/logger-lib"
)

type Server struct {
	chat.UnimplementedChatServiceServer
	repository DBRepo
}

func New(repo DBRepo) *Server {
	return &Server{
		repository: repo,
	}
}

func (s *Server) GetChats(ctx context.Context, in *chat.GetChatsIn) (*chat.GetChatsOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetChats")

	chats, err := s.repository.GetChats(in.Uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get chats: %v", err))
		return nil, fmt.Errorf("failed to get chats: %v", err)
	}

	return &chat.GetChatsOut{
		Chats: chats.FromDTO(),
	}, nil
}

func (s *Server) GetRecentMessages(ctx context.Context, in *chat.GetRecentMessagesIn) (*chat.GetRecentMessagesOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetRecentMessages")

	data, err := s.repository.GetRecentMessages(in.Uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to fetch chat: %v", err))
		return nil, fmt.Errorf("failed to fetch chat: %v", err)
	}

	out := &chat.GetRecentMessagesOut{
		Messages: make([]*chat.Message, len(*data)),
	}

	for i, message := range *data {
		out.Messages[i] = &chat.Message{
			Uuid:    message.Uuid.String(),
			Content: message.Content,
			SentAt:  message.SentAt.Format(time.RFC3339),
		}
	}

	return out, nil
}

func (s *Server) EditMessage(ctx context.Context, in *chat.EditMessageIn) (*chat.EditMessageOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("EditMessage")

	data, err := s.repository.EditMessage(in.UuidMessage, in.NewContent)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to edit message: %v", err))
		return nil, fmt.Errorf("failed to edit message: %v", err)
	}

	out := &chat.EditMessageOut{
		UuidMessage: data.MessageID.String(),
		NewContent:  data.Content,
	}

	return out, nil
}

func (s *Server) DeleteMessage(ctx context.Context, in *chat.DeleteMessageIn) (*chat.DeleteMessageOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("DeleteMessage")

	if in.Mode != model.Self && in.Mode != model.All {
		return nil, fmt.Errorf("invalid mode: %s", in.Mode)
	}

	isDeleted, err := s.repository.DeleteMessage(in.UuidMessage, in.Mode)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to delete message: %v", err))
		return nil, fmt.Errorf("failed to delete message: %v", err)
	}

	return &chat.DeleteMessageOut{
		DeletionStatus: isDeleted,
	}, nil
}
