package model

import (
	"time"

	"github.com/google/uuid"
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
