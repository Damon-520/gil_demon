package task

import (
	"gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	behaviorDao "gil_teacher/app/dao/behavior"
	"gil_teacher/app/service/gil_internal/admin_service"
	"gil_teacher/app/service/gil_internal/question_service"
	"gil_teacher/app/service/task_service"
)

type TaskReportHandler struct {
	taskService         *task_service.TaskService
	taskResourceService *task_service.TaskResourceService
	taskAssignService   *task_service.TaskAssignService
	taskReportService   *task_service.TaskReportService
	ucenterService      *admin_service.UcenterClient
	questionAPI         *question_service.Client
	redisClient         *dao.ApiRdbClient
	taskAnswerService   *task_service.TaskAnswerService
	behaviorDAO         behaviorDao.BehaviorDAO
	log                 *logger.ContextLogger
}

func NewTaskReportHandler(
	taskService *task_service.TaskService,
	taskResourceService *task_service.TaskResourceService,
	taskAssignService *task_service.TaskAssignService,
	taskStatService *task_service.TaskReportService,
	ucenterService *admin_service.UcenterClient,
	questionAPI *question_service.Client,
	redisClient *dao.ApiRdbClient,
	taskAnswerService *task_service.TaskAnswerService,
	behaviorDAO behaviorDao.BehaviorDAO,
	logger *logger.ContextLogger,
) *TaskReportHandler {
	return &TaskReportHandler{
		taskService:         taskService,
		taskResourceService: taskResourceService,
		taskAssignService:   taskAssignService,
		taskReportService:   taskStatService,
		ucenterService:      ucenterService,
		questionAPI:         questionAPI,
		redisClient:         redisClient,
		taskAnswerService:   taskAnswerService,
		behaviorDAO:         behaviorDAO,
		log:                 logger,
	}
}
