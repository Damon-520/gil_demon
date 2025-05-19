package auth

import (
	"context"
	"strconv"

	"gil_teacher/app/consts"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

func ParseHeader(logger log.Logger) middleware.Middleware {

	return func(handler middleware.Handler) middleware.Handler {

		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {

			// 解析transport
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			// 获取请求头
			header := tr.RequestHeader()
			// fmt.Printf("%+v\n", header)
			// fmt.Printf("%+v\n", header.Keys())

			if userId := header.Get("uid"); userId != "" {
				if uId, err := strconv.ParseInt(userId, 10, 32); err == nil {
					ctx = context.WithValue(ctx, consts.FormHeaderCtxUserIdKey, uId)
				}
			}

			return handler(ctx, req)

		}
	}
}
