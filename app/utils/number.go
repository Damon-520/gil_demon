package utils

import (
	"math"
	"strconv"
)

func Atoi64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func Atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func AtoBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

// 计算平均值float32
func AvgFloat32(values []float32) float32 {
	if len(values) == 0 {
		return 0
	}
	sum := float32(0)
	for _, value := range values {
		sum += value
	}
	return sum / float32(len(values))
}

// 计算平均值 float64
func AvgFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := float64(0)
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// 计算平均值 int
func AvgInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	sum := 0
	for _, value := range values {
		sum += value
	}
	return sum / len(values)
}

// 计算平均值 int64
func AvgInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	sum := int64(0)
	for _, value := range values {
		sum += value
	}
	return sum / int64(len(values))
}

// int64 转 string
func I64ToStr(i int64) string {
	return strconv.FormatInt(i, 10)
}

// float64 转 string
func F64ToStr(f float64) string {
	if f == math.MaxFloat64 {
		return "+inf"
	}
	return F64ToString(f, -1, "")
}

// float64 转 string，支持精度控制和默认值
func F64ToString(f float64, precision int, defaultStr string) string {
	if f == math.MaxFloat64 {
		return "+inf"
	}
	if f == 0 && defaultStr != "" {
		return defaultStr
	}

	if precision < 0 {
		// 使用默认精度
		return strconv.FormatFloat(f, 'f', -1, 64)
	} else {
		// 限制精度范围
		if precision > 4 {
			precision = 4
		}
		return strconv.FormatFloat(f, 'f', precision, 64)
	}
}

// 数字除法，要考虑分母 0 的情况
func I64Div(a, b int64) int64 {
	if b == 0 {
		return 0
	}
	return a / b
}

// 浮点数除法，要考虑分母 0 的情况，可以指定返回精度，默认为 4
func F64Div(a, b float64, precision int) float64 {
	if b == 0 {
		return 0
	}

	if precision < 0 {
		precision = 4
	}
	return math.Round(a/b*math.Pow(10, float64(precision))) / math.Pow(10, float64(precision))
}

// 计算百分比
func F64Percent(a, b float64, precision int) float64 {
	return F64Div(a, b, precision) * 100
}

// 计算最小值
func F64Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
