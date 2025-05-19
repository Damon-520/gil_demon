package gorm_builder

import (
	"fmt"
	"testing"
)

func Test_Sql(t *testing.T) {

	options := Options{
		Conditions: map[string]interface{}{
			"eq|id":    1,
			"in|name":  "demo",
			"not|test": []string{"1", "2", "3"},
		},
	}

	sql, vals := BuildWhere(options)

	fmt.Println(sql)  // test not in (?) and id = ? and name in (?)
	fmt.Println(vals) // [1 demo [1 2 3]]
}
