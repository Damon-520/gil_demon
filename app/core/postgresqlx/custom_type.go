package postgresqlx

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// Int64Array 自定义类型，用于处理PostgreSQL的bigint[]类型
type Int64Array []int64

// Value 实现driver.Valuer接口，将[]int64转换为PostgreSQL数组格式
func (a Int64Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	if len(a) == 0 {
		return "{}", nil
	}

	// 将[]int64转换为PostgreSQL数组格式，如{1,2,3}
	str := "{"
	for i, v := range a {
		if i > 0 {
			str += ","
		}
		str += fmt.Sprintf("%d", v)
	}
	str += "}"
	return str, nil
}

// Scan 实现sql.Scanner接口，将PostgreSQL数组格式转换为[]int64
func (a *Int64Array) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	// 将PostgreSQL数组格式转换为[]int64
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("无法将%T转换为string", value)
	}

	// 去掉{}并分割
	str = strings.Trim(str, "{}")
	if str == "" {
		*a = Int64Array{}
		return nil
	}

	// 分割并转换为[]int64
	parts := strings.Split(str, ",")
	result := make(Int64Array, len(parts))
	for i, part := range parts {
		var v int64
		_, err := fmt.Sscanf(part, "%d", &v)
		if err != nil {
			return err
		}
		result[i] = v
	}
	*a = result
	return nil
}

// StringArray 自定义字符串数组，用于处理 PostgreSQL 的 text[] 类型
type StringArray []string

// Value 实现 driver.Valuer 接口，将 []string 转换为 PostgreSQL 的 text[] 类型
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}

	// 将 []string 转换为 PostgreSQL 的 text[] 类型
	str := "{"
	for i, v := range a {
		if i > 0 {
			str += ","
		}
		str += fmt.Sprintf("'%s'", v)
	}
	str += "}"
	return str, nil
}

// Scan 实现 sql.Scanner 接口，将 PostgreSQL 的 text[] 类型转换为 []string
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	// 将 PostgreSQL 的 text[] 类型转换为 []string
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("无法将%T转换为string", value)
	}

	// 去掉{}并分割
	str = strings.Trim(str, "{}")
	if str == "" {
		*a = StringArray{}
		return nil
	}

	// 分割并转换为[]string
	parts := strings.Split(str, ",")
	result := make(StringArray, len(parts))
	for i, part := range parts {
		result[i] = strings.Trim(part, "'")
	}
	*a = result
	return nil
}
