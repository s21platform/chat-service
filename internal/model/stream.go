package model

import (
	"time"

	"github.com/s21platform/chat-service/pkg/chat"
)

const (
	PrivateStreamType = "private"

	TextMessageType = "text"
)

type PrivateStreamPreviewList []PrivateStreamPreview

type PrivateStreamPreview struct {
	StreamID             string     `db:"stream_id"`
	LastMessageContent   string     `db:"last_message_content"`
	StreamName           string     `db:"stream_name"`
	AvatarURL            string     `db:"avatar_url"`
	LastMessageTimestamp *time.Time `db:"last_message_timestamp"`
}

func (p *PrivateStreamPreviewList) FromDTO() []*chat.PrivateStream {
	result := make([]*chat.PrivateStream, 0, len(*p))

	for _, stream := range *p {
		result = append(result, &chat.PrivateStream{
			StreamId:             stream.StreamID,
			LastMessageContent:   stream.LastMessageContent,
			StreamName:           stream.StreamName,
			AvatarUrl:            stream.AvatarURL,
			LastMessageTimestamp: stream.LastMessageTimestamp.Format(time.RFC3339),
		})
	}

	return result
}
