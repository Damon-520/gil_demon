package utils

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestUTF8ToGBK(t *testing.T) {
	testCases := []struct {
		input  string
		expect string
	}{
		{"测试", "b2e2cad4"}, // "测试"的GBK编码是 b2e2cad4
		{"中文", "d6d0cec4"}, // "中文"的GBK编码是 d6d0cec4
		{"你好", "c4e3bac3"}, // "你好"的GBK编码是 c4e3bac3
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			gbk, err := UTF8ToGBK(tc.input)
			if err != nil {
				t.Fatalf("UTF8ToGBK(%q) error: %v", tc.input, err)
			}
			
			// 将GBK编码的字符串转换为十六进制
			hexStr := hex.EncodeToString([]byte(gbk))
			if hexStr != tc.expect {
				t.Errorf("UTF8ToGBK(%q) = %s, want %s", tc.input, hexStr, tc.expect)
			} else {
				fmt.Printf("UTF8ToGBK(%q) = %s ✓\n", tc.input, hexStr)
			}
		})
	}
} 