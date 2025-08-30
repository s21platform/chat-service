package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/s21platform/chat-service/internal/model"
)

type Generator struct {
	secret []byte
}

func New(secret string) *Generator {
	return &Generator{
		secret: []byte(secret),
	}
}

func (g *Generator) GenerateConnectToken(userID string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	claims := model.CentrifugoConnectClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(g.secret)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign connect JWT token: %w", err)
	}

	return tokenString, expiresAt.Unix(), nil
}

func (g *Generator) GenerateSubscribeToken(userID, streamID string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	claims := model.CentrifugoSubscribeClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Channel:  streamID,
		UserID:   userID,
		StreamID: streamID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(g.secret)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign subscribe JWT token: %w", err)
	}

	return tokenString, expiresAt.Unix(), nil
}

func (g *Generator) ValidateConnectToken(tokenString string) (*model.CentrifugoConnectClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.CentrifugoConnectClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse connect JWT token: %w", err)
	}

	if claims, ok := token.Claims.(*model.CentrifugoConnectClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid connect JWT token")
}

func (g *Generator) ValidateSubscribeToken(tokenString string) (*model.CentrifugoSubscribeClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.CentrifugoSubscribeClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse subscribe JWT token: %w", err)
	}

	if claims, ok := token.Claims.(*model.CentrifugoSubscribeClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid subscribe JWT token")
}
