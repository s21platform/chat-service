package model

import (
	"time"
)

const (
	PrivateStreamType = "private"

	TextMessageType = "text"
)

type PrivateStreamPreviewList []PrivateStreamPreview

type PrivateStreamPreview struct {
	StreamID             string     `db:"stream_id"`
	LastMessageContent   *string    `db:"last_message_content"`
	StreamName           string     `db:"stream_name"`
	AvatarURL            string     `db:"avatar_url"`
	LastMessageTimestamp *time.Time `db:"last_message_timestamp"`
}
