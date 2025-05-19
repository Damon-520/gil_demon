package cache

import (
	"context"
)

type DictCache interface {
	Set(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
}
