package time

import (
	"strings"
	"time"
)

// ParseDuration 将简单时间字符串转换为 time.Duration
// 支持的格式：
// s: 秒 (例如 "3s")
// m: 分钟 (例如 "3m")
// h: 小时 (例如 "3h")
func ParseDuration(s string) time.Duration {
	s = strings.ReplaceAll(s, "秒", "s")
	s = strings.ReplaceAll(s, "分钟", "m")
	s = strings.ReplaceAll(s, "小时", "h")
	t, err := time.ParseDuration(s)
	if err != nil {
		return time.Duration(0)
	}
	return t
}
