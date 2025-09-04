package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/client/centrifugo"
	"github.com/s21platform/chat-service/internal/client/user"
	"github.com/s21platform/chat-service/internal/config"
	api "github.com/s21platform/chat-service/internal/generated"
	"github.com/s21platform/chat-service/internal/infra"
	"github.com/s21platform/chat-service/internal/pkg/jwt"
	"github.com/s21platform/chat-service/internal/pkg/tx"
	"github.com/s21platform/chat-service/internal/pkg/validator"
	db "github.com/s21platform/chat-service/internal/repository/postgres"
	"github.com/s21platform/chat-service/internal/rest"
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

	chatService := service.New()
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			infra.AuthInterceptorGRPC,
			infra.LoggerGRPC(logger),
			tx.TxMiddlewareGRPC(dbRepo),
		),
	)
	chat.RegisterChatServiceServer(grpcServer, chatService)

	handler := rest.New(dbRepo, userClient, centrifugeClient, vldtr, jwtGenerator)
	router := chi.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return infra.AuthInterceptorHTTP(next)
	})
	router.Use(func(next http.Handler) http.Handler {
		return infra.LoggerHTTP(next, logger)
	})
	router.Use(func(next http.Handler) http.Handler {
		return tx.TxMiddlewareHTTP(dbRepo)(next)
	})

	api.HandlerFromMux(handler, router)
	httpServer := &http.Server{
		Handler: router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Service.Port))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to start TCP listener: %v", err))
	}

	m := cmux.New(listener)

	grpcListener := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("gRPC server error: %v", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := httpServer.Serve(httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %v", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := m.Serve(); err != nil {
			return fmt.Errorf("cannot start service: %v", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("server error: %v", err))
	}
}
