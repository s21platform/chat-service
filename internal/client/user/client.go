package client

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	userproto "github.com/s21platform/user-proto/user-proto"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
)

type Service struct {
	client userproto.UserServiceClient
}

func NewService(cfg *config.Config) *Service {
	connStr := fmt.Sprintf("%s:%s", cfg.UserService.Host, cfg.UserService.Port)

	conn, err := grpc.NewClient(connStr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to user-service: %v", err)
	}

	client := userproto.NewUserServiceClient(conn)

	return &Service{client: client}
}

func (s *Service) GetUserInfoByUUID(ctx context.Context, userUUID string) (*model.ChatMemberParams, error) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("uuid", userUUID))

	resp, err := s.client.GetUserInfoByUUID(ctx, &userproto.GetUserInfoByUUIDIn{Uuid: userUUID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from user-service: %v", err)
	}

	return &model.ChatMemberParams{
		UserUUID:   userUUID,
		Nickname:   resp.Nickname,
		AvatarLink: resp.Avatar,
	}, nil
}
