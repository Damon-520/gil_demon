package gorm_builder

import (
	"fmt"
	"strings"
)

type Options struct {
	Conditions map[string]interface{} // "eq|id": 1
	Order      string
	Offset     int
	Limit      int
	IsCount    bool // 是否统计总数
}

func BuildWhere(options Options) (whereSql string, values []interface{}) {

	for key, val := range options.Conditions {
		keys := strings.Split(key, "|")
		if len(keys) < 2 {
			continue
		}

		operator := strings.Trim(keys[0], " ") // 操作符
		field := strings.Trim(keys[1], " ")    // 字段名
		if len(field) == 0 {
			continue
		}
		switch operator {
		case "eq", "=":
			whereSql += fmt.Sprint(field, " = ? and ")
			values = append(values, val)
		case "in":
			whereSql += fmt.Sprint(field, " in (?) and ")
			values = append(values, val)
		case "notIn":
			whereSql += fmt.Sprint(field, " not in (?) and ")
			values = append(values, val)
		case "gt", ">":
			whereSql += fmt.Sprint(field, " > ? and ")
			values = append(values, val)
		case "lt", "<":
			whereSql += fmt.Sprint(field, " < ? and ")
			values = append(values, val)
		case "gte", ">=":
			whereSql += fmt.Sprint(field, " >= ? and ")
			values = append(values, val)
		case "lte", "<=":
			whereSql += fmt.Sprint(field, " <= ? and ")
			values = append(values, val)
		case "neq", "!=":
			whereSql += fmt.Sprint(field, " != ? and ")
			values = append(values, val)
		case "like":
			whereSql += fmt.Sprint(field, " like ? and ")
			values = append(values, fmt.Sprint("%", val, "%"))
		default:
			continue
		}
	}

	return strings.TrimSuffix(whereSql, " and "), values
}
