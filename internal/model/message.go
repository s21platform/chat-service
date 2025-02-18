package model

import (
	"time"

	"github.com/google/uuid"

	chat_proto "github.com/s21platform/chat-proto/chat-proto"
)

type Message struct {
	Uuid    uuid.UUID `db:"sender_uuid"` // uuid пользователя
	Content string    `db:"content"`     // само сообщение
	SentAt  time.Time `db:"created_at"`  // время отправки
}

type MessageList []Message

func (m *MessageList) FromDTO() []*chat_proto.Message {
	result := make([]*chat_proto.Message, 0, len(*m))

	for _, message := range *m {
		result = append(result, &chat_proto.Message{
			Uuid:    message.Uuid.String(),
			Content: message.Content,
			SentAt:  message.SentAt.Format(time.RFC3339),
		})
	}

	return result
}
