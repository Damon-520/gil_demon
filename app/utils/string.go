package utils

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// UTF8ToGBK 将UTF-8编码的字符串转换为GBK编码
func UTF8ToGBK(s string) (string, error) {
	reader := transform.NewReader(bytes.NewReader([]byte(s)), simplifiedchinese.GBK.NewEncoder())
	d, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

// 把 list[string|int64|float64] 合并为 string
func JoinList(list []any, separator string) string {
	parts := make([]string, 0)
	for _, item := range list {
		parts = append(parts, fmt.Sprintf("%v", item))
	}
	return strings.Join(parts, separator)
}
