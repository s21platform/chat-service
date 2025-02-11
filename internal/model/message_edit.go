package model

import "github.com/google/uuid"

type EditMessageRequest struct {
	MessageID uuid.UUID `db:"id"`      // uuid сообщения
	Content   string    `db:"content"` // новый текст сообщения
}

type EditedMessage struct {
	MessageID uuid.UUID `db:"id"`      // uuid измененного сообщения
	Content   string    `db:"content"` // новый текст сообщения
}
