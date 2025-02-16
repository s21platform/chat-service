package model

import (
	"time"

	chat_proto "github.com/s21platform/chat-proto/chat-proto"
)

type ChatInfoList []ChatInfo

type ChatInfo struct {
	LastMessage          string     `db:"content"`
	ChatName             string     `db:"chat_name"`
	AvatarURL            string     `db:"avatar_link"`
	LastMessageTimestamp *time.Time `db:"created_at"`
	ChatUUID             string     `db:"uuid"`
}

func (c *ChatInfoList) FromDTO() []*chat_proto.Chat {
	result := make([]*chat_proto.Chat, 0, len(*c))

	for _, chat := range *c {
		result = append(result, &chat_proto.Chat{
			LastMessage:          chat.LastMessage,
			ChatName:             chat.ChatName,
			AvatarUrl:            chat.AvatarURL,
			LastMessageTimestamp: chat.convertTimestamp(),
			ChatUuid:             chat.ChatUUID,
		})
	}

	return result
}

func (c *ChatInfo) convertTimestamp() string {
	if c.LastMessageTimestamp == nil {
		return ""
	}

	return c.LastMessageTimestamp.Format(time.RFC3339)
}
