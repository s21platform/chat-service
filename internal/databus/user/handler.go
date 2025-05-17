package user

import (
	"context"
	"encoding/json"
	"fmt"

	logger_lib "github.com/s21platform/logger-lib"
	"github.com/s21platform/metrics-lib/pkg"
	"github.com/s21platform/user-service/pkg/user"

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

	var msg user.UserNicknameUpdated
	err := convertMessage(in, &msg)
	if err != nil {
		m.Increment("update_nickname.error")
		logger.Error(fmt.Sprintf("failed to convert message: %v", err))
		return err
	}

	err = h.dbR.UpdateUserNickname(ctx, msg.UserUuid, msg.Nickname)
	if err != nil {
		m.Increment("update_nickname.error")
		logger.Error(fmt.Sprintf("failed to update nickname: %v", err))
		return err
	}

	m.Increment("update_nickname.success")

	return nil
}
