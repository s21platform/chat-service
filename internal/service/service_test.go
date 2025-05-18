package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
	"github.com/s21platform/chat-service/pkg/chat"
)

func TestServer_CreatePrivateChat(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	initiatorUUID := uuid.New().String()
	companionUUID := uuid.New().String()

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, initiatorUUID)

	s := New(mockRepo, mockUserClient)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreatePrivateChat")

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, initiatorUUID).
			Return(&model.ChatMemberParams{
				Nickname:   "test_initiator",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
			Return(&model.ChatMemberParams{
				Nickname:   "test_companion",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockRepo.EXPECT().CreatePrivateChat(ctx).
			Return("chat_uuid", nil)
		mockRepo.EXPECT().AddPrivateChatMember(ctx, gomock.Any(), gomock.Any()).
			Return(nil)
		mockRepo.EXPECT().AddPrivateChatMember(ctx, gomock.Any(), gomock.Any()).
			Return(nil)

		chatUUID, err := s.CreatePrivateChat(ctx, &chat.CreatePrivateChatIn{
			CompanionUuid: companionUUID,
		})

		assert.NoError(t, err)
		assert.NotNil(t, chatUUID)
		assert.Equal(t, "chat_uuid", chatUUID.NewChatUuid)
	})

	t.Run("no_initiatorUUID", func(t *testing.T) {
		badCtx := context.WithValue(context.Background(), config.KeyLogger, mockLogger)

		mockLogger.EXPECT().AddFuncName("CreatePrivateChat")
		mockLogger.EXPECT().Error("failed to get initiatorID")

		_, err := s.CreatePrivateChat(badCtx, &chat.CreatePrivateChatIn{
			CompanionUuid: companionUUID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get initiatorID")
	})

	t.Run("get_initiatorSetup_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreatePrivateChat")
		mockLogger.EXPECT().Error(gomock.Any())

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, initiatorUUID).
			Return(nil, fmt.Errorf("failed to get initiator info"))

		_, err := s.CreatePrivateChat(ctx, &chat.CreatePrivateChatIn{
			CompanionUuid: companionUUID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get initiator info")
	})

	t.Run("get_companionSetup_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreatePrivateChat")
		mockLogger.EXPECT().Error(gomock.Any())

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, initiatorUUID).
			Return(&model.ChatMemberParams{
				Nickname:   "test_initiator",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
			Return(nil, fmt.Errorf("failed to get companion info"))

		_, err := s.CreatePrivateChat(ctx, &chat.CreatePrivateChatIn{
			CompanionUuid: companionUUID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get companion info")
	})

	t.Run("DB_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreatePrivateChat")
		mockLogger.EXPECT().Error(gomock.Any())

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, initiatorUUID).
			Return(&model.ChatMemberParams{
				Nickname:   "test_initiator",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
			Return(&model.ChatMemberParams{
				Nickname:   "test_companion",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockRepo.EXPECT().CreatePrivateChat(ctx).
			Return("", fmt.Errorf("failed to create chat"))

		_, err := s.CreatePrivateChat(ctx, &chat.CreatePrivateChatIn{
			CompanionUuid: companionUUID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create chat")
	})

	t.Run("add_initiator_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreatePrivateChat")
		mockLogger.EXPECT().Error(gomock.Any())

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, initiatorUUID).
			Return(&model.ChatMemberParams{
				Nickname:   "test_initiator",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
			Return(&model.ChatMemberParams{
				Nickname:   "test_companion",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockRepo.EXPECT().CreatePrivateChat(ctx).
			Return("chat_uuid", nil)

		mockRepo.EXPECT().AddPrivateChatMember(ctx, "chat_uuid", gomock.Any()).
			Return(fmt.Errorf("failed to add initiator"))

		_, err := s.CreatePrivateChat(ctx, &chat.CreatePrivateChatIn{
			CompanionUuid: companionUUID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add initiator")
	})

	t.Run("add_companion_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreatePrivateChat")
		mockLogger.EXPECT().Error(gomock.Any())

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, initiatorUUID).
			Return(&model.ChatMemberParams{
				Nickname:   "test_initiator",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
			Return(&model.ChatMemberParams{
				Nickname:   "test_companion",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockRepo.EXPECT().CreatePrivateChat(ctx).
			Return("chat_uuid", nil)

		mockRepo.EXPECT().AddPrivateChatMember(ctx, "chat_uuid", gomock.Any()).
			Return(nil)

		mockRepo.EXPECT().AddPrivateChatMember(ctx, "chat_uuid", gomock.Any()).
			Return(fmt.Errorf("failed to add companion"))

		_, err := s.CreatePrivateChat(ctx, &chat.CreatePrivateChatIn{
			CompanionUuid: companionUUID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add companion")
	})
}

func TestServer_GetPrivateRecentMessages(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	chatUUID := uuid.New().String()

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, userUUID)

	s := New(mockRepo, mockUserClient)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetRecentMessages")

		expectedMessages := &model.MessageList{
			{
				Uuid:       uuid.New(),
				Content:    "message 1",
				SentAt:     time.Now().Add(-10 * time.Minute),
				UpdatedAt:  time.Now().Add(-5 * time.Minute),
				RootUUID:   uuid.Nil,
				ParentUUID: uuid.Nil,
			},
			{
				Uuid:       uuid.New(),
				Content:    "message 2",
				SentAt:     time.Now().Add(-10 * time.Minute),
				UpdatedAt:  time.Now().Add(-5 * time.Minute),
				RootUUID:   uuid.Nil,
				ParentUUID: uuid.Nil,
			},
		}

		mockRepo.EXPECT().GetPrivateRecentMessages(ctx, chatUUID, userUUID).Return(expectedMessages, nil)

		messages, err := s.GetPrivateRecentMessages(ctx, &chat.GetPrivateRecentMessagesIn{
			ChatUuid: chatUUID,
		})

		assert.NoError(t, err)
		assert.NotNil(t, messages)
		assert.Len(t, messages.Messages, 2)
	})

	t.Run("no_userUUID", func(t *testing.T) {
		badCtx := context.WithValue(context.Background(), config.KeyLogger, mockLogger)

		mockLogger.EXPECT().AddFuncName("GetRecentMessages")
		mockLogger.EXPECT().Error("failed to find uuid")

		_, err := s.GetPrivateRecentMessages(badCtx, &chat.GetPrivateRecentMessagesIn{
			ChatUuid: chatUUID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find uuid")
	})

	t.Run("DB_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetRecentMessages")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().GetPrivateRecentMessages(ctx, gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("failed to fetch chat"))

		_, err := s.GetPrivateRecentMessages(ctx, &chat.GetPrivateRecentMessagesIn{
			ChatUuid: chatUUID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch chat")
	})
}

func TestServer_EditPrivateMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	chatUUID := uuid.New().String()
	messageUUID := uuid.New()
	newContent := "this is the new content"
	updateAt := time.Now()

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, userUUID)

	s := New(mockRepo, mockUserClient)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditPrivateMessage")

		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).
			Return(true, nil)

		deletionInfo := &model.DeletionInfo{}
		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID.String()).Return(deletionInfo, nil)

		mockRepo.EXPECT().IsMessageOwner(ctx, chatUUID, messageUUID.String(), userUUID).
			Return(true, nil)

		updatedMessage := &model.EditedMessage{
			MessageUUID: messageUUID,
			Content:     newContent,
			UpdateAt:    updateAt,
		}
		mockRepo.EXPECT().EditPrivateMessage(ctx, messageUUID.String(), newContent).Return(updatedMessage, nil)

		response, err := s.EditPrivateMessage(ctx, &chat.EditPrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID.String(),
			NewContent:  newContent,
		})

		assert.NoError(t, err)
		assert.Equal(t, newContent, response.NewContent)
		assert.Equal(t, updateAt.Format(time.RFC3339), response.UpdatedAt)
	})

	t.Run("no_userUUID", func(t *testing.T) {
		badCtx := context.WithValue(context.Background(), config.KeyLogger, mockLogger)

		mockLogger.EXPECT().AddFuncName("EditPrivateMessage")
		mockLogger.EXPECT().Error("failed to find uuid")

		_, err := s.EditPrivateMessage(badCtx, &chat.EditPrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID.String(),
			NewContent:  newContent,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find uuid")
	})

	t.Run("IsChatMember_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditPrivateMessage")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().IsChatMember(ctx, gomock.Any(), gomock.Any()).
			Return(false, fmt.Errorf("failed to check user in chat"))

		_, err := s.EditPrivateMessage(ctx, &chat.EditPrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID.String(),
			NewContent:  newContent,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check user in chat")
	})

	t.Run("isMember_false", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditPrivateMessage")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().IsChatMember(ctx, gomock.Any(), gomock.Any()).
			Return(false, nil)

		_, err := s.EditPrivateMessage(ctx, &chat.EditPrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID.String(),
			NewContent:  newContent,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to user is not chat member")
	})

	t.Run("GetPrivateDeletionInfo_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditPrivateMessage")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().IsChatMember(ctx, gomock.Any(), gomock.Any()).
			Return(true, nil)

		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID.String()).
			Return(nil, fmt.Errorf("failed to check deletion status"))

		_, err := s.EditPrivateMessage(ctx, &chat.EditPrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID.String(),
			NewContent:  newContent,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check deletion status")
	})

	t.Run("Error_checking_deletion_status", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditPrivateMessage")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().IsChatMember(ctx, gomock.Any(), gomock.Any()).
			Return(true, nil)

		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID.String()).
			Return(&model.DeletionInfo{
				DeletedAt: "time",
			}, nil)

		_, err := s.EditPrivateMessage(ctx, &chat.EditPrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID.String(),
			NewContent:  newContent,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to edit deleted message")
	})

	t.Run("DB_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditPrivateMessage")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).
			Return(true, nil)

		deletionInfo := &model.DeletionInfo{}
		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID.String()).Return(deletionInfo, nil)

		mockRepo.EXPECT().IsMessageOwner(ctx, chatUUID, messageUUID.String(), userUUID).
			Return(true, nil)

		mockRepo.EXPECT().EditPrivateMessage(ctx, messageUUID.String(), newContent).Return(nil, fmt.Errorf("failed to edit private message"))

		_, err := s.EditPrivateMessage(ctx, &chat.EditPrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID.String(),
			NewContent:  newContent,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to edit private message")
	})

	t.Run("NotMessageOwner", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditPrivateMessage")
		mockLogger.EXPECT().Error("failed to user is not message owner")

		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).
			Return(true, nil)

		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID.String()).
			Return(&model.DeletionInfo{}, nil)

		mockRepo.EXPECT().IsMessageOwner(ctx, chatUUID, messageUUID.String(), userUUID).
			Return(false, nil)

		_, err := s.EditPrivateMessage(ctx, &chat.EditPrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID.String(),
			NewContent:  newContent,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to user is not message owner")
	})

	t.Run("IsMessageOwner_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditPrivateMessage")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).
			Return(true, nil)

		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID.String()).
			Return(&model.DeletionInfo{}, nil)

		mockRepo.EXPECT().IsMessageOwner(ctx, chatUUID, messageUUID.String(), userUUID).
			Return(false, fmt.Errorf("some DB error"))

		_, err := s.EditPrivateMessage(ctx, &chat.EditPrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID.String(),
			NewContent:  newContent,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check message owner")
	})
}

func TestServer_GetChats(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	expectedLastMessageTime := time.Now()

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, userUUID)

	s := New(mockRepo, mockUserClient)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetChats")

		expPrivateChats := &model.ChatInfoList{
			{
				LastMessage:          "How are you?",
				ChatName:             "Private chat name",
				AvatarURL:            "standart avatar url",
				LastMessageTimestamp: &expectedLastMessageTime,
				ChatUUID:             uuid.New().String(),
			},
		}

		expGroupChats := &model.ChatInfoList{
			{
				LastMessage:          "Hello!",
				ChatName:             "Group chat name",
				AvatarURL:            "standart avatar url",
				LastMessageTimestamp: &expectedLastMessageTime,
				ChatUUID:             uuid.New().String(),
			},
		}

		mockRepo.EXPECT().GetPrivateChats(ctx, userUUID).Return(expPrivateChats, nil)
		mockRepo.EXPECT().GetGroupChats(ctx, userUUID).Return(expGroupChats, nil)

		chats, err := s.GetChats(ctx, &chat.ChatEmpty{})

		assert.NoError(t, err)
		assert.NotNil(t, chats)
		assert.Len(t, chats.Chats, 2)
	})

	t.Run("no_userUUID", func(t *testing.T) {
		badCtx := context.WithValue(context.Background(), config.KeyLogger, mockLogger)

		mockLogger.EXPECT().AddFuncName("GetChats")
		mockLogger.EXPECT().Error("failed to find userUUID")

		_, err := s.GetChats(badCtx, &chat.ChatEmpty{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find userUUID")
	})

	t.Run("DB_private_error", func(t *testing.T) {
		expectedErr := fmt.Errorf("failed to get private chats")

		mockLogger.EXPECT().AddFuncName("GetChats")
		mockLogger.EXPECT().Error(gomock.Any())
		mockRepo.EXPECT().GetPrivateChats(ctx, userUUID).Return(nil, expectedErr)

		_, err := s.GetChats(ctx, &chat.ChatEmpty{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})

	t.Run("DB_group_error", func(t *testing.T) {
		expectedErr := fmt.Errorf("failed to get group chats")

		mockLogger.EXPECT().AddFuncName("GetChats")
		mockLogger.EXPECT().Error(gomock.Any())
		mockRepo.EXPECT().GetPrivateChats(ctx, userUUID).Return(&model.ChatInfoList{}, nil)
		mockRepo.EXPECT().GetGroupChats(ctx, userUUID).Return(nil, expectedErr)

		_, err := s.GetChats(ctx, &chat.ChatEmpty{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})
}

func TestServer_DeletePrivateMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	chatUUID := uuid.New().String()
	messageUUID := uuid.New().String()

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, userUUID)

	s := New(mockRepo, mockUserClient)

	t.Run("success_self_to_all", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeletePrivateMessage")
		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID).Return(&model.DeletionInfo{
			DeletedBy:    uuid.New().String(),
			DeleteFormat: model.Self,
			DeletedAt:    time.Now().Format(time.RFC3339),
		}, nil)
		mockRepo.EXPECT().DeletePrivateMessage(ctx, userUUID, messageUUID, model.All).Return(true, nil)

		isDeleted, err := s.DeletePrivateMessage(ctx, &chat.DeletePrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID,
			Mode:        model.All,
		})

		assert.NoError(t, err)
		assert.Equal(t, true, isDeleted.DeletionStatus)
	})

	t.Run("success_direct_all", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeletePrivateMessage")
		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID).Return(&model.DeletionInfo{}, nil)
		mockRepo.EXPECT().DeletePrivateMessage(ctx, userUUID, messageUUID, model.All).Return(true, nil)

		isDeleted, err := s.DeletePrivateMessage(ctx, &chat.DeletePrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID,
			Mode:        model.All,
		})

		assert.NoError(t, err)
		assert.Equal(t, true, isDeleted.DeletionStatus)
	})

	t.Run("no_userUUID", func(t *testing.T) {
		badCtx := context.WithValue(context.Background(), config.KeyLogger, mockLogger)
		mockLogger.EXPECT().AddFuncName("DeletePrivateMessage")
		mockLogger.EXPECT().Error("failed to find uuid")

		_, err := s.DeletePrivateMessage(badCtx, &chat.DeletePrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID,
			Mode:        model.All,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find uuid")
	})

	t.Run("not_chat_member", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeletePrivateMessage")
		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).Return(false, nil)
		mockLogger.EXPECT().Error("failed to user is not chat member")

		_, err := s.DeletePrivateMessage(ctx, &chat.DeletePrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID,
			Mode:        model.All,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to user is not chat member")
	})

	t.Run("is_chat_member_error", func(t *testing.T) {
		expectedErr := fmt.Errorf("db error")
		mockLogger.EXPECT().AddFuncName("DeletePrivateMessage")
		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).Return(false, expectedErr)
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to check if user is chat member: %v", expectedErr))

		_, err := s.DeletePrivateMessage(ctx, &chat.DeletePrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID,
			Mode:        model.All,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check user is chat member")
	})

	t.Run("invalid_mode", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeletePrivateMessage")
		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).Return(true, nil)
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to invalid mode: %s", "invalid"))

		_, err := s.DeletePrivateMessage(ctx, &chat.DeletePrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID,
			Mode:        "invalid",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to invalid mode")
	})

	t.Run("already_deleted_all", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeletePrivateMessage")
		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID).Return(&model.DeletionInfo{
			DeletedBy:    uuid.New().String(),
			DeleteFormat: model.All,
			DeletedAt:    time.Now().Format(time.RFC3339),
		}, nil)
		mockLogger.EXPECT().Error("failed to message is already deleted")

		_, err := s.DeletePrivateMessage(ctx, &chat.DeletePrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID,
			Mode:        model.All,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to message is already deleted")
	})

	t.Run("already_deleted_self_by_user", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeletePrivateMessage")
		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID).Return(&model.DeletionInfo{
			DeletedBy:    userUUID,
			DeleteFormat: model.Self,
			DeletedAt:    time.Now().Format(time.RFC3339),
		}, nil)
		mockLogger.EXPECT().Error("failed to message is already deleted")

		_, err := s.DeletePrivateMessage(ctx, &chat.DeletePrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID,
			Mode:        model.Self,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to message is already deleted")
	})

	t.Run("get_deletion_info_error", func(t *testing.T) {
		expectedErr := fmt.Errorf("db error")
		mockLogger.EXPECT().AddFuncName("DeletePrivateMessage")
		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID).Return(nil, expectedErr)
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to get private message deletion info: %v", expectedErr))

		_, err := s.DeletePrivateMessage(ctx, &chat.DeletePrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID,
			Mode:        model.All,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get private message deletion info")
	})

	t.Run("delete_message_error", func(t *testing.T) {
		expectedErr := fmt.Errorf("db error")
		mockLogger.EXPECT().AddFuncName("DeletePrivateMessage")
		mockRepo.EXPECT().IsChatMember(ctx, chatUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetPrivateDeletionInfo(ctx, messageUUID).Return(&model.DeletionInfo{}, nil)
		mockRepo.EXPECT().DeletePrivateMessage(ctx, userUUID, messageUUID, model.All).Return(false, expectedErr)
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to delete private message: %v", expectedErr))

		_, err := s.DeletePrivateMessage(ctx, &chat.DeletePrivateMessageIn{
			ChatUuid:    chatUUID,
			MessageUuid: messageUUID,
			Mode:        model.All,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete private message")
	})
}
