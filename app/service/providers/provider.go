package provider

import (
	"gil_teacher/app/service/behavior"
	db_test_service "gil_teacher/app/service/db_test_service"
	"gil_teacher/app/service/gil_internal/admin_service"
	"gil_teacher/app/service/gil_internal/question_service"
	"gil_teacher/app/service/live_service"
	"gil_teacher/app/service/resource_favorite"
	"gil_teacher/app/service/schedule"
	"gil_teacher/app/service/task_service"

	"github.com/google/wire"
)

var ServiceProviderSet = wire.NewSet(
	live_service.NewLiveRoomService,
	db_test_service.NewDBTestService,
	task_service.NewTaskService,
	task_service.NewTaskAssignService,
	task_service.NewTaskAnswerService,
	task_service.NewTaskStatService,
	task_service.NewTempSelectionService,
	task_service.NewTaskResourceService,
	resource_favorite.NewResourceFavoriteService,
	question_service.NewClient,
	admin_service.NewUcenterClient,
	admin_service.NewAdminClient,
	schedule.NewScheduleCacheService,
	behavior.NewBehaviorService,
)
