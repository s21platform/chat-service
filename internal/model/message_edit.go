package model

import (
	"time"

	"github.com/google/uuid"
)

type EditedMessage struct {
	MessageUUID uuid.UUID `db:"uuid"`       // uuid измененного сообщения
	Content     string    `db:"content"`    // новый текст сообщения
	UpdateAt    time.Time `db:"updated_at"` // время обновления сообщения
}
