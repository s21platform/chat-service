package infra

import (
	"context"
	"net/http"

	"google.golang.org/grpc"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
)

func LoggerGRPC(logger *logger_lib.Logger) func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = context.WithValue(ctx, config.KeyLogger, logger)
		return handler(ctx, req)
	}
}

func LoggerHTTP(next http.Handler, logger *logger_lib.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), config.KeyLogger, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
