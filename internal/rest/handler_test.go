package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	api "github.com/s21platform/chat-service/internal/generated"
	"github.com/s21platform/chat-service/internal/model"
	"github.com/s21platform/chat-service/internal/pkg/tx"
)

func createTxContext(ctx context.Context, mockRepo *MockDBRepo) context.Context {
	return context.WithValue(ctx, tx.KeyTx, tx.Tx{DbRepo: mockRepo})
}

func TestHandler_CreateStream(t *testing.T) {
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

		handler := New(mockRepo, mockUserClient, nil, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("CreateStream")
		mockValidator.EXPECT().ValidateCreateStream(gomock.Any(), creatorUUID).Return(nil)

		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()

		mockUserClient.EXPECT().GetUserInfoByUUID(gomock.Any(), creatorUUID).
			Return(&model.StreamMemberParams{
				UserID:    creatorUUID,
				Nickname:  "test_creator",
				AvatarURL: "test_avatar",
			}, nil)

		mockUserClient.EXPECT().GetUserInfoByUUID(gomock.Any(), companionUUID).
			Return(&model.StreamMemberParams{
				UserID:    companionUUID,
				Nickname:  "test_companion",
				AvatarURL: "test_avatar",
			}, nil)

		mockRepo.EXPECT().AddNewUser(gomock.Any(), gomock.Any()).Return(nil).Times(2)
		mockRepo.EXPECT().CreateStream(gomock.Any(), "private", gomock.Any(), creatorUUID).Return("test-stream-id", nil)
		mockRepo.EXPECT().AddStreamMembers(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockRepo.EXPECT().AddUserSubscriptions(gomock.Any(), gomock.Any()).Return(nil)

		requestBody := api.CreateStreamRequest{
			Users: []api.ChatUser{
				{Id: companionUUID, Metadata: stringPtr("metadata1")},
			},
			Type:            "private",
			ChatMetadata:    "chat metadata",
			CreatorMetadata: "creator metadata",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/chat/streams", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, creatorUUID)
		reqCtx = createTxContext(reqCtx, mockRepo)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.CreateStream(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.CreateStreamResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "test-stream-id", response.Id)
	})

	t.Run("invalid_json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := New(mockRepo, mockUserClient, nil, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("CreateStream")
		mockLogger.EXPECT().Error(gomock.Any())

		req := httptest.NewRequest(http.MethodPost, "/api/chat/streams", strings.NewReader("invalid json"))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, creatorUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.CreateStream(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Error, "invalid request body")
	})
}

func TestHandler_SendMessage(t *testing.T) {
	t.Parallel()

	senderUUID := uuid.New().String()
	streamID := uuid.New().String()

	t.Run("success_simple", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockCentrifuge := NewMockCetrifugeClient(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := New(mockRepo, mockUserClient, mockCentrifuge, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("SendMessage")
		mockValidator.EXPECT().ValidateSendMessage(gomock.Any()).Return(nil)

		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()

		mockRepo.EXPECT().IsStreamMember(gomock.Any(), streamID, senderUUID).Return(true, nil)
		mockRepo.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Return(nil)
		mockCentrifuge.EXPECT().Publish(gomock.Any(), streamID, gomock.Any()).Return(nil)

		requestBody := api.SendMessageRequest{
			Content:     "Hello world",
			MessageType: "text",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/chat/streams/%s/messages", streamID), bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, senderUUID)
		reqCtx = createTxContext(reqCtx, mockRepo)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("stream_id", streamID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.SendMessage(w, req, streamID)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.SendMessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.MessageId)
		assert.NotEmpty(t, response.SentAt)
	})

	t.Run("no_senderID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockUserClient := NewMockUserClient(ctrl)
		mockValidator := NewMockValidator(ctrl)
		mockCentrifuge := NewMockCetrifugeClient(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := New(mockRepo, mockUserClient, mockCentrifuge, mockValidator, nil)

		mockLogger.EXPECT().AddFuncName("SendMessage")
		mockLogger.EXPECT().Error("failed to get sender ID")

		requestBody := api.SendMessageRequest{
			Content:     "Hello",
			MessageType: "text",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/chat/streams/%s/messages", streamID), bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("stream_id", streamID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.SendMessage(w, req, streamID)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Error, "failed to get sender ID")
	})
}

func TestHandler_GetPrivateStreams(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)
	mockValidator := NewMockValidator(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()

	handler := New(mockRepo, mockUserClient, nil, mockValidator, nil)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetPrivateStreams")

		ptr := "Hello there!"

		expectedStreams := &model.PrivateStreamPreviewList{
			{
				StreamID:             uuid.New().String(),
				StreamName:           "John Doe",
				AvatarURL:            "avatar.jpg",
				LastMessageContent:   &ptr,
				LastMessageTimestamp: func() *time.Time { t := time.Now().Add(-10 * time.Minute); return &t }(),
			},
		}

		mockRepo.EXPECT().GetPrivateStreams(gomock.Any(), userUUID).Return(expectedStreams, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/chat/streams/private", nil)

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.GetPrivateStreams(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.GetPrivateStreamsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.Streams, 1)
	})
}

func TestHandler_GetStreamRecentMessages(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockUserClient := NewMockUserClient(ctrl)
	mockValidator := NewMockValidator(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	streamID := uuid.New().String()

	handler := New(mockRepo, mockUserClient, nil, mockValidator, nil)

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
		}

		mockRepo.EXPECT().IsStreamMember(gomock.Any(), streamID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetStreamRecentMessages(gomock.Any(), streamID, "", int32(20)).Return(expectedMessages, nil)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/chat/streams/%s/messages", streamID), nil)

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("stream_id", streamID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.GetStreamRecentMessages(w, req, streamID, api.GetStreamRecentMessagesParams{})

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.GetStreamRecentMessagesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.Messages, 1)
	})
}

func stringPtr(s string) *string {
	return &s
}
