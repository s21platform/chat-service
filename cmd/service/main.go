package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/client/centrifugo"
	"github.com/s21platform/chat-service/internal/client/user"
	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/infra"
	"github.com/s21platform/chat-service/internal/pkg/jwt"
	"github.com/s21platform/chat-service/internal/pkg/tx"
	"github.com/s21platform/chat-service/internal/pkg/validator"
	db "github.com/s21platform/chat-service/internal/repository/postgres"
	"github.com/s21platform/chat-service/internal/service"
	"github.com/s21platform/chat-service/pkg/chat"
)

func main() {
	cfg := config.MustLoad()
	logger := logger_lib.New(cfg.Logger.Host, cfg.Logger.Port, cfg.Service.Name, cfg.Platform.Env)

	dbRepo := db.New(cfg)
	defer dbRepo.Close()

	userClient := user.New(cfg)

	centrifugeClient := centrifugo.New(cfg)
	defer centrifugeClient.Close()

	vldtr := validator.New()
	jwtGenerator := jwt.New(cfg.Centrifuge.JWTSecret)

	chatService := service.New(dbRepo, userClient, centrifugeClient, vldtr, jwtGenerator)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			infra.AuthInterceptor,
			infra.Logger(logger),
			tx.TxMiddleware(dbRepo),
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
