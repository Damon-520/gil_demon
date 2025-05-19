/**
   @author: tony.zhao
   @data : 2022/12/12 22:07
   @desc:
**/

package sinterx

import (
	"fmt"
	"strconv"
	"testing"
)

func Test_Sinter(t *testing.T) {
	slice1 := make([]string, 0)
	slice2 := make([]string, 0)
	for i := 1; i < 300000; i++ {
		slice1 = append(slice1, strconv.FormatInt(int64(i), 10))
	}

	for j := 10000; j < 300000; j++ {
		slice2 = append(slice2, strconv.FormatInt(int64(j), 10))
	}

	//fmt.Println("交集", len(intersenct(slice1, slice2)))
	fmt.Println("并集", len(union(slice1, slice2)))
	//fmt.Println("差集", len(difference(slice1, slice2)))

}
