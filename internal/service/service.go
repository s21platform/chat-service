package service

import (
	"context"
	"fmt"

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
		UserID:     initiatorID,
		Nickname:   initiatorSetup.UserName,
		AvatarLink: initiatorSetup.AvatarLink,
	}

	companionParams := &model.ChatMemberParams{
		UserID:     in.CompanionUuid,
		Nickname:   companionSetup.UserName,
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

//func (s *Server) EditMessage(ctx context.Context, in *chat.EditMessageIn) (*chat.EditMessageOut, error) {
//	logger := logger_lib.FromContext(ctx, config.KeyLogger)
//	logger.AddFuncName("EditMessage")
//
//	data, err := s.repository.EditMessage(in.UuidMessage, in.NewContent)
//	if err != nil {
//		logger.Error(fmt.Sprintf("failed to edit message: %v", err))
//		return nil, fmt.Errorf("failed to edit message: %v", err)
//	}
//
//	return &chat.EditMessageOut{
//		UuidMessage: data.MessageID.String(),
//		NewContent:  data.Content,
//	}, nil
//}

func (s *Server) DeletePrivateMessage(ctx context.Context, in *chat.DeletePrivateMessageIn) (*chat.DeletePrivateMessageOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("DeletePrivateMessage")

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to find uuid")
		return nil, fmt.Errorf("failed to find uuid")
	}

	deletionInfo, err := s.repository.GetPrivateDeletionInfo(in.UuidMessage)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get private message deletion info: %v", err))
		return nil, fmt.Errorf("failed to get private message deletion info: %v", err)
	}

	if in.Mode != model.Self && in.Mode != model.All {
		logger.Error(fmt.Sprintf("invalid mode: %s", in.Mode))
		return nil, fmt.Errorf("invalid mode: %s", in.Mode)
	}

	if (deletionInfo.DeleteFormat == model.All) || (deletionInfo.DeleteFormat == model.Self && deletionInfo.DeletedBy == userUUID) {
		logger.Error("failed to message is already deleted")
		return nil, fmt.Errorf("failed to message is already deleted")
	}

	if deletionInfo.DeleteFormat == model.Self && deletionInfo.DeletedBy != userUUID {
		in.Mode = model.All
	}

	isDeleted, err := s.repository.DeletePrivateMessage(userUUID, in.UuidMessage, in.Mode)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to delete private message: %v", err))
		return nil, fmt.Errorf("failed to delete private message: %v", err)
	}

	return &chat.DeletePrivateMessageOut{
		DeletionStatus: isDeleted,
	}, nil
}
