package service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	chat "github.com/s21platform/chat-proto/chat-proto"
	"github.com/stretchr/testify/assert"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
	logger_lib "github.com/s21platform/logger-lib"
)

func TestServer_CreatePrivateChat_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	initiatorUUID := uuid.New().String()
	companionUUID := uuid.New().String()

	mockLogger := logger_lib.New("localhost", "8080", "chat-service", "test")
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, initiatorUUID)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)

	mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
		Return(&model.UserInfo{
			UserName:   "test_user",
			AvatarLink: "test_avatar_link",
		}, nil)

	mockRepo.EXPECT().CreatePrivateChat(gomock.Any()).
		Return("chat_uuid", nil)

	s := New(mockRepo, mockUserClient)

	out, err := s.CreatePrivateChat(ctx, &chat.CreatePrivateChatIn{
		CompanionUuid: companionUUID,
	})
	
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, "chat_uuid", out.NewChatUuid)
}
