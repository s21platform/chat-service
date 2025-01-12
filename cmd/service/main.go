package main

import (
	"fmt"
	"net"

	"github.com/s21platform/chat-service/internal/infra"
	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	db "github.com/s21platform/chat-service/internal/repository/postgres"
	"github.com/s21platform/chat-service/internal/service"
	"google.golang.org/grpc"

	chat "github.com/s21platform/chat-proto/chat-proto"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	cfg := config.MustLoad()
	logger := logger_lib.New(cfg.Logger.Host, cfg.Logger.Port, cfg.Service.Name, cfg.Platform.Env)

	dbRepo := db.New(cfg)
	defer dbRepo.Close()

	chatService := service.New(dbRepo)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(infra.Logger(logger)),
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
