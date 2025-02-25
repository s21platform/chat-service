package user

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

type Client struct {
	client userproto.UserServiceClient
}

func MustConnect(cfg *config.Config) *Client {
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", cfg.UserService.Host, cfg.UserService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to user-service: %v", err)
	}
	
	uClient := userproto.NewUserServiceClient(conn)

	return &Client{client: uClient}
}

func (c *Client) GetUserInfoByUUID(ctx context.Context, userUUID string) (*model.UserInfo, error) {
	// Запрос по gRPC
	resp, err := c.client.GetUserInfoByUUID(ctx, &userproto.GetUserInfoByUUIDIn{Uuid: userUUID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from user-service: %v", err)
	}
	// Преобразование приходящего ответа в model проекта
	return &model.UserInfo{
		UserName:   resp.Nickname,
		AvatarLink: resp.Avatar,
	}, nil
}
