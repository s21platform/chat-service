package service

import (
	"context"
	"fmt"
	"time"

	chat "github.com/s21platform/chat-proto/chat-proto"
	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
)

type Server struct {
	chat.UnimplementedChatServiceServer
	repository DBRepo
	userClient UserClient
}

func New(repo DBRepo, userClient UserClient) *Server {
	return &Server{
		repository: repo,
		userClient: userClient,
	}
}

func (s *Server) CreatePrivateChat(ctx context.Context, in *chat.CreatePrivateChatIn) (*chat.CreatePrivateChatOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("CreatePrivateChat")

	initiatorID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get initiatorID")
		return nil, fmt.Errorf("failed to get initiatorID")
	}

	initiatorSetup, err := s.userClient.GetUserInfoByUUID(ctx, initiatorID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get initiator info: %v", err))
		return nil, fmt.Errorf("failed to get initiator info: %v", err)
	}

	companionSetup, err := s.userClient.GetUserInfoByUUID(ctx, in.CompanionUuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get companion info: %v", err))
		return nil, fmt.Errorf("failed to get companion info: %v", err)
	}

	initiatorParams := &model.ChatMemberParams{
		UserUUID:   initiatorID,
		Nickname:   initiatorSetup.Nickname,
		AvatarLink: initiatorSetup.AvatarLink,
	}

	companionParams := &model.ChatMemberParams{
		UserUUID:   in.CompanionUuid,
		Nickname:   companionSetup.Nickname,
		AvatarLink: companionSetup.AvatarLink,
	}

	chatUUID, err := s.repository.CreatePrivateChat(initiatorParams, companionParams)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create chat: %v", err))
		return nil, fmt.Errorf("failed to create chat: %v", err)
	}

	return &chat.CreatePrivateChatOut{
		NewChatUuid: chatUUID,
	}, nil
}

func (s *Server) GetChats(ctx context.Context, _ *chat.ChatEmpty) (*chat.GetChatsOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetChats")

	uuid, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		return nil, fmt.Errorf("failed to find uuid")
	}

	chats, err := s.repository.GetChats(uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get chats: %v", err))
		return nil, fmt.Errorf("failed to get chats: %v", err)
	}

	return &chat.GetChatsOut{
		Chats: chats.FromDTO(),
	}, nil
}

func (s *Server) GetPrivateRecentMessages(ctx context.Context, in *chat.GetPrivateRecentMessagesIn) (*chat.GetPrivateRecentMessagesOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetRecentMessages")

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to find uuid")
		return nil, fmt.Errorf("failed to find uuid")
	}

	messages, err := s.repository.GetPrivateRecentMessages(in.ChatUuid, userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to fetch chat: %v", err))
		return nil, fmt.Errorf("failed to fetch chat: %v", err)
	}

	return &chat.GetPrivateRecentMessagesOut{
		Messages: messages.FromDTO(),
	}, nil
}

func (s *Server) EditPrivateMessage(ctx context.Context, in *chat.EditPrivateMessageIn) (*chat.EditPrivateMessageOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("EditPrivateMessage")

	data, err := s.repository.EditPrivateMessage(in.UuidMessage, in.NewContent)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to edit private message: %v", err))
		return nil, fmt.Errorf("failed to edit private message: %v", err)
	}

	return &chat.EditPrivateMessageOut{
		UuidMessage: data.MessageUUID.String(),
		NewContent:  data.Content,
		UpdatedAt:   data.UpdateAt.Format(time.RFC3339),
	}, nil
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
