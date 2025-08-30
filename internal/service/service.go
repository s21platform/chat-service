package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
	"github.com/s21platform/chat-service/internal/pkg/tx"
	"github.com/s21platform/chat-service/pkg/chat"
)

type Server struct {
	chat.UnimplementedChatServiceServer
	repository       DBRepo
	userClient       UserClient
	centrifugeClient CetrifugeClient
	validator        Validator
	jwtGenerator     JWTGenerator
}

func New(
	repo DBRepo,
	userClient UserClient,
	centrifugeClient CetrifugeClient,
	validator Validator,
	jwtGenerator JWTGenerator,
) *Server {
	return &Server{
		repository:       repo,
		userClient:       userClient,
		centrifugeClient: centrifugeClient,
		validator:        validator,
		jwtGenerator:     jwtGenerator,
	}
}

func (s *Server) CreateStream(ctx context.Context, in *chat.CreateStreamIn) (*chat.CreateStreamOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("CreateStream")

	creatorID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get creator ID")
		return nil, status.Error(codes.Internal, "failed to get creator ID")
	}

	if err := s.validator.ValidateStreamByType(in, creatorID); err != nil {
		logger.Error(fmt.Sprintf("stream validation failed: %v", err))
		return nil, status.Errorf(codes.InvalidArgument, "stream validation failed: %v", err)
	}

	var streamID string
	err := tx.TxExecute(ctx, func(ctx context.Context) error {
		allUserIDs := []string{creatorID}
		for _, user := range in.Users {
			if strings.TrimSpace(user.Id) != "" && user.Id != creatorID {
				allUserIDs = append(allUserIDs, user.Id)
			}
		}

		for _, userID := range allUserIDs {
			userInfo, err := s.userClient.GetUserInfoByUUID(ctx, userID)
			if err != nil {
				logger.Error(fmt.Sprintf("failed to get user info for %s: %v", userID, err))
				return fmt.Errorf("failed to get user info for %s: %v", userID, err)
			}

			err = s.repository.AddNewUser(ctx, userInfo)
			if err != nil {
				logger.Error(fmt.Sprintf("failed to add user %s to users table: %v", userID, err))
				return fmt.Errorf("failed to add user %s to users table: %v", userID, err)
			}
		}

		var err error
		streamID, err = s.repository.CreateStream(ctx, in.Type, in.ChatMetadata, creatorID)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to create stream: %v", err))
			return err
		}

		var members []model.StreamMember
		members = append(members, model.StreamMember{
			UserID:   creatorID,
			Metadata: in.CreatorMetadata,
		})

		for _, user := range in.Users {
			if strings.TrimSpace(user.Id) != "" && user.Id != creatorID {
				members = append(members, model.StreamMember{
					UserID:   user.Id,
					Metadata: user.Metadata,
				})
			}
		}

		err = s.repository.AddStreamMembers(ctx, streamID, members)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to add stream members: %v", err))
			return err
		}

		var subscriptions []model.UserSubscription

		for _, member := range members {
			subscriptions = append(subscriptions, model.UserSubscription{
				UserID:  member.UserID,
				Channel: streamID,
			})
		}

		err = s.repository.AddUserSubscriptions(ctx, subscriptions)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to create subscriptions: %v", err))
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error(fmt.Sprintf("failed to complete stream creation transaction: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to create stream: %v", err)
	}

	return &chat.CreateStreamOut{
		Id: streamID,
	}, nil
}

func (s *Server) SendMessage(ctx context.Context, in *chat.SendMessageIn) (*chat.SendMessageOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("SendMessage")

	senderID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get sender ID")
		return nil, status.Error(codes.Internal, "failed to get sender ID")
	}

	if err := s.validator.ValidateSendMessage(in); err != nil {
		logger.Error(fmt.Sprintf("message validation failed: %v", err))
		return nil, status.Errorf(codes.InvalidArgument, "message validation failed: %v", err)
	}

	var message model.Message
	err := tx.TxExecute(ctx, func(ctx context.Context) error {
		isMember, err := s.repository.IsStreamMember(ctx, in.StreamId, senderID)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to check stream membership: %v", err))
			return fmt.Errorf("failed to check stream membership: %v", err)
		}

		if !isMember {
			logger.Error(fmt.Sprintf("user %s is not a member of stream %s", senderID, in.StreamId))
			return fmt.Errorf("user is not a member of this stream")
		}

		message = model.Message{
			ID:       uuid.New(),
			StreamID: uuid.MustParse(in.StreamId),
			SenderID: uuid.MustParse(senderID),
			Type:     in.MessageType,
			Content:  in.Content,
			SentAt:   time.Now(),
		}

		if in.ParentId != nil && *in.ParentId != "" {
			parentUUID := uuid.MustParse(*in.ParentId)
			message.ParentID = &parentUUID
		}

		if in.RootId != nil && *in.RootId != "" {
			rootUUID := uuid.MustParse(*in.RootId)
			message.RootID = &rootUUID
		}

		err = s.repository.SaveMessage(ctx, &message)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to save message: %v", err))
			return fmt.Errorf("failed to save message: %v", err)
		}

		return nil
	})

	if err != nil {
		logger.Error(fmt.Sprintf("failed to send message transaction: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
	}

	// todo спросить будет ли формировать аву сообщения, никнейм отправителя и тп фронт по uuid пользака
	channel := message.StreamID.String()
	err = s.centrifugeClient.Publish(ctx, channel, message)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to publish message to stream: %v", err))
	}

	return &chat.SendMessageOut{
		MessageId: message.ID.String(),
		SentAt:    message.SentAt.Format(time.RFC3339),
	}, nil
}

func (s *Server) GetPrivateStreams(ctx context.Context, _ *emptypb.Empty) (*chat.GetPrivateStreamsOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetPrivateStreams")

	requesterID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get requester id")
		return nil, status.Error(codes.Internal, "failed to get requester id")
	}

	privateStreams, err := s.repository.GetPrivateStreams(ctx, requesterID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get private streams: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get private streams: %v", err)
	}

	return &chat.GetPrivateStreamsOut{
		Streams: privateStreams.FromDTO(),
	}, nil
}

func (s *Server) GetStreamRecentMessages(ctx context.Context, in *chat.GetStreamRecentMessagesIn) (*chat.GetStreamRecentMessagesOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetStreamRecentMessages")

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to find uuid")
		return nil, status.Error(codes.Internal, "failed to find uuid")
	}

	isMember, err := s.repository.IsStreamMember(ctx, in.StreamId, userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to check stream membership: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to check stream membership: %v", err)
	}

	if !isMember {
		logger.Error("user is not a member of the stream")
		return nil, status.Error(codes.PermissionDenied, "user is not a member of the stream")
	}

	messages, err := s.repository.GetStreamRecentMessages(ctx, in.StreamId, in.Offset, in.Limit)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to fetch messages: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to fetch messages: %v", err)
	}

	return &chat.GetStreamRecentMessagesOut{
		Messages: messages.FromDTO(),
	}, nil
}

func (s *Server) GetConnectAccessToken(ctx context.Context, _ *emptypb.Empty) (*chat.GetConnectAccessTokenOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetConnectAccessToken")

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get user UUID")
		return nil, status.Error(codes.Internal, "failed to get user UUID")
	}

	token, expiresAt, err := s.jwtGenerator.GenerateConnectToken(userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to generate access token: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
	}

	logger.Info(fmt.Sprintf("generated access token for user %s", userUUID))

	return &chat.GetConnectAccessTokenOut{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Server) GetStreamSubscribeToken(ctx context.Context, in *chat.GetStreamSubscribeTokenIn) (*chat.GetStreamSubscribeTokenOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetStreamSubscribeToken")

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get user UUID")
		return nil, status.Error(codes.Internal, "failed to get user UUID")
	}

	isMember, err := s.repository.IsStreamMember(ctx, in.StreamId, userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to check stream membership: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to check stream membership: %v", err)
	}

	if !isMember {
		logger.Error("user is not a member of the stream")
		return nil, status.Error(codes.PermissionDenied, "user is not a member of the stream")
	}

	token, expiresAt, err := s.jwtGenerator.GenerateSubscribeToken(userUUID, in.StreamId)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to generate subscribe token: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to generate subscribe token: %v", err)
	}

	logger.Info(fmt.Sprintf("generated subscribe token for user %s, stream %s", userUUID, in.StreamId))

	return &chat.GetStreamSubscribeTokenOut{
		Token:     token,
		ExpiresAt: expiresAt,
		Channel:   in.StreamId,
	}, nil
}

func (s *Server) GetUserActiveStreams(ctx context.Context, _ *emptypb.Empty) (*chat.GetUserActiveStreamsOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetUserActiveStreams")

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get user UUID")
		return nil, status.Error(codes.Internal, "failed to get user UUID")
	}

	streamIDs, err := s.repository.GetUserActiveStreams(ctx, userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get user active streams: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get user active streams: %v", err)
	}

	return &chat.GetUserActiveStreamsOut{
		StreamIds: streamIDs,
	}, nil
}

func (s *Server) GetBatchSubscribeTokens(ctx context.Context, in *chat.GetBatchSubscribeTokensIn) (*chat.GetBatchSubscribeTokensOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetBatchSubscribeTokens")

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get user UUID")
		return nil, status.Error(codes.Internal, "failed to get user UUID")
	}

	var subscriptions []*chat.StreamSubscription

	for _, streamID := range in.StreamIds {
		isMember, err := s.repository.IsStreamMember(ctx, streamID, userUUID)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to check stream membership for %s: %v", streamID, err))
			continue
		}

		if !isMember {
			logger.Warn(fmt.Sprintf("user %s is not a member of stream %s, skipping", userUUID, streamID))
			continue
		}

		token, expiresAt, err := s.jwtGenerator.GenerateSubscribeToken(userUUID, streamID)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to generate subscribe token for stream %s: %v", streamID, err))
			continue
		}

		subscriptions = append(subscriptions, &chat.StreamSubscription{
			StreamId:  streamID,
			Token:     token,
			ExpiresAt: expiresAt,
			Channel:   streamID,
		})
	}

	return &chat.GetBatchSubscribeTokensOut{
		Subscriptions: subscriptions,
	}, nil
}
