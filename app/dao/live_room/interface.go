package live_room

import (
	"context"

	"gil_teacher/app/third_party/gorm_builder"
)

type ILiveRoomDao interface {
	Info(ctx context.Context, options gorm_builder.Options) (info LiveRoom, err error)
	List(ctx context.Context, options gorm_builder.Options) (list []LiveRoom, total int64, err error)
	Count(ctx context.Context, options gorm_builder.Options) (total int, err error)
	Save(ctx context.Context, data LiveRoom) (rows int, err error)
	Update(ctx context.Context, options gorm_builder.Options, data map[string]interface{}) (int64, error)
}
