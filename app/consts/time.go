package consts

import "time"

// 时间常量定义
const (
	// 秒级时间常量
	OneSecond   = 1
	ThreeMinute = 3 * OneSecond
	OneMinute   = 60 * OneSecond
	OneHour     = 60 * OneMinute
	OneDay      = 24 * OneHour
	OneWeek     = 7 * OneDay
	OneMonth    = 30 * OneDay
	OneYear     = 365 * OneDay

	// 默认过期时间（秒）
	ExpireSeconds1Sec   = 1 * OneSecond  // 1秒
	ExpireSeconds30Min  = 30 * OneMinute // 30分钟
	ExpireSeconds2Hour  = 2 * OneHour    // 2小时
	ExpireSeconds24Hour = 24 * OneHour   // 1天
	ExpireSeconds7Day   = 7 * OneDay     // 7天
	ExpireSeconds15Min  = 15 * OneMinute // 15分钟
)

// 时间格式常量
const (
	// 标准时间格式
	TimeFormatSecond = "2006-01-02 15:04:05"
	TimeFormatMinute = "2006-01-02 15:04"
	TimeFormatHour   = "2006-01-02 15"
	TimeFormatDate   = "2006-01-02"
	TimeFormatMonth  = "2006-01"
	TimeFormatYear   = "2006"

	// 带时区的时间格式
	TimeFormatWithZone        = "2006-01-02 15:04:05 -0700"
	TimeFormatRFC3339         = time.RFC3339
	TimeFormatISO8601         = "2006-01-02T15:04:05Z0700"
	TimeFormatWithZoneISO8601 = "2006-01-02 15:04:05 -07:00"
	TimeFormatCompact         = "20060102150405"

	// 时间，不包含日期
	TimeFormatTimeOnly = "15:04:05"

	// 时区
	TimeZoneShanghai = "Asia/Shanghai"
	TimeZoneUTC      = "UTC"

	// TimeFormatHHMMSS 时分秒格式 HH:MM:SS
	TimeFormatHHMMSS = "15:04:05"
)

// 时间转换函数
var (
	// 默认时区位置
	LocationShanghai = time.FixedZone("CST", 8*3600)
	LocationUTC      = time.UTC
)

// 时间间隔常量
var (
	DurationSecond = time.Second
	DurationMinute = time.Minute
	DurationHour   = time.Hour
	DurationDay    = 24 * time.Hour
	DurationWeek   = 7 * 24 * time.Hour
	DurationMonth  = 30 * 24 * time.Hour
	DurationYear   = 365 * 24 * time.Hour
)
