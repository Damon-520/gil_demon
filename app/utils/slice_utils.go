package utils

import (
	"strconv"
	"strings"
)

// RemoveDuplicateInt64 移除切片中的重复元素
func RemoveDuplicateInt64(slice []int64) []int64 {
	uniqueMap := make(map[int64]struct{})
	uniqueSlice := []int64{}
	for _, item := range slice {
		if _, ok := uniqueMap[item]; !ok {
			uniqueMap[item] = struct{}{}
			uniqueSlice = append(uniqueSlice, item)
		}
	}
	return uniqueSlice
}

// Int64SliceToString 将 int64 切片转换为以逗号分隔的 string
func Int64SliceToString(slice []int64) string {
	strSlice := make([]string, len(slice))
	for i, item := range slice {
		strSlice[i] = strconv.FormatInt(item, 10)
	}
	return strings.Join(strSlice, ",")
}
