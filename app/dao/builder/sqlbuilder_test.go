package builder

import (
	"fmt"
	"testing"
)

func TestBulider(t *testing.T) {

	Cond := make(ConditionOptions)
	Cond[JoinAnd] = []WhereCondition{
		{"id", WhereBetween, []int{10, 20}},
		{"id", WhereIn, []string{"1", "2"}},
	}
	Cond[JoinOr] = []WhereCondition{
		{"num", WhereEq, "abc"},
		{"id", WhereLt, 2},
		{"id", WhereIn, []int{1, 2}},
	}

	fmt.Println(GetConditions(Cond))

}
