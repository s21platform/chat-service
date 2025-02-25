package service

import (
	"context"
	"fmt"
	"github.com/s21platform/chat-service/internal/rpc/user"

	chat "github.com/s21platform/chat-proto/chat-proto"
	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
)

type Server struct {
	chat.UnimplementedChatServiceServer
	repository DBRepo
	userClient *user.Service
}

func New(repo DBRepo, userClient *user.Service) *Server {
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
		return nil, fmt.Errorf("failed to find initiatorID")
	}

	companionInfo, err := s.userClient.GetUserInfoByUUID(ctx, in.CompanionUuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get companion info: %v", err))
		return nil, fmt.Errorf("failed to get companion info: %v", err)
	}

	chatParams := &model.CreatePrivatChatParams{
		InitiatorID:     initiatorID,
		CompanionID:     in.CompanionUuid,
		CompanionName:   companionInfo.UserName,
		CompanionAvatar: companionInfo.AvatarLink,
	}

	chatUUID, err := s.repository.CreatePrivateChat(chatParams)
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

func (s *Server) GetRecentMessages(ctx context.Context, _ *chat.ChatEmpty) (*chat.GetRecentMessagesOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetRecentMessages")

	uuid, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		return nil, fmt.Errorf("failed to find uuid")
	}

	messageList, err := s.repository.GetRecentMessages(uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to fetch chat: %v", err))
		return nil, fmt.Errorf("failed to fetch chat: %v", err)
	}

	return &chat.GetRecentMessagesOut{
		Messages: messageList.FromDTO(),
	}, nil
}

func (s *Server) EditMessage(ctx context.Context, in *chat.EditMessageIn) (*chat.EditMessageOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("EditMessage")

	data, err := s.repository.EditMessage(in.UuidMessage, in.NewContent)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to edit message: %v", err))
		return nil, fmt.Errorf("failed to edit message: %v", err)
	}

	return &chat.EditMessageOut{
		UuidMessage: data.MessageID.String(),
		NewContent:  data.Content,
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
