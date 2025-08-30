package validator

import (
	"fmt"
	"strings"

	"github.com/s21platform/chat-service/internal/model"
	"github.com/s21platform/chat-service/pkg/chat"
)

type Validator struct{}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateStreamByType(in *chat.CreateStreamIn, creatorID string) error {
	if strings.TrimSpace(in.Type) == "" {
		return fmt.Errorf("stream type is required")
	}

	uniqueUsers := make(map[string]struct{})
	for _, user := range in.Users {
		if strings.TrimSpace(user.Id) != "" && user.Id != creatorID {
			uniqueUsers[user.Id] = struct{}{}
		}
	}

	totalParticipants := len(uniqueUsers) + 1

	switch in.Type {
	case model.PrivateStreamType:
		if totalParticipants != 2 {
			return fmt.Errorf("private stream requires exactly 2 participants, got %d", totalParticipants)
		}
	default:
		return fmt.Errorf("stream type '%s' is not supported", in.Type)
	}

	return nil
}

func (v *Validator) ValidateSendMessage(in *chat.SendMessageIn) error {
	if strings.TrimSpace(in.StreamId) == "" {
		return fmt.Errorf("stream_id is required")
	}

	if strings.TrimSpace(in.Content) == "" {
		return fmt.Errorf("content cannot be empty")
	}

	if strings.TrimSpace(in.MessageType) == "" {
		return fmt.Errorf("message_type is required")
	}

	if len([]rune(in.Content)) > 500 {
		return fmt.Errorf("content exceeds maximum length of 500 characters")
	}

	if in.MessageType != model.TextMessageType {
		return fmt.Errorf("message type '%s' is not supported yet", in.MessageType)
	}

	return nil
}
