package pagex

import (
	"fmt"
	"testing"
)

func TestGetPageInfo(t *testing.T) {
	params := PageParams{
		Page:         1,
		PageSize:     5,
		LimitDefault: 0,
	}

	fmt.Println("info", GetPageInfo(params))
}
