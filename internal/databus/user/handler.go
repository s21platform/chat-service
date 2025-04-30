package user

import (
	"context"
	"encoding/json"
	"log"

	"github.com/s21platform/metrics-lib/pkg"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
)

type Handler struct {
	dbR DBRepo
}

func New(dbR DBRepo) *Handler {
	return &Handler{dbR: dbR}
}

func convertMessage(bMessage []byte, target interface{}) error {
	err := json.Unmarshal(bMessage, target)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Handler(ctx context.Context, in []byte) error {
	m := pkg.FromContext(ctx, config.KeyMetrics)
	var msg model.UpdateNicknameMessage

	log.Printf("Received message: %s", string(in))

	err := convertMessage(in, &msg)
	if err != nil {
		m.Increment("update_nickname.error")
		log.Printf("failed to convert message: %v", err)
		return err
	}

	log.Printf("Parsed message: UUID=%s, Nickname=%s", msg.UUID, msg.Nickname)

	err = h.dbR.UpdateUserNickname(ctx, msg.UUID, msg.Nickname)
	if err != nil {
		m.Increment("update_nickname.error")
		log.Printf("failed to update nickname: %v", err)
		return err
	}

	log.Printf("Successfully updated nickname for user %s", msg.UUID)
	m.Increment("update_nickname.success")

	return nil
}
