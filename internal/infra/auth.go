package infra

import (
	"context"
	"github.com/s21platform/chat-service/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	_ = info

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no info in metadata")
	}

	userIDs, ok := md["user_uuid"]
	if !ok || len(userIDs) != 1 {
		return nil, status.Errorf(codes.Unauthenticated, "no uuid or more than one in metadata")
	}

	ctx = context.WithValue(ctx, config.KeyUUID, userIDs[0])

	return handler(ctx, req)
}
