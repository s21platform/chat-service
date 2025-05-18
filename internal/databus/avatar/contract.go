package avatar

import "context"

type DBRepo interface {
	UpdateUserAvatar(ctx context.Context, userUUID, avatarLink string) error
}
