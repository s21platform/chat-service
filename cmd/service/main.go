package main

import (
	"fmt"
	"github.com/s21platform/chat-service/internal/client/user"
	"net"

	_ "github.com/lib/pq" // PostgreSQL driver
	"google.golang.org/grpc"

	chat "github.com/s21platform/chat-proto/chat-proto"
	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/infra"
	db "github.com/s21platform/chat-service/internal/repository/postgres"
	"github.com/s21platform/chat-service/internal/service"
)

func main() {
	cfg := config.MustLoad()
	logger := logger_lib.New(cfg.Logger.Host, cfg.Logger.Port, cfg.Service.Name, cfg.Platform.Env)

	dbRepo := db.New(cfg)
	defer dbRepo.Close()

	userClient := client.NewService(cfg)

	chatService := service.New(dbRepo, userClient)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			infra.AuthInterceptor,
			infra.Logger(logger),
		),
	)

	chat.RegisterChatServiceServer(server, chatService)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Service.Port))
	if err != nil {
		logger.Error(fmt.Sprintf("Cannot listen port: %s; Error: %v", cfg.Service.Port, err))
	}
	if err = server.Serve(lis); err != nil {
		logger.Error(fmt.Sprintf("Cannot start grpc, port: %s; Error: %v", cfg.Service.Port, err))
	}
}
