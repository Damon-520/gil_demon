package idtools

import (
	"math/rand"

	"github.com/google/uuid"
)

// 返回长度为l的字符串，只包含数字和小写字母
func GetStr(l int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, l)
	for i := range l {
		result[i] = charset[GetRandomInt(len(charset))]
	}
	return string(result)
}

func GetRandomInt(n int) int {
	return rand.Intn(n)
}

func GetUUID() string {
	return uuid.New().String()
}
