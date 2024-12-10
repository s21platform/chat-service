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
