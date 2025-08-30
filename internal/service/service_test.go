package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
	"github.com/s21platform/chat-service/internal/pkg/tx"
	"github.com/s21platform/chat-service/pkg/chat"
)

func createTxContext(ctx context.Context, mockRepo *MockDBRepo) context.Context {
	return context.WithValue(ctx, tx.KeyTx, tx.Tx{DbRepo: mockRepo})
}

func TestServer_GetStreamRecentMessages(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)
	mockValidator := NewMockValidator(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	streamID := uuid.New().String()
	offset := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	limit := int32(50)

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, userUUID)

	s := New(mockRepo, mockUserClient, nil, mockValidator, nil)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetStreamRecentMessages")

		expectedMessages := &model.MessageList{
			{
				ID:        uuid.New(),
				StreamID:  uuid.MustParse(streamID),
				SenderID:  uuid.New(),
				Type:      "text",
				Content:   "message 1",
				RootID:    nil,
				ParentID:  nil,
				SentAt:    time.Now().Add(-10 * time.Minute),
				UpdatedAt: nil,
			},
			{
				ID:        uuid.New(),
				StreamID:  uuid.MustParse(streamID),
				SenderID:  uuid.New(),
				Type:      "text",
				Content:   "message 2",
				RootID:    nil,
				ParentID:  nil,
				SentAt:    time.Now().Add(-5 * time.Minute),
				UpdatedAt: nil,
			},
		}

		mockRepo.EXPECT().IsStreamMember(ctx, streamID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetStreamRecentMessages(ctx, streamID, offset, limit).Return(expectedMessages, nil)

		messages, err := s.GetStreamRecentMessages(ctx, &chat.GetStreamRecentMessagesIn{
			StreamId: streamID,
			Offset:   offset,
			Limit:    limit,
		})

		assert.NoError(t, err)
		assert.NotNil(t, messages)
		assert.Len(t, messages.Messages, 2)
	})

	t.Run("no_userUUID", func(t *testing.T) {
		badCtx := context.WithValue(context.Background(), config.KeyLogger, mockLogger)

		mockLogger.EXPECT().AddFuncName("GetStreamRecentMessages")
		mockLogger.EXPECT().Error("failed to find uuid")

		_, err := s.GetStreamRecentMessages(badCtx, &chat.GetStreamRecentMessagesIn{
			StreamId: streamID,
			Offset:   offset,
			Limit:    limit,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find uuid")
	})

	t.Run("not_stream_member", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetStreamRecentMessages")

		mockRepo.EXPECT().IsStreamMember(ctx, streamID, userUUID).Return(false, nil)
		mockLogger.EXPECT().Error("user is not a member of the stream")

		_, err := s.GetStreamRecentMessages(ctx, &chat.GetStreamRecentMessagesIn{
			StreamId: streamID,
			Offset:   offset,
			Limit:    limit,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user is not a member of the stream")
	})

	t.Run("DB_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetStreamRecentMessages")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().IsStreamMember(ctx, streamID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetStreamRecentMessages(ctx, streamID, offset, limit).Return(nil, fmt.Errorf("failed to fetch messages"))

		_, err := s.GetStreamRecentMessages(ctx, &chat.GetStreamRecentMessagesIn{
			StreamId: streamID,
			Offset:   offset,
			Limit:    limit,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch messages")
	})

	t.Run("membership_check_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetStreamRecentMessages")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().IsStreamMember(ctx, streamID, userUUID).
			Return(false, fmt.Errorf("db connection failed"))

		_, err := s.GetStreamRecentMessages(ctx, &chat.GetStreamRecentMessagesIn{
			StreamId: streamID,
			Offset:   offset,
			Limit:    limit,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check stream membership")
	})

	t.Run("success_empty_offset", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetStreamRecentMessages")

		emptyMessages := &model.MessageList{}
		mockRepo.EXPECT().IsStreamMember(ctx, streamID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetStreamRecentMessages(ctx, streamID, "", int32(10)).Return(emptyMessages, nil)

		messages, err := s.GetStreamRecentMessages(ctx, &chat.GetStreamRecentMessagesIn{
			StreamId: streamID,
			Offset:   "",
			Limit:    10,
		})

		assert.NoError(t, err)
		assert.NotNil(t, messages)
		assert.Len(t, messages.Messages, 0)
	})
}

func TestServer_CreateStream(t *testing.T) {
	t.Parallel()

	creatorUUID := uuid.New().String()
	companionUUID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, creatorUUID)

		ctx = createTxContext(ctx, mockRepo)

		s := New(mockRepo, mockUserClient, nil, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("CreateStream")

		mockValidator.EXPECT().ValidateStreamByType(gomock.Any(), creatorUUID).Return(nil)

		mockRepo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, creatorUUID).
			Return(&model.StreamMemberParams{
				UserID:    creatorUUID,
				Nickname:  "test_creator",
				AvatarURL: "test_avatar",
			}, nil)

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
			Return(&model.StreamMemberParams{
				UserID:    companionUUID,
				Nickname:  "test_companion",
				AvatarURL: "test_avatar",
			}, nil)

		mockRepo.EXPECT().AddNewUser(ctx, gomock.Any()).Return(nil).Times(2)
		mockRepo.EXPECT().CreateStream(ctx, "private", gomock.Any(), creatorUUID).Return("test-stream-id", nil)
		mockRepo.EXPECT().AddStreamMembers(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockRepo.EXPECT().AddUserSubscriptions(ctx, gomock.Any()).Return(nil)

		result, err := s.CreateStream(ctx, &chat.CreateStreamIn{
			Users: []*chat.ChatUser{
				{Id: companionUUID, Metadata: "metadata1"},
			},
			Type:            "private",
			ChatMetadata:    "chat metadata",
			CreatorMetadata: "creator metadata",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-stream-id", result.Id)
	})

	t.Run("no_creatorID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)

		s := New(mockRepo, mockUserClient, nil, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("CreateStream")
		mockLogger.EXPECT().Error("failed to get creator ID")

		_, err := s.CreateStream(ctx, &chat.CreateStreamIn{
			Users: []*chat.ChatUser{
				{Id: companionUUID},
			},
			Type: "private",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get creator ID")
	})

	t.Run("validation_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, creatorUUID)

		s := New(mockRepo, mockUserClient, nil, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("CreateStream")
		mockLogger.EXPECT().Error(gomock.Any())

		mockValidator.EXPECT().ValidateStreamByType(gomock.Any(), creatorUUID).
			Return(fmt.Errorf("validation failed"))

		_, err := s.CreateStream(ctx, &chat.CreateStreamIn{
			Users: []*chat.ChatUser{
				{Id: companionUUID},
			},
			Type: "private",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stream validation failed")
	})

	t.Run("transaction_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, creatorUUID)

		ctx = createTxContext(ctx, mockRepo)

		s := New(mockRepo, mockUserClient, nil, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("CreateStream")
		mockLogger.EXPECT().Error(gomock.Any())

		mockValidator.EXPECT().ValidateStreamByType(gomock.Any(), creatorUUID).Return(nil)

		mockRepo.EXPECT().WithTx(ctx, gomock.Any()).Return(fmt.Errorf("transaction failed"))

		_, err := s.CreateStream(ctx, &chat.CreateStreamIn{
			Users: []*chat.ChatUser{
				{Id: companionUUID},
			},
			Type: "private",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create stream")
	})

	t.Run("user_info_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, creatorUUID)

		ctx = createTxContext(ctx, mockRepo)

		s := New(mockRepo, mockUserClient, nil, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("CreateStream")
		mockLogger.EXPECT().Error(gomock.Any()).Times(2)

		mockValidator.EXPECT().ValidateStreamByType(gomock.Any(), creatorUUID).Return(nil)

		mockRepo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, creatorUUID).
			Return(&model.StreamMemberParams{
				UserID:    creatorUUID,
				Nickname:  "test_creator",
				AvatarURL: "test_avatar",
			}, nil)

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
			Return(nil, fmt.Errorf("user not found"))

		mockRepo.EXPECT().AddNewUser(ctx, gomock.Any()).Return(nil)

		_, err := s.CreateStream(ctx, &chat.CreateStreamIn{
			Users: []*chat.ChatUser{
				{Id: companionUUID},
			},
			Type: "private",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create stream")
	})

	t.Run("create_stream_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, creatorUUID)

		ctx = createTxContext(ctx, mockRepo)

		s := New(mockRepo, mockUserClient, nil, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("CreateStream")
		mockLogger.EXPECT().Error(gomock.Any()).Times(2)

		mockValidator.EXPECT().ValidateStreamByType(gomock.Any(), creatorUUID).Return(nil)

		mockRepo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, creatorUUID).
			Return(&model.StreamMemberParams{
				UserID:    creatorUUID,
				Nickname:  "test_creator",
				AvatarURL: "test_avatar",
			}, nil)

		mockUserClient.EXPECT().GetUserInfoByUUID(ctx, companionUUID).
			Return(&model.StreamMemberParams{
				UserID:    companionUUID,
				Nickname:  "test_companion",
				AvatarURL: "test_avatar",
			}, nil)

		mockRepo.EXPECT().AddNewUser(ctx, gomock.Any()).Return(nil).Times(2)
		mockRepo.EXPECT().CreateStream(ctx, "private", gomock.Any(), creatorUUID).
			Return("", fmt.Errorf("failed to create stream"))

		_, err := s.CreateStream(ctx, &chat.CreateStreamIn{
			Users: []*chat.ChatUser{
				{Id: companionUUID},
			},
			Type: "private",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create stream")
	})
}

func TestServer_SendMessage(t *testing.T) {
	t.Parallel()

	senderUUID := uuid.New().String()
	streamID := uuid.New().String()
	parentID := uuid.New().String()
	rootID := uuid.New().String()

	t.Run("success_simple", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockCentrifuge := NewMockCetrifugeClient(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, senderUUID)

		ctx = createTxContext(ctx, mockRepo)

		s := New(mockRepo, mockUserClient, mockCentrifuge, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("SendMessage")

		mockValidator.EXPECT().ValidateSendMessage(gomock.Any()).Return(nil)
		mockRepo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		mockRepo.EXPECT().IsStreamMember(ctx, streamID, senderUUID).Return(true, nil)
		mockRepo.EXPECT().SaveMessage(ctx, gomock.Any()).Return(nil)
		mockCentrifuge.EXPECT().Publish(ctx, streamID, gomock.Any()).Return(nil)

		result, err := s.SendMessage(ctx, &chat.SendMessageIn{
			StreamId:    streamID,
			Content:     "Hello world",
			MessageType: "text",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.MessageId)
	})

	t.Run("success_with_parent_and_root", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockCentrifuge := NewMockCetrifugeClient(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, senderUUID)

		ctx = createTxContext(ctx, mockRepo)

		s := New(mockRepo, mockUserClient, mockCentrifuge, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("SendMessage")

		mockValidator.EXPECT().ValidateSendMessage(gomock.Any()).Return(nil)
		mockRepo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		mockRepo.EXPECT().IsStreamMember(ctx, streamID, senderUUID).Return(true, nil)
		mockRepo.EXPECT().SaveMessage(ctx, gomock.Any()).Return(nil)
		mockCentrifuge.EXPECT().Publish(ctx, streamID, gomock.Any()).Return(nil)

		result, err := s.SendMessage(ctx, &chat.SendMessageIn{
			StreamId:    streamID,
			Content:     "Reply message",
			ParentId:    &parentID,
			RootId:      &rootID,
			MessageType: "text",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.MessageId)
	})

	t.Run("no_senderID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockCentrifuge := NewMockCetrifugeClient(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)

		s := New(mockRepo, mockUserClient, mockCentrifuge, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("SendMessage")
		mockLogger.EXPECT().Error("failed to get sender ID")

		_, err := s.SendMessage(ctx, &chat.SendMessageIn{
			StreamId:    streamID,
			Content:     "Hello",
			MessageType: "text",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get sender ID")
	})
}

func TestServer_GetPrivateStreams(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)
	mockValidator := NewMockValidator(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, userUUID)

	s := New(mockRepo, mockUserClient, nil, mockValidator, nil)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetPrivateStreams")

		expectedStreams := &model.PrivateStreamPreviewList{
			{
				StreamID:             uuid.New().String(),
				StreamName:           "John Doe",
				AvatarURL:            "avatar.jpg",
				LastMessageContent:   "Hello there!",
				LastMessageTimestamp: func() *time.Time { t := time.Now().Add(-10 * time.Minute); return &t }(),
			},
			{
				StreamID:             uuid.New().String(),
				StreamName:           "Jane Smith",
				AvatarURL:            "avatar2.jpg",
				LastMessageContent:   "How are you?",
				LastMessageTimestamp: func() *time.Time { t := time.Now().Add(-5 * time.Minute); return &t }(),
			},
		}

		mockRepo.EXPECT().GetPrivateStreams(ctx, userUUID).Return(expectedStreams, nil)

		result, err := s.GetPrivateStreams(ctx, &emptypb.Empty{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Streams, 2)
	})

	t.Run("no_userUUID", func(t *testing.T) {
		badCtx := context.WithValue(context.Background(), config.KeyLogger, mockLogger)

		mockLogger.EXPECT().AddFuncName("GetPrivateStreams")
		mockLogger.EXPECT().Error("failed to get requester id")

		_, err := s.GetPrivateStreams(badCtx, &emptypb.Empty{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get requester id")
	})

	t.Run("empty_streams", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetPrivateStreams")

		emptyStreams := &model.PrivateStreamPreviewList{}
		mockRepo.EXPECT().GetPrivateStreams(ctx, userUUID).Return(emptyStreams, nil)

		result, err := s.GetPrivateStreams(ctx, &emptypb.Empty{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Streams, 0)
	})

	t.Run("DB_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetPrivateStreams")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().GetPrivateStreams(ctx, userUUID).
			Return(nil, fmt.Errorf("database connection failed"))

		_, err := s.GetPrivateStreams(ctx, &emptypb.Empty{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get private streams")
	})
}
