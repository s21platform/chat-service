package service

import (
	"context"
	"fmt"
	"time"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
	"github.com/s21platform/chat-service/pkg/chat"
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

	chatUUID, err := s.repository.CreatePrivateChat()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create chat: %v", err))
		return nil, fmt.Errorf("failed to create chat: %v", err)
	}

	err = s.repository.AddPrivateChatMember(chatUUID, initiatorParams)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to add initiator to private chat: %v", err))
		return nil, fmt.Errorf("failed to add initiator to private chat: %v", err)
	}

	err = s.repository.AddPrivateChatMember(chatUUID, companionParams)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to add companion to private chat: %v", err))
		return nil, fmt.Errorf("failed to add companion to private chat: %v", err)
	}

	return &chat.CreatePrivateChatOut{
		NewChatUuid: chatUUID,
	}, nil
}

func (s *Server) GetChats(ctx context.Context, _ *chat.ChatEmpty) (*chat.GetChatsOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetChats")

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to find userUUID")
		return nil, fmt.Errorf("failed to find userUUID")
	}

	privateChats, err := s.repository.GetPrivateChats(userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get private chats: %v", err))
		return nil, fmt.Errorf("failed to get private chats: %v", err)
	}

	groupChats, err := s.repository.GetGroupChats(userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get group chats: %v", err))
		return nil, fmt.Errorf("failed to get group chats: %v", err)
	}

	allChats := append(*privateChats, *groupChats...)

	return &chat.GetChatsOut{
		Chats: allChats.FromDTO(),
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

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to find uuid")
		return nil, fmt.Errorf("failed to find uuid")
	}

	isMember, err := s.repository.IsChatMember(in.ChatUuid, userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to check user in chat: %v", err))
		return nil, fmt.Errorf("failed to check user in chat: %v", err)
	}

	if !isMember {
		logger.Error("failed to user is not chat member")
		return nil, fmt.Errorf("failed to user is not chat member")
	}

	isDeleted, err := s.repository.GetPrivateDeletionInfo(in.MessageUuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to check deletion status: %v", err))
		return nil, fmt.Errorf("failed to check deletion status: %v", err)
	}

	if isDeleted != nil && isDeleted.DeletedAt != "" {
		logger.Error("failed to edit deleted message")
		return nil, fmt.Errorf("attempt to edit deleted message")
	}

	isUserMessageOwner, err := s.repository.IsMessageOwner(in.ChatUuid, in.MessageUuid, userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to check message owner: %v", err))
		return nil, fmt.Errorf("failed to check message owner: %v", err)
	}

	if !isUserMessageOwner {
		logger.Error("failed to user is not message owner")
		return nil, fmt.Errorf("failed to user is not message owner")
	}

	data, err := s.repository.EditPrivateMessage(in.MessageUuid, in.NewContent)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to edit private message: %v", err))
		return nil, fmt.Errorf("failed to edit private message: %v", err)
	}

	return &chat.EditPrivateMessageOut{
		MessageUuid: data.MessageUUID.String(),
		NewContent:  data.Content,
		UpdatedAt:   data.UpdateAt.Format(time.RFC3339),
	}, nil
}

func (s *Server) DeletePrivateMessage(ctx context.Context, in *chat.DeletePrivateMessageIn) (*chat.DeletePrivateMessageOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("DeletePrivateMessage")

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to find uuid")
		return nil, fmt.Errorf("failed to find uuid")
	}

	userIsChatMember, err := s.repository.IsChatMember(in.ChatUuid, userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to check if user is chat member: %v", err))
		return nil, fmt.Errorf("failed to check if user is chat member: %v", err)
	}

	if !userIsChatMember {
		logger.Error("failed to user is not chat member")
		return nil, fmt.Errorf("failed to user is not chat member")
	}

	if in.Mode != model.Self && in.Mode != model.All {
		logger.Error(fmt.Sprintf("failed to invalid mode: %s", in.Mode))
		return nil, fmt.Errorf("failed to invalid mode: %s", in.Mode)
	}

	deletionInfo, err := s.repository.GetPrivateDeletionInfo(in.MessageUuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get private message deletion info: %v", err))
		return nil, fmt.Errorf("failed to get private message deletion info: %v", err)
	}

	if (deletionInfo.DeleteFormat == model.All) || (deletionInfo.DeleteFormat == model.Self && deletionInfo.DeletedBy == userUUID) {
		logger.Error("failed to message is already deleted")
		return nil, fmt.Errorf("failed to message is already deleted")
	}

	if deletionInfo.DeleteFormat == model.Self && deletionInfo.DeletedBy != userUUID {
		in.Mode = model.All
	}

	isDeleted, err := s.repository.DeletePrivateMessage(userUUID, in.MessageUuid, in.Mode)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to delete private message: %v", err))
		return nil, fmt.Errorf("failed to delete private message: %v", err)
	}

	return &chat.DeletePrivateMessageOut{
		DeletionStatus: isDeleted,
	}, nil
}
