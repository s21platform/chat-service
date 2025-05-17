package model

import (
	"time"

	"github.com/google/uuid"

	chat_proto "github.com/s21platform/chat-service/pkg/chat"
)

type Message struct {
	Uuid       uuid.UUID `db:"sender_uuid"` // uuid пользователя
	Content    string    `db:"content"`     // само сообщение
	SentAt     time.Time `db:"sent_at"`     // время отправки
	UpdatedAt  time.Time `db:"updated_at"`  // время обновления
	RootUUID   uuid.UUID `db:"root_uuid"`   // uuid корневого сообщения
	ParentUUID uuid.UUID `db:"parent_uuid"` // uuid сообщения, на которое идет прямой ответ
}

type MessageList []Message

func (m *MessageList) FromDTO() []*chat_proto.Message {
	result := make([]*chat_proto.Message, 0, len(*m))

	for _, message := range *m {
		result = append(result, &chat_proto.Message{
			Uuid:       message.Uuid.String(),
			Content:    message.Content,
			SentAt:     message.SentAt.Format(time.RFC3339),
			UpdatedAt:  message.UpdatedAt.Format(time.RFC3339),
			RootUuid:   message.RootUUID.String(),
			ParentUuid: message.ParentUUID.String(),
		})
	}

	return result
}
