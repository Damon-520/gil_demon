package builder

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type ConditionOptions map[string][]WhereCondition

type WhereCondition []interface{}

func GetConditions(cond ConditionOptions) (condition string, args []interface{}, err error) {
	defer func(err error) {
		if e := recover(); e != nil {
			condition = ""
			args = make([]interface{}, 0)
			err = errors.New("Where参数不正确")
		}
	}(err)
	condition = ""
	if len(cond) == 0 {
		args = make([]interface{}, 0)
		return condition, args, errors.New("Where参数不正确")
	}
	Join := " and "
	Conditions := make([]string, 0)
	for joinCond, where := range cond {
		SubJoin := " and "
		switch joinCond {
		case JoinAnd:
			SubJoin = " and "
		case JoinOr:
			SubJoin = " or "

		}
		if len(where) == 0 {
			return condition, args, errors.New("join参数不正确")
		}

		SubConditions := make([]string, 0, len(where))
		for _, whereCond := range where {
			if len(whereCond) < 3 {
				continue
			}
			filed := fmt.Sprintf("%v", whereCond[0])
			option := fmt.Sprintf("%v", whereCond[1])
			value := whereCond[2]

			switch option {
			case WhereBetween:
				SubConditions = append(SubConditions, fmt.Sprintf(" (%v %v ? and ? )", filed, whereMap[option]))
				t := reflect.ValueOf(value)
				for i := 0; i < t.Len(); i++ {
					args = append(args, t.Index(i).Interface())
				}
			case WhereIn:
				t := reflect.ValueOf(value)
				inVal := make([]string, 0)
				for i := 0; i < t.Len(); i++ {
					v := fmt.Sprintf("%v", t.Index(i).Interface())
					if v == "" {
						continue
					}
					inVal = append(inVal, v)
				}
				if len(inVal) == 0 {
					return option, args, errors.New("WhereIn参数不正确")
				}
				if len(inVal) == 1 {
					SubConditions = append(SubConditions, fmt.Sprintf(" %v %v ? ", filed, whereMap[WhereEq]))
					args = append(args, inVal[0])
				} else {
					SubConditions = append(SubConditions, fmt.Sprintf(" %v %v ? ", filed, whereMap[WhereIn]))

					args = append(args, inVal)
				}

			default:
				SubConditions = append(SubConditions, fmt.Sprintf(" %v %v ? ", filed, whereMap[option]))
				args = append(args, value)
			}

		}

		Conditions = append(Conditions, fmt.Sprintf(" ( %v ) ", strings.Join(SubConditions, SubJoin)))

	}
	condition = strings.Join(Conditions, Join)

	return condition, args, nil
}
