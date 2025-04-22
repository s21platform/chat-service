package new_nickname

import (
	"context"
	"encoding/json"
	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
	"log"

	"github.com/s21platform/metrics-lib/pkg"
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

	err = h.dbR.UpdateUserNickname(msg.UUID, msg.Nickname)
	if err != nil {
		m.Increment("update_nickname.error")
		log.Printf("failed to update nickname: %v", err)
		return err
	}

	log.Printf("Successfully updated nickname for user %s", msg.UUID)
	m.Increment("update_nickname.success")

	return nil
}
