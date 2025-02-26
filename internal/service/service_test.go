package service

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/s21platform/chat-service/internal/client/user"
	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
	"testing"
)

func TestServer_CreatePrivateChat_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	initiatorUUID := uuid.New().String()
	companionUUID := uuid.New().String()

	ctx = context.WithValue(ctx, config.KeyUUID, initiatorUUID)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)

	userService := &client.Service{client: mockUserClient}

	mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
		Return(&model.UserInfo{
			UserName:   "test_user",
			AvatarLink: "test_avatar_link",
		}, nil)

	mockRepo.EXPECT().CreatePrivateChat(gomock.Any()).
		Return("chat_uuid", nil)

	s := New(mockRepo, userService)

	out, err := s.repository.CreatePrivateChat()
}
