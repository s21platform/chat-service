package main

import (
	"fmt"
	"net"

	_ "github.com/lib/pq" // PostgreSQL driver
	"google.golang.org/grpc"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/client/user"
	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/infra"
	db "github.com/s21platform/chat-service/internal/repository/postgres"
	"github.com/s21platform/chat-service/internal/service"
	"github.com/s21platform/chat-service/pkg/chat"
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

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Service.Port))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to start TCP listener: %v", err))
	}

	if err = server.Serve(listener); err != nil {
		logger.Error(fmt.Sprintf("failed to start gRPC listener: %v", err))
	}
}
