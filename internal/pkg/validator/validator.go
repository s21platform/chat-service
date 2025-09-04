package validator

import (
	"fmt"
	"strings"

	api "github.com/s21platform/chat-service/internal/generated"
	"github.com/s21platform/chat-service/internal/model"
)

type Validator struct{}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateCreateStream(req *api.CreateStreamRequest, creatorID string) error {
	if strings.TrimSpace(req.Type) == "" {
		return fmt.Errorf("stream type is required")
	}

	uniqueUsers := make(map[string]struct{})
	for _, user := range req.Users {
		if strings.TrimSpace(user.Id) != "" && user.Id != creatorID {
			uniqueUsers[user.Id] = struct{}{}
		}
	}

	totalParticipants := len(uniqueUsers) + 1

	switch req.Type {
	case model.PrivateStreamType:
		if totalParticipants != 2 {
			return fmt.Errorf("private stream requires exactly 2 participants, got %d", totalParticipants)
		}
	default:
		return fmt.Errorf("stream type '%s' is not supported", req.Type)
	}

	return nil
}

func (v *Validator) ValidateSendMessage(req *api.SendMessageRequest) error {
	if strings.TrimSpace(req.Content) == "" {
		return fmt.Errorf("content cannot be empty")
	}

	if strings.TrimSpace(req.MessageType) == "" {
		return fmt.Errorf("message_type is required")
	}

	if len([]rune(req.Content)) > 500 {
		return fmt.Errorf("content exceeds maximum length of 500 characters")
	}

	if req.MessageType != model.TextMessageType {
		return fmt.Errorf("message type '%s' is not supported yet", req.MessageType)
	}

	return nil
}
