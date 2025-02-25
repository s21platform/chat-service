package client

import (
	"context"
	"fmt"
	"log"

	userproto "github.com/s21platform/user-proto/user-proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

func (c *Service) GetUserInfoByUUID(ctx context.Context, userUUID string) (*model.UserInfo, error) {
	resp, err := c.client.GetUserInfoByUUID(ctx, &userproto.GetUserInfoByUUIDIn{Uuid: userUUID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from user-service: %v", err)
	}

	return &model.UserInfo{
		UserName:   resp.Nickname,
		AvatarLink: resp.Avatar,
	}, nil
}
