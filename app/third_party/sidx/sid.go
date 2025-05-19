package sidx

import (
	"time"

	"github.com/sony/sonyflake"
)

type Sid struct {
	sf *sonyflake.Sonyflake
}

func NewSid() *Sid {
	st := sonyflake.Settings{
		StartTime: time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC),
	}
	sf := sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("sonyflake not created")
	}
	return &Sid{sf}
}

func (s Sid) GenUint64() int64 {
	// 生成分布式ID
	// 这里不处理err，源码中只有 sf.elapsedTime >= 1<<BitLenTime 时才会返回错误
	// 这个判断逻辑是说初始化时设定的 StartTime，运行超过174年左右后会报错
	id, _ := s.sf.NextID()

	// 162942888881287		  当前生成id
	// 9223372036854775807    int64最大值

	return int64(id)
}
