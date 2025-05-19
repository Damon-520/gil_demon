package base62x

import (
	"bytes"
	"errors"
	"math"
	"math/big"
)

var base62Charset = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

// IntToBase62 将整型转换为字符串
//
// 转换逻辑其实和10进制转16进制数字类似，只是这里将10进制转换为62进制
func IntToBase62(num int64) string {
	if num <= 0 {
		return ""
	}

	var result bytes.Buffer
	x := big.NewInt(num)

	for x.Cmp(big.NewInt(0)) == 1 {
		r := new(big.Int)
		x.DivMod(x, big.NewInt(62), r)
		result.WriteByte(base62Charset[r.Int64()])
		x.Set(x)
	}

	return reverse(result.String())
}

func reverse(s string) string {
	bytes := []byte(s)
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return string(bytes)
}

// Base62ToInt 将字符串转换为整型
//
// 转换逻辑其实和16进制转10进制数字类似，只是这里将62进制转换为10进制
func Base62ToInt(str string) (int64, error) {
	var result big.Int

	for i := range str {
		index := bytes.IndexByte(base62Charset, str[i])
		if index == -1 {
			return 0, errors.New("invalid base62 string")
		}

		result.Mul(&result, big.NewInt(62))
		result.Add(&result, big.NewInt(int64(index)))
	}

	if result.Cmp(big.NewInt(math.MaxInt64)) == 1 {
		return 0, errors.New("integer overflow")
	}

	return result.Int64(), nil
}
