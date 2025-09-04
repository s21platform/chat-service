package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/chat-service/internal/config"
	api "github.com/s21platform/chat-service/internal/generated"
	"github.com/s21platform/chat-service/internal/model"
	"github.com/s21platform/chat-service/internal/pkg/tx"
)

type Handler struct {
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
) *Handler {
	return &Handler{
		repository:       repo,
		userClient:       userClient,
		centrifugeClient: centrifugeClient,
		validator:        validator,
		jwtGenerator:     jwtGenerator,
	}
}

func (h *Handler) CreateStream(w http.ResponseWriter, r *http.Request) {
	logger := logger_lib.FromContext(r.Context(), config.KeyLogger)
	logger.AddFuncName("CreateStream")

	var req api.CreateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(fmt.Sprintf("failed to decode request: %v", err))
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	creatorID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get creator ID")
		h.writeError(w, "failed to get creator ID", http.StatusInternalServerError)
		return
	}

	if err := h.validator.ValidateCreateStream(&req, creatorID); err != nil {
		logger.Error(fmt.Sprintf("stream validation failed: %v", err))
		h.writeError(w, fmt.Sprintf("stream validation failed: %v", err), http.StatusBadRequest)
		return
	}

	var streamID string
	err := tx.TxExecute(r.Context(), func(ctx context.Context) error {
		allUserIDs := []string{creatorID}
		for _, user := range req.Users {
			if user.Id != "" && user.Id != creatorID {
				allUserIDs = append(allUserIDs, user.Id)
			}
		}

		for _, userID := range allUserIDs {
			userInfo, err := h.userClient.GetUserInfoByUUID(ctx, userID)
			if err != nil {
				logger.Error(fmt.Sprintf("failed to get user info for %s: %v", userID, err))
				return fmt.Errorf("failed to get user info for %s: %v", userID, err)
			}

			err = h.repository.AddNewUser(ctx, userInfo)
			if err != nil {
				logger.Error(fmt.Sprintf("failed to add user %s to users table: %v", userID, err))
				return fmt.Errorf("failed to add user %s to users table: %v", userID, err)
			}
		}

		var err error
		streamID, err = h.repository.CreateStream(ctx, req.Type, req.ChatMetadata, creatorID)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to create stream: %v", err))
			return err
		}

		var members []model.StreamMember
		members = append(members, model.StreamMember{
			UserID:   creatorID,
			Metadata: req.CreatorMetadata,
		})

		for _, user := range req.Users {
			if user.Id != "" && user.Id != creatorID {
				metadata := ""
				if user.Metadata != nil {
					metadata = *user.Metadata
				}
				members = append(members, model.StreamMember{
					UserID:   user.Id,
					Metadata: metadata,
				})
			}
		}

		err = h.repository.AddStreamMembers(ctx, streamID, members)
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

		err = h.repository.AddUserSubscriptions(ctx, subscriptions)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to create subscriptions: %v", err))
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error(fmt.Sprintf("failed to complete stream creation transaction: %v", err))
		h.writeError(w, fmt.Sprintf("failed to create stream: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.CreateStreamResponse{
		Id: streamID,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) GetPrivateStreams(w http.ResponseWriter, r *http.Request) {
	logger := logger_lib.FromContext(r.Context(), config.KeyLogger)
	logger.AddFuncName("GetPrivateStreams")

	requesterID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get requester id")
		h.writeError(w, "failed to get requester id", http.StatusInternalServerError)
		return
	}

	privateStreams, err := h.repository.GetPrivateStreams(r.Context(), requesterID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get private streams: %v", err))
		h.writeError(w, fmt.Sprintf("failed to get private streams: %v", err), http.StatusInternalServerError)
		return
	}

	streams := make([]api.PrivateStream, len(*privateStreams))
	for i, stream := range *privateStreams {
		var lastMessageTimestamp *string
		if stream.LastMessageTimestamp != nil {
			timestamp := stream.LastMessageTimestamp.Format(time.RFC3339)
			lastMessageTimestamp = &timestamp
		}

		streams[i] = api.PrivateStream{
			StreamId:             stream.StreamID,
			LastMessageContent:   stream.LastMessageContent,
			StreamName:           stream.StreamName,
			AvatarUrl:            &stream.AvatarURL,
			LastMessageTimestamp: lastMessageTimestamp,
		}
	}

	response := api.GetPrivateStreamsResponse{
		Streams: streams,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) GetStreamRecentMessages(w http.ResponseWriter, r *http.Request, streamId string, params api.GetStreamRecentMessagesParams) {
	logger := logger_lib.FromContext(r.Context(), config.KeyLogger)
	logger.AddFuncName("GetStreamRecentMessages")

	userUUID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to find uuid")
		h.writeError(w, "failed to find uuid", http.StatusInternalServerError)
		return
	}

	isMember, err := h.repository.IsStreamMember(r.Context(), streamId, userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to check stream membership: %v", err))
		h.writeError(w, fmt.Sprintf("failed to check stream membership: %v", err), http.StatusInternalServerError)
		return
	}

	if !isMember {
		logger.Error("user is not a member of the stream")
		h.writeError(w, "user is not a member of the stream", http.StatusForbidden)
		return
	}

	offset := ""
	if params.Offset != nil {
		offset = *params.Offset
	}

	limit := int32(20)
	if params.Limit != nil {
		limit = int32(*params.Limit)
	}

	messages, err := h.repository.GetStreamRecentMessages(r.Context(), streamId, offset, limit)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to fetch messages: %v", err))
		h.writeError(w, fmt.Sprintf("failed to fetch messages: %v", err), http.StatusInternalServerError)
		return
	}

	apiMessages := make([]api.Message, len(*messages))
	for i, msg := range *messages {
		var updatedAt *string
		if msg.UpdatedAt != nil {
			timestamp := msg.UpdatedAt.Format(time.RFC3339)
			updatedAt = &timestamp
		}

		var rootUuid *string
		if msg.RootID != nil {
			uuid := msg.RootID.String()
			rootUuid = &uuid
		}

		var parentUuid *string
		if msg.ParentID != nil {
			uuid := msg.ParentID.String()
			parentUuid = &uuid
		}

		apiMessages[i] = api.Message{
			Uuid:       msg.ID.String(),
			Content:    msg.Content,
			SentAt:     msg.SentAt.Format(time.RFC3339),
			UpdatedAt:  updatedAt,
			RootUuid:   rootUuid,
			ParentUuid: parentUuid,
		}
	}

	response := api.GetStreamRecentMessagesResponse{
		Messages: apiMessages,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request, streamId string) {
	logger := logger_lib.FromContext(r.Context(), config.KeyLogger)
	logger.AddFuncName("SendMessage")

	var req api.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(fmt.Sprintf("failed to decode request: %v", err))
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	senderID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get sender ID")
		h.writeError(w, "failed to get sender ID", http.StatusInternalServerError)
		return
	}

	if err := h.validator.ValidateSendMessage(&req); err != nil {
		logger.Error(fmt.Sprintf("message validation failed: %v", err))
		h.writeError(w, fmt.Sprintf("message validation failed: %v", err), http.StatusBadRequest)
		return
	}

	var message model.Message
	err := tx.TxExecute(r.Context(), func(ctx context.Context) error {
		isMember, err := h.repository.IsStreamMember(ctx, streamId, senderID)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to check stream membership: %v", err))
			return fmt.Errorf("failed to check stream membership: %v", err)
		}

		if !isMember {
			logger.Error(fmt.Sprintf("user %s is not a member of stream %s", senderID, streamId))
			return fmt.Errorf("user is not a member of this stream")
		}

		message = model.Message{
			ID:       uuid.New(),
			StreamID: uuid.MustParse(streamId),
			SenderID: uuid.MustParse(senderID),
			Type:     req.MessageType,
			Content:  req.Content,
			SentAt:   time.Now(),
		}

		if req.ParentId != nil && *req.ParentId != "" {
			parentUUID := uuid.MustParse(*req.ParentId)
			message.ParentID = &parentUUID
		}

		if req.RootId != nil && *req.RootId != "" {
			rootUUID := uuid.MustParse(*req.RootId)
			message.RootID = &rootUUID
		}

		err = h.repository.SaveMessage(ctx, &message)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to save message: %v", err))
			return fmt.Errorf("failed to save message: %v", err)
		}

		return nil
	})

	if err != nil {
		logger.Error(fmt.Sprintf("failed to send message transaction: %v", err))
		h.writeError(w, fmt.Sprintf("failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	channel := message.StreamID.String()
	err = h.centrifugeClient.Publish(r.Context(), channel, message)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to publish message to stream: %v", err))
	}

	response := api.SendMessageResponse{
		MessageId: message.ID.String(),
		SentAt:    message.SentAt.Format(time.RFC3339),
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) GetConnectAccessToken(w http.ResponseWriter, r *http.Request) {
	logger := logger_lib.FromContext(r.Context(), config.KeyLogger)
	logger.AddFuncName("GetConnectAccessToken")

	userUUID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get user UUID")
		h.writeError(w, "failed to get user UUID", http.StatusInternalServerError)
		return
	}

	token, expiresAt, err := h.jwtGenerator.GenerateConnectToken(userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to generate access token: %v", err))
		h.writeError(w, fmt.Sprintf("failed to generate access token: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Info(fmt.Sprintf("generated access token for user %s", userUUID))

	response := api.GetConnectAccessTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) GetStreamSubscribeToken(w http.ResponseWriter, r *http.Request, streamId string) {
	logger := logger_lib.FromContext(r.Context(), config.KeyLogger)
	logger.AddFuncName("GetStreamSubscribeToken")

	userUUID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get user UUID")
		h.writeError(w, "failed to get user UUID", http.StatusInternalServerError)
		return
	}

	isMember, err := h.repository.IsStreamMember(r.Context(), streamId, userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to check stream membership: %v", err))
		h.writeError(w, fmt.Sprintf("failed to check stream membership: %v", err), http.StatusInternalServerError)
		return
	}

	if !isMember {
		logger.Error("user is not a member of the stream")
		h.writeError(w, "user is not a member of the stream", http.StatusForbidden)
		return
	}

	token, expiresAt, err := h.jwtGenerator.GenerateSubscribeToken(userUUID, streamId)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to generate subscribe token: %v", err))
		h.writeError(w, fmt.Sprintf("failed to generate subscribe token: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Info(fmt.Sprintf("generated subscribe token for user %s, stream %s", userUUID, streamId))

	response := api.GetStreamSubscribeTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Channel:   streamId,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) GetUserActiveStreams(w http.ResponseWriter, r *http.Request) {
	logger := logger_lib.FromContext(r.Context(), config.KeyLogger)
	logger.AddFuncName("GetUserActiveStreams")

	userUUID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get user UUID")
		h.writeError(w, "failed to get user UUID", http.StatusInternalServerError)
		return
	}

	streamIDs, err := h.repository.GetUserActiveStreams(r.Context(), userUUID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get user active streams: %v", err))
		h.writeError(w, fmt.Sprintf("failed to get user active streams: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.GetUserActiveStreamsResponse{
		StreamIds: streamIDs,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) GetBatchSubscribeTokens(w http.ResponseWriter, r *http.Request) {
	logger := logger_lib.FromContext(r.Context(), config.KeyLogger)
	logger.AddFuncName("GetBatchSubscribeTokens")

	var req api.GetBatchSubscribeTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(fmt.Sprintf("failed to decode request: %v", err))
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userUUID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to get user UUID")
		h.writeError(w, "failed to get user UUID", http.StatusInternalServerError)
		return
	}

	var subscriptions []api.StreamSubscription

	for _, streamID := range req.StreamIds {
		isMember, err := h.repository.IsStreamMember(r.Context(), streamID, userUUID)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to check stream membership for %s: %v", streamID, err))
			continue
		}

		if !isMember {
			logger.Warn(fmt.Sprintf("user %s is not a member of stream %s, skipping", userUUID, streamID))
			continue
		}

		token, expiresAt, err := h.jwtGenerator.GenerateSubscribeToken(userUUID, streamID)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to generate subscribe token for stream %s: %v", streamID, err))
			continue
		}

		subscriptions = append(subscriptions, api.StreamSubscription{
			StreamId:  streamID,
			Token:     token,
			ExpiresAt: expiresAt,
			Channel:   streamID,
		})
	}

	response := api.GetBatchSubscribeTokensResponse{
		Subscriptions: subscriptions,
	}

	h.writeJSON(w, response, http.StatusOK)
}

// ----------------------------- helpers -----------------------------

func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(api.Error{Error: message})
}
