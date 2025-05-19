package utils

import (
	"gil_teacher/app/consts"
	"time"
)

// 验证Unix时间戳是否为合法的秒数值
func IsValidUnixTimestamp(timestamps ...int64) bool {
	// 检查时间戳是否在合理范围内
	// 2025-01-01 00:00:00 UTC 到 2100-01-01 00:00:00 UTC
	minTimestamp := int64(1735689600) // 2025-01-01 00:00:00 UTC
	maxTimestamp := int64(4102444800) // 2100-01-01 00:00:00 UTC

	for _, timestamp := range timestamps {
		// 检查时间戳是否为正数
		if timestamp <= 0 {
			return false
		}
		// 检查时间戳是否在合理范围内
		if timestamp < minTimestamp || timestamp > maxTimestamp {
			return false
		}
	}
	return true
}

// 日期转时间戳，时区：上海
func DateFirstSecondTimestamp(date string) (int64, error) {
	if date == "" {
		return 0, nil
	}

	time, err := time.Parse(consts.TimeFormatDate, date)
	if err != nil {
		return 0, err
	}
	return time.Unix(), nil
}

// 获取日期的最后一秒时间戳，时区：上海，比如，2025-01-01 返回 2025-01-01 23:59:59 的时间戳
func GetDateLastSecondTimestamp(date string) (int64, error) {
	if date == "" {
		return 0, nil
	}

	t, err := time.Parse(consts.TimeFormatDate, date)
	if err != nil {
		return 0, err
	}
	// 将时间设置为当天的最后一秒
	lastSecond := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, consts.LocationShanghai)
	return lastSecond.Unix(), nil
}
