package provider

import (
	"gil_teacher/app/middleware"

	"github.com/google/wire"
)

var ServerProviderSet = wire.NewSet(
	middleware.NewMiddleware,
	middleware.NewTeacherMiddleware,
)
