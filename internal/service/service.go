package service

import (
	"context"
	"fmt"
	"github.com/s21platform/chat-service/internal/model"
	"google.golang.org/grpc/metadata"
	"time"

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

func (s *Server) CreateChat(ctx context.Context, in *chat.CreateChatIn) (*chat.CreateChatOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("CreateChat")

	metadataCtx, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Error(fmt.Sprintf("user_uuid is missing in metadata"))
		return nil, fmt.Errorf("user_uuid is missing in metadata")
	}

	userUUIDs, exists := metadataCtx["user_uuid"]
	if !exists || len(userUUIDs) == 0 || userUUIDs[0] == "" {
		logger.Error("user_uuid is missing in metadata")
		return nil, fmt.Errorf("unauthorized: user_uuid is missing")
	}

	initiatorID := userUUIDs[0]
	companionID := in.CompanionUuid
	if len(companionID) == 0 || companionID == "" {
		logger.Error(fmt.Sprintf("companion_uuid is empty"))
		return nil, fmt.Errorf("companion_uuid is empty")
	}

	newChatUUID, err := s.repository.CreateChat(initiatorID, companionID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create chat: %v", err))
		return nil, fmt.Errorf("failed to create chat: %v", err)
	}

	out := &chat.CreateChatOut{
		NewChatUuid: newChatUUID,
	}
	return out, nil
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
