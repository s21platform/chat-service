package model

import "github.com/google/uuid"

type CreateChatRequest struct {
	CompanionID uuid.UUID `json:"companion_uuid"`
}

type CreateChatResponse struct {
	NewChatUUID uuid.UUID `json:"new_chat_uuid"`
}
