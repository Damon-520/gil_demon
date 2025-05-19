package errorx

import (
	"fmt"
	"github.com/pkg/errors"
)

var (
	codes = map[int32]struct{}{}
)

// New Error

func New(code int32, msg string) Error {
	if code < 1000 {
		panic("error code must be greater than 1000")
	}
	return add(code, msg)
}

// add only inner error

func add(code int32, msg string) Error {

	if _, ok := codes[code]; ok {

		panic(fmt.Sprintf("ecode: %d already exist", code))

	}

	codes[code] = struct{}{}

	return Error{

		code: code, message: msg,
	}

}

type Errors interface {

	// sometimes Error return Code in string form

	Error() string

	// Code get error code.

	Code() int32

	// Message get code message.

	Message() string

	// Detail get error detail,it may be nil.

	Details() []interface{}

	// Equal for compatible.

	Equal(error) bool

	// Reload Message

	Reload(string) Error
}

type Error struct {
	code    int32
	message string
}

func (e Error) Error() string {

	return e.message

}

func (e Error) Message() string {

	return e.message

}

func (e Error) Reload(message string) Error {

	e.message = message

	return e

}

func (e Error) Code() int32 {

	return e.code

}

func (e Error) Details() []interface{} { return nil }

func (e Error) Equal(err error) bool { return Equal(err, e) }

func String(e string) Error {

	if e == "" {

		return Ok

	}

	return Error{

		code: -1, message: e,
	}

}

func Cause(err error) Errors {

	if err == nil {
		return Ok
	}
	if ec, ok := errors.Cause(err).(Errors); ok {
		return ec
	}

	return String(err.Error())
}

// Equal

func Equal(err error, e Error) bool {

	return Cause(err).Code() == e.Code()

}
