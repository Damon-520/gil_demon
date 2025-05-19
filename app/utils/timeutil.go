package utils

import (
	"gil_teacher/app/consts"
	"strings"
	"time"
)

// ParseTimeHHMM 解析HH:MM:SS格式的时间字符串
func ParseTimeHHMM(timeStr string) (time.Time, error) {
	// 如果时间字符串不包含秒数，添加:00
	if len(strings.Split(timeStr, ":")) == 2 {
		timeStr = timeStr + ":00"
	}
	return time.ParseInLocation(consts.TimeFormatHHMMSS, timeStr, consts.LocationShanghai)
}

// GetTodayDateTime 获取今天的指定时间
func GetTodayDateTime(t time.Time) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, now.Location())
}

// IsTimeInRange 检查当前时间是否在课程时间范围内（可进入课程）
// 规则：
// 1. 如果课程日期在当前日期之前，返回 true
// 2. 如果是当前日期，且当前时间在课程开始时间之后，返回 true（不考虑结束时间）
// 3. 其他情况返回 false
func IsTimeInRange(courseDate string, startTime, endTime string) (bool, error) {
	// 使用北京时区
	loc := consts.LocationShanghai

	// 解析课程日期
	courseDateTime, err := time.ParseInLocation(consts.TimeFormatDate, courseDate, loc)
	if err != nil {
		return false, err
	}

	// 获取当前北京时间
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	// 如果课程日期在今天之前，返回 true
	if courseDateTime.Before(today) {
		return true, nil
	}

	// 如果课程日期在今天之后，返回 false
	if courseDateTime.After(today) {
		return false, nil
	}

	// 解析课程开始时间
	start, err := ParseTimeHHMM(startTime)
	if err != nil {
		return false, err
	}

	// 将开始时间设置为今天的日期
	startDateTime := time.Date(now.Year(), now.Month(), now.Day(), start.Hour(), start.Minute(), start.Second(), 0, loc)

	// 如果当前时间在课程开始时间之后，可以进入（不考虑结束时间）
	return now.After(startDateTime) || now.Equal(startDateTime), nil
}

// IsInClass 检查当前时间是否在课程时间段内（必须是当天且在课程时间段内）
// 参数：
// - courseDate: 课程日期，格式为 "2006-01-02"
// - currentDate: 当前日期，格式为 "2006-01-02"
// - currentTime: 当前时间，格式为 "15:04:05"
// - startTime: 课程开始时间，格式为 "15:04:05"
// - endTime: 课程结束时间，格式为 "15:04:05"
// 返回：
// - bool: 是否在课程时间段内
func IsInClass(courseDate, currentDate, currentTime, startTime, endTime string) bool {
	return courseDate == currentDate && currentTime >= startTime && currentTime <= endTime
}
