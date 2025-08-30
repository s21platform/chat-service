package tx

import "context"

type DBRepo interface {
	WithTx(ctx context.Context, cb func(ctx context.Context) error) (err error)
}
