package validator

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
)

type validator interface {
	Validate() error
}

var REASON = "VALIDATOR"

func ErrorValidator(format string, args ...interface{}) *errors.Error {
	return errors.New(200, REASON, fmt.Sprintf(format, args...)).WithMetadata(map[string]string{"stat": "0", "code": "100"})
}

func IsErrorValidator(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	stat := e.Metadata["stat"]
	code := e.Metadata["code"]
	return (e.Reason == REASON || (stat == "0" && code == "100")) && e.Code == 200
}

// Validator is a validator middleware.
func Validator() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if v, ok := req.(validator); ok {
				if err := v.Validate(); err != nil {
					return nil, ErrorValidator("参数验证失败:%s", err.Error())
				}
			}
			return handler(ctx, req)
		}
	}
}
