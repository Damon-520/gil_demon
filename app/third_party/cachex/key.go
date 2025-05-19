package cachex

import (
	"fmt"
	"time"
)

const (
	// 用户消息缓存
	MessageChannel        CacheKey = "message-channel"
	UserMessageChannelKey CacheKey = "user-message:uid_%d:cid_inf_%d" //用户渠道缓存 用户ID:频道ID
	PlatformMessage       CacheKey = "platform-message:id_%d"         //平台消息缓存 平台ID
)

var CacheKeyTTLs = map[CacheKey]time.Duration{
	MessageChannel:  time.Minute * 10, //10分钟
	PlatformMessage: time.Hour * 24,   //24小时
}

type CacheKey string

func (c CacheKey) Sprintf(v ...interface{}) string {
	return fmt.Sprintf(c.String(), v...)
}

func (c CacheKey) String() string {
	return string(c)
}

func (c CacheKey) TTL() time.Duration {
	return CacheKeyTTLs[c]
}
