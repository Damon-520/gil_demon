package utils

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gil_teacher/app/third_party/response"

	"github.com/gin-gonic/gin"
)

// ParseInt64 解析字符串为int64类型
// 如果解析失败或值小于等于0，返回0
func ParseInt64(str string) int64 {
	if str == "" {
		return 0
	}
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// ParseInt64WithMin 解析字符串为int64类型，并设置最小值
// 如果解析失败或值小于min，返回min
func ParseInt64WithMin(str string, min int64) int64 {
	val := ParseInt64(str)
	if val < min {
		return min
	}
	return val
}

// ParseInt64WithRange 解析字符串为int64类型，并设置最小值和最大值
// 如果解析失败或值小于min，返回min
// 如果值大于max，返回max
func ParseInt64WithRange(str string, min, max int64) int64 {
	val := ParseInt64(str)
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// ParseTime 解析时间字符串
// 支持RFC3339格式
func ParseTime(str string) (time.Time, error) {
	if str == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, str)
}

// ParseBool 解析字符串为bool类型
// 如果解析失败，返回false
func ParseBool(str string) bool {
	if str == "" {
		return false
	}
	val, err := strconv.ParseBool(str)
	if err != nil {
		return false
	}
	return val
}

// ParseString 解析字符串，如果为空则返回默认值
func ParseString(str, defaultValue string) string {
	if str == "" {
		return defaultValue
	}
	return str
}

// ParseInt 解析字符串为int类型
// 如果解析失败或值小于等于0，返回0
func ParseInt(str string) int {
	if str == "" {
		return 0
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return val
}

// ParseIntWithMin 解析字符串为int类型，并设置最小值
// 如果解析失败或值小于min，返回min
func ParseIntWithMin(str string, min int) int {
	val := ParseInt(str)
	if val < min {
		return min
	}
	return val
}

// ParseIntWithRange 解析字符串为int类型，并设置最小值和最大值
// 如果解析失败或值小于min，返回min
// 如果值大于max，返回max
func ParseIntWithRange(str string, min, max int) int {
	val := ParseInt(str)
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// ValidateParams 统一的参数验证方法
func ValidateParams(c *gin.Context, params map[string]interface{}) bool {
	for field, value := range params {
		switch v := value.(type) {
		case int64:
			if v <= 0 {
				response.Error(c, http.StatusBadRequest, fmt.Sprintf("%s必须大于0", field))
				return false
			}
		case string:
			if v == "" {
				response.Error(c, http.StatusBadRequest, fmt.Sprintf("%s不能为空", field))
				return false
			}
		case []string:
			if len(v) == 0 {
				response.Error(c, http.StatusBadRequest, fmt.Sprintf("%s不能为空", field))
				return false
			}
		}
	}
	return true
}
