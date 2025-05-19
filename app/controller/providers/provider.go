//go:build wireinject
// +build wireinject

package providers

import (
	"gil_teacher/app/conf"
	"gil_teacher/app/controller/grpc_server/live_http"
	"gil_teacher/app/controller/grpc_server/user"
	"gil_teacher/app/controller/http"
	"gil_teacher/app/controller/http_server/behavior"
	"gil_teacher/app/controller/http_server/resource_favorite"
	"gil_teacher/app/controller/http_server/route"
	"gil_teacher/app/controller/http_server/schedule"
	controller_task "gil_teacher/app/controller/http_server/task"
	"gil_teacher/app/controller/http_server/teacher"
	"gil_teacher/app/controller/http_server/upload"
	utilproviders "gil_teacher/app/utils/providers"

	"github.com/google/wire"
)

var GRPCProviderSet = wire.NewSet(
	live_http.NewLiveRoomHttp,
	user.NewUserServer,
	NewServerRegisters,
)

var HttpProviderSet = wire.NewSet(
	http.NewDBTestController,
	controller_task.NewTaskController,
	controller_task.NewTempSelectionController,
	teacher.NewTeacherController,
	upload.NewUploadController,
	resource_favorite.NewResourceFavoriteController,
	behavior.NewBehaviorController,
	controller_task.NewTaskReportController,
	schedule.NewScheduleController,
)

var ControllerProviderSet = wire.NewSet(
	GRPCProviderSet,
	HttpProviderSet,
	route.NewHttpRouter,
	utilproviders.OSSClientProvider,
	conf.NewBootstrap,
)
