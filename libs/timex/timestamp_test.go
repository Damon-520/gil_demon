package timex

import (
	"fmt"
	"github.com/golang-module/carbon"
	"testing"
	"time"
)

func Test_DataFormatToStr(t *testing.T) {
	fmt.Println("format", carbon.Parse("2022-08-31 15:11:33 +0800 CST").ToDateTimeString())
	ts := time.Now()
	fmt.Println("ts", ts)
	fmt.Println("today", carbon.Time2Carbon(ts).IsToday())

	//fmt.Println("今天：", DateTimeToShow("2022-09-01 11:06:07"))
	//fmt.Println("昨天：", DateTimeToShow("2022-08-31 11:06:07"))
	//fmt.Println("后天：", DateTimeToShow("2022-08-20 11:06:07"))
}
