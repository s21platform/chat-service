package avatar

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/s21platform/avatar-service/pkg/avatar"
	logger_lib "github.com/s21platform/logger-lib"
	"github.com/s21platform/metrics-lib/pkg"

	"github.com/s21platform/chat-service/internal/config"
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
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("Handler")

	m := pkg.FromContext(ctx, config.KeyMetrics)

	var msg avatar.NewAvatarRegister
	err := convertMessage(in, &msg)
	if err != nil {
		m.Increment("update_avatar.error")
		logger.Error(fmt.Sprintf("failed to convert message: %v", err))
		return err
	}

	err = h.dbR.UpdateUserAvatar(ctx, msg.Uuid, msg.Link)
	if err != nil {
		m.Increment("update_avatar.error")
		logger.Error(fmt.Sprintf("failed to update avatar: %v", err))
		return err
	}

	m.Increment("update_avatar.success")

	return nil
}
