package utils

import (
	"gil_teacher/app/consts"
	"strconv"
	"time"
)

// CalculateWeekDates 计算指定日期所在周的周一到周日的日期
// startDate: 开始日期，格式为 "2006-01-02"
// 返回一个map，key为周几（1-7），value为对应的日期（格式：2006-01-02）
func CalculateWeekDates(startDate string) (map[string]string, error) {
	// 解析开始日期
	startTime, err := time.Parse(consts.TimeFormatDate, startDate)
	if err != nil {
		return nil, err
	}

	// 计算周一到周日的日期
	dates := make(map[string]string)
	for i := 1; i <= 7; i++ {
		// 计算当前是周几
		currentWeekday := int(startTime.Weekday())
		if currentWeekday == 0 {
			currentWeekday = 7
		}

		// 计算需要往前或往后调整的天数
		daysToAdjust := i - currentWeekday
		date := startTime.AddDate(0, 0, daysToAdjust)

		// 格式化日期并存储
		dates[strconv.Itoa(i)] = date.Format(consts.TimeFormatDate)
	}

	return dates, nil
}

// GetWeekday 获取当前是周几（1-7）
func GetWeekday() int {
	weekday := int(time.Now().Weekday())
	if weekday == 0 {
		weekday = 7 // 将周日从0转为7
	}
	return weekday
}

// GetWeekDateRange 获取指定日期所在周的周一和周日日期（格式：YYYY-MM-DD）
func GetWeekDateRange(date time.Time) (string, string) {
	// 计算本周一的日期
	weekday := date.Weekday()
	if weekday == 0 { // 如果是周日，算作上周日，往前调整7天而不是6天
		weekday = 7
	}
	mondayOffset := int(weekday) - 1
	monday := date.AddDate(0, 0, -mondayOffset)

	// 计算本周日的日期
	sunday := monday.AddDate(0, 0, 6)

	// 格式化为YYYY-MM-DD
	return monday.Format(consts.TimeFormatDate), sunday.Format(consts.TimeFormatDate)
}
