package time

import (
	"fmt"
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

// IsDateInRange 判断日期字符串是否在时间范围内（包含边界）
func IsDateInRange(targetStr string, start, end time.Time) (bool, error) {
	// 解析日期字符串（假设格式为 "2006-01-02"）
	layout := "2006-01-02"
	loc := start.Location() // 使用 start 的时区
	target, err := time.ParseInLocation(layout, targetStr, loc)
	if err != nil {
		return false, fmt.Errorf("解析日期失败: %v", err)
	}

	// 将时间转换为当天的零点（仅保留日期部分）
	truncateToDay := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}
	targetDate := truncateToDay(target)
	startDate := truncateToDay(start)
	endDate := truncateToDay(end)

	// 判断是否在范围内（包含边界）
	return (targetDate.Equal(startDate) || targetDate.After(startDate)) &&
		(targetDate.Equal(endDate) || targetDate.Before(endDate)), nil
}
