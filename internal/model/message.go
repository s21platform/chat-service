package model

import (
	"time"

	"github.com/google/uuid"

	chat_proto "github.com/s21platform/chat-service/pkg/chat"
)

type MessageList []Message

type Message struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	StreamID  uuid.UUID  `db:"stream_id" json:"stream_id"`
	SenderID  uuid.UUID  `db:"sender_id" json:"sender_id"`
	Type      string     `db:"type" json:"type"`
	Content   string     `db:"content" json:"content"`
	RootID    *uuid.UUID `db:"root_id" json:"root_id,omitempty"`
	ParentID  *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`
	SentAt    time.Time  `db:"sent_at" json:"sent_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

func (m *MessageList) FromDTO() []*chat_proto.Message {
	result := make([]*chat_proto.Message, 0, len(*m))

	for _, message := range *m {
		msg := &chat_proto.Message{
			Uuid:    message.ID.String(),
			Content: message.Content,
			SentAt:  message.SentAt.Format(time.RFC3339),
		}

		if message.UpdatedAt != nil {
			msg.UpdatedAt = message.UpdatedAt.Format(time.RFC3339)
		}

		if message.RootID != nil {
			msg.RootUuid = message.RootID.String()
		}

		if message.ParentID != nil {
			msg.ParentUuid = message.ParentID.String()
		}

		result = append(result, msg)
	}

	return result
}
