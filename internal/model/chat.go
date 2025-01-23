package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Uuid    uuid.UUID `db:"sender_uuid"` // uuid пользователя
	Content string    `db:"content"`     // само сообщение
	SentAt  time.Time `db:"created_at"`  // время отправки
}

type MessageData struct {
	Messages []Message
}

type EditMessageRequest struct {
	MessageUUID uuid.UUID `db:"id"`      // uuid сообщения
	Content     string    `db:"content"` // новый текст сообщения
}

type EditedMessage struct {
	MessageUUID uuid.UUID `db:"id"`      // uuid измененного сообщения
	Content     string    `db:"content"` // новый текст сообщения
}
