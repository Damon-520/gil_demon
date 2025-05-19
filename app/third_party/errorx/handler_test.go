package errorx

import (
	errors2 "errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"testing"
)

func Foo1() error {
	return ErrRequest
}

func Foo2() error {
	return errors2.New("原始错误")
}

func Foo3() error {
	return errors.New(100, "NotFund", "没有找到")
}

func Test_Demo1(t *testing.T) {

	e := Foo1()
	ex := Cause(e)
	fmt.Println(ex.Code())
	fmt.Println(ex.Message())
}

func Test_Demo2(t *testing.T) {
	e := Foo2()
	ex := Cause(e)
	fmt.Println(ex.Code())
	fmt.Println(ex.Message())
}

func Test_Demo3(t *testing.T) {
	e := Foo3()
	ex := Cause(e)
	fmt.Println(ex.Code())
	fmt.Println(ex.Message())
}
