package time

import (
	"fmt"
	"time"
)

// GetWeekRange 根据给定时间返回所在周的开始日期和结束日期
// 周一为一周的开始，周日为一周的结束
func GetWeekRange(currentTime string) (string, string, error) {
	// 解析输入的时间字符串
	t, err := time.Parse("2006-01-02", currentTime)
	if err != nil {
		return "", "", err
	}

	// 计算当前是周几 (Go中，time.Weekday()返回0-6，0是周日)
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // 将周日视为7，使周一为1
	}

	// 计算到本周一的偏移天数
	offset := weekday - 1

	// 计算本周一的日期
	weekStart := t.AddDate(0, 0, -offset)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	// 计算本周日的日期
	weekEnd := weekStart.AddDate(0, 0, 6)
	weekEnd = time.Date(weekEnd.Year(), weekEnd.Month(), weekEnd.Day(), 23, 59, 59, 0, weekEnd.Location())

	// 格式化返回
	startStr := weekStart.Format("2006-01-02")
	endStr := weekEnd.Format("2006-01-02")

	return startStr, endStr, nil
}

// GetWeekNumber 计算当前日期是距离开始日期的第几周
// startDate: 开始日期，格式为 "2006-01-02"
// endDate: 当前日期，格式为 "2006-01-02"
// 返回值: 周数（开始日期所在周为第1周）
func GetWeekNumber(startDate string, currentDate string) (int, error) {
	// 解析开始日期
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0, err
	}

	// 解析当前日期
	current, err := time.Parse("2006-01-02", currentDate)
	if err != nil {
		return 0, err
	}

	// 计算开始日期是周几 (0-6，0是周日)
	startWeekday := int(start.Weekday())
	if startWeekday == 0 {
		startWeekday = 7 // 将周日视为7，使周一为1
	}

	// 计算开始日期所在周的周一
	startWeekMonday := start.AddDate(0, 0, -(startWeekday - 1))
	startWeekMonday = time.Date(startWeekMonday.Year(), startWeekMonday.Month(), startWeekMonday.Day(), 0, 0, 0, 0, startWeekMonday.Location())

	// 计算当前日期是周几
	currentWeekday := int(current.Weekday())
	if currentWeekday == 0 {
		currentWeekday = 7 // 将周日视为7，使周一为1
	}

	// 计算当前日期所在周的周一
	currentWeekMonday := current.AddDate(0, 0, -(currentWeekday - 1))
	currentWeekMonday = time.Date(currentWeekMonday.Year(), currentWeekMonday.Month(), currentWeekMonday.Day(), 0, 0, 0, 0, currentWeekMonday.Location())

	// 计算两个周一之间相差的天数
	days := int(currentWeekMonday.Sub(startWeekMonday).Hours() / 24)

	// 计算周数（开始周为第1周）
	weekNumber := days/7 + 1

	return weekNumber, nil
}

// GetCycleWeek 根据周数和周期循环数，计算当前是循环中的第几周
// weekNumber: 当前周数
// cycleWeeks: 循环周数（例如：如果是4周一循环，则cycleWeeks为4）
// 返回值: 循环中的第几周（范围为1到cycleWeeks）
func GetCycleWeek(weekNumber int64, cycleWeeks int64) (int64, error) {

	// 参数校验
	if cycleWeeks <= 0 {
		err := fmt.Errorf("循环周数必须大于0")
		return 0, err
	}

	if weekNumber < 0 {
		err := fmt.Errorf("周数必须大于等于0")
		return 0, err
	}

	// 计算循环中的第几周
	// 例如：如果是4周一循环，第5周实际上是循环中的第1周，第6周是循环中的第2周，以此类推
	cycleWeek := weekNumber%cycleWeeks + 1

	return cycleWeek, nil
}

// GetDayOfWeekName 获取星期几的中文名称
// dayOfWeek: 星期几（1-7，1表示星期一，7表示星期日）
// 返回值: 星期几的中文名称
func GetDayOfWeekName(dayOfWeek int64) string {
	switch dayOfWeek {
	case 1:
		return "一"
	case 2:
		return "二"
	case 3:
		return "三"
	case 4:
		return "四"
	case 5:
		return "五"
	case 6:
		return "六"
	case 7:
		return "日"
	default:
		return fmt.Sprintf("%d", dayOfWeek)
	}
}

// IsWeekStartAndEnd 判断两个日期是否分别是同一周的开始日期（周一）和结束日期（周日）
// startDate: 开始日期，格式为 "2006-01-02"
// endDate: 结束日期，格式为 "2006-01-02"
// 返回值:
//   - bool: 是否是同一周的开始和结束日期
//   - error: 错误信息
func IsWeekStartAndEnd(startDate, endDate string) (bool, error) {
	// 解析开始日期
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return false, fmt.Errorf("无效的开始日期格式: %v", err)
	}

	// 解析结束日期
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return false, fmt.Errorf("无效的结束日期格式: %v", err)
	}

	// 检查开始日期是否为周一
	if start.Weekday() != time.Monday {
		return false, nil
	}

	// 检查结束日期是否为周日
	if end.Weekday() != time.Sunday {
		return false, nil
	}

	// 计算两个日期之间的天数差
	daysDiff := int(end.Sub(start).Hours() / 24)

	// 检查是否相差6天（周一到周日）
	if daysDiff != 6 {
		return false, nil
	}

	// 如果开始日期是周一，结束日期是周日，且相差6天，则它们是同一周的开始和结束日期
	return true, nil
}

// IsMonday 判断给定的日期是否为周一
// date: 日期字符串，支持两种格式："2006-01-02" 或 "2006-01-02"
// 返回值:
//   - bool: 是否为周一
//   - error: 错误信息
func IsMonday(date string) (bool, error) {
	var t time.Time
	var err error

	// 尝试解析日期时间格式
	t, err = time.Parse("2006-01-02", date)
	if err != nil {
		// 如果解析失败，尝试解析日期格式
		t, err = time.Parse("2006-01-02", date)
		if err != nil {
			return false, fmt.Errorf("无效的日期格式: %v", err)
		}
	}

	// 检查是否为周一 (Go中，time.Monday = 1)
	return t.Weekday() == time.Monday, nil
}

// GetDateByWeekday 根据给定的一周开始日期和结束日期，计算特定星期几对应的日期
// weekStart: 周的开始日期，格式为 "2006-01-02"
// weekEnd: 周的结束日期，格式为 "2006-01-02"
// dayOfWeek: 星期几（1-7，1表示星期一，7表示星期日）
// 返回值:
//   - string: 对应星期几的日期，格式为 "2006-01-02"
//   - error: 错误信息
func GetDateByWeekday(weekStart, weekEnd string, dayOfWeek int) (string, error) {
	// 参数校验
	if dayOfWeek < 1 || dayOfWeek > 7 {
		return "", fmt.Errorf("星期几必须是1到7之间的整数，1表示周一，7表示周日")
	}

	// 检查提供的开始和结束日期是否确实是同一周的周一和周日
	isValid, err := IsWeekStartAndEnd(weekStart, weekEnd)
	if err != nil {
		return "", err
	}
	if !isValid {
		return "", fmt.Errorf("提供的日期不是有效的一周开始（周一）和结束（周日）日期")
	}

	// 解析开始日期
	start, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return "", fmt.Errorf("无效的开始日期格式: %v", err)
	}

	// 计算偏移量：dayOfWeek为1表示周一（偏移0天），dayOfWeek为7表示周日（偏移6天）
	offset := dayOfWeek - 1

	// 计算指定星期几的日期
	targetDate := start.AddDate(0, 0, offset)

	// 格式化返回
	return targetDate.Format("2006-01-02"), nil
}
