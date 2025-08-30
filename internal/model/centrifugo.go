package model

import "github.com/golang-jwt/jwt/v5"

type CentrifugoEvent struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type CentrifugoEventParams struct {
	Channel string  `json:"channel"`
	Data    Message `json:"data"`
}

type CentrifugoConnectClaims struct {
	jwt.RegisteredClaims
}

type CentrifugoSubscribeClaims struct {
	jwt.RegisteredClaims

	// Centrifugo специфичные поля
	Channel string `json:"channel"`
	Client  string `json:"client,omitempty"`

	// Кастомные поля для безопасности
	UserID   string `json:"user_id"`
	StreamID string `json:"stream_id"`
}
