package timex

import (
	"github.com/golang-module/carbon"
	"time"
)

const (
	Layout = "2006-01-02 15:04:05"
)

// DateTimeToShow 时间转成对应的日期文本格式
func DateTimeToShow(t time.Time) string {
	// 当天的，显示具体时间
	strTime := ""
	newCarbon := carbon.Time2Carbon(t)
	if newCarbon.IsToday() {
		strTime = newCarbon.ToTimeString()
		// 处理秒
		strTime = strTime[0 : len(strTime)-3]
	} else if newCarbon.IsYesterday() {
		strTime = "昨天"
	} else {
		strTime = newCarbon.ToDateString()
	}

	return strTime
}

func TimeFormat(t time.Time, layout string) string {
	if layout == "" {
		layout = DefaultLayout
	}
	return t.Format(layout)
}
