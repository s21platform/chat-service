package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	chat "github.com/s21platform/chat-proto/chat-proto"
	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
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

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
			Return(&model.UserInfo{
				UserName:   "test_user",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockRepo.EXPECT().CreatePrivateChat(gomock.Any()).
			Return("chat_uuid", nil)

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

	t.Run("get_companionInfo_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreatePrivateChat")
		mockLogger.EXPECT().Error(gomock.Any())

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

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
			Return(&model.UserInfo{
				UserName:   "test_user",
				AvatarLink: "test_avatar_link",
			}, nil)

		mockRepo.EXPECT().CreatePrivateChat(gomock.Any()).
			Return("", fmt.Errorf("failed to create chat"))

		_, err := s.CreatePrivateChat(ctx, &chat.CreatePrivateChatIn{
			CompanionUuid: companionUUID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create chat")
	})
}
