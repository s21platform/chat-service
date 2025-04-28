package user

import "context"

type DBRepo interface {
	UpdateUserNickname(ctx context.Context, userUUID, newNickname string) error
}
