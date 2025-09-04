package infra

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"google.golang.org/grpc"

	"github.com/s21platform/chat-service/internal/config"
)

func AuthInterceptorGRPC(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	// todo добавить логику обработки межсервисного токена

	return handler(ctx, req)
}

func AuthInterceptorHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		userID = strings.TrimSpace(userID)

		if userID == "" {
			writeErrorResponse(w, "missing or empty X-User-ID header", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), config.KeyUUID, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]string{
		"error": message,
	}

	_ = json.NewEncoder(w).Encode(errorResponse)
}
