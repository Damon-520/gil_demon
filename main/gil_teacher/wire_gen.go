// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"context"
	"gil_teacher/app/conf"
	"gil_teacher/app/controller/grpc_server/live_http"
	"gil_teacher/app/controller/grpc_server/user"
	"gil_teacher/app/controller/http"
	behavior3 "gil_teacher/app/controller/http_server/behavior"
	resource_favorite2 "gil_teacher/app/controller/http_server/resource_favorite"
	"gil_teacher/app/controller/http_server/route"
	schedule2 "gil_teacher/app/controller/http_server/schedule"
	"gil_teacher/app/controller/http_server/task"
	"gil_teacher/app/controller/http_server/teacher"
	"gil_teacher/app/controller/http_server/upload"
	"gil_teacher/app/controller/providers"
	"gil_teacher/app/core/kafka"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/core/zipkinx"
	"gil_teacher/app/dao"
	"gil_teacher/app/dao/behavior"
	"gil_teacher/app/dao/live_room/impl"
	providers3 "gil_teacher/app/dao/providers"
	"gil_teacher/app/dao/task"
	behavior2 "gil_teacher/app/domain/behavior"
	"gil_teacher/app/domain/task"
	"gil_teacher/app/middleware"
	"gil_teacher/app/server"
	"gil_teacher/app/service/db_test_service"
	"gil_teacher/app/service/gil_internal/admin_service"
	"gil_teacher/app/service/gil_internal/question_service"
	"gil_teacher/app/service/live_service"
	"gil_teacher/app/service/resource_favorite"
	"gil_teacher/app/service/schedule"
	"gil_teacher/app/service/task_service"
	"gil_teacher/app/third_party/volc_ai"
	providers2 "gil_teacher/app/utils/providers"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(ctx context.Context, serverConf *conf.Server, cnf *conf.Conf, data *conf.Data, config *conf.Config, logger2 log.Logger) (*kratos.App, func(), error) {
	tracer, cleanup, err := zipkinx.NewTracer(cnf)
	if err != nil {
		return nil, nil, err
	}
	contextLogger := logger.NewContextLogger(logger2)
	middlewareMiddleware := middleware.NewMiddleware(cnf, tracer, contextLogger)
	activityDB, cleanup2, err := dao.NewActivityDB(cnf, logger2)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	iLiveRoomDao := impl.NewLiveRoomDao(activityDB, logger2)
	liveRoomService := live_service.NewLiveRoomService(logger2, iLiveRoomDao)
	liveRoomHttp := live_http.NewLiveRoomHttp(liveRoomService, logger2, config)
	userServer := user.NewUserServer(contextLogger)
	v := providers.NewServerRegisters(liveRoomHttp, userServer)
	grpcServer, err := server.NewGRPCServer(serverConf, logger2, middlewareMiddleware, v)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	dbTestClient := dao.NewDBTestClient(data, logger2)
	dbTestService := db_test_service.NewDBTestService(logger2, dbTestClient, config)
	dbTestController := http.NewDBTestController(logger2, dbTestService, config)
	ossClient := providers2.NewOSSClient(config)
	postgreSQLClient, cleanup3, err := dao.NewPostgreSQLClient(data, logger2)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	db := providers3.ProvidePostgreSQLDB(postgreSQLClient)
	resourceDAO := providers3.NewResourceDAO(db, contextLogger)
	bootstrap := conf.NewBootstrap(cnf)
	uploadController := upload.NewUploadController(ossClient, resourceDAO, contextLogger, bootstrap)
	taskDAO := dao_task.NewTaskDAO(db)
	taskAssignDAO := dao_task.NewTaskAssignDAO(db, contextLogger)
	adminClient, err := admin_service.NewAdminClient(cnf, contextLogger)
	if err != nil {
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	client := question_service.NewClient(cnf, contextLogger, adminClient)
	apiRdbClient := dao.NewApiRedisClient(cnf, contextLogger)
	taskService := task_service.NewTaskService(contextLogger, taskDAO, taskAssignDAO, client, apiRdbClient)
	taskResourceDAO := dao_task.NewTaskResourceDAO(db)
	taskResourceService := task_service.NewTaskResourceService(contextLogger, taskResourceDAO)
	ucenterClient, err := admin_service.NewUcenterClient(cnf, apiRdbClient, contextLogger)
	if err != nil {
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	teacherMiddleware := middleware.NewTeacherMiddleware(contextLogger, ucenterClient)
	volc_aiClient := volc_ai.NewClient(cnf, contextLogger)
	taskController := controller_task.NewTaskController(contextLogger, taskService, taskResourceService, client, ucenterClient, teacherMiddleware, volc_aiClient)
	teacherTempSelectionDAO := dao_task.NewTeacherTempSelectionDAO(db)
	tempSelectionService := task_service.NewTempSelectionService(contextLogger, teacherTempSelectionDAO)
	tempSelectionController := controller_task.NewTempSelectionController(contextLogger, tempSelectionService, teacherMiddleware)
	taskStudentDAO := dao_task.NewTaskStudentDao(db, contextLogger)
	taskAssignService := task_service.NewTaskAssignService(taskAssignDAO, taskStudentDAO, contextLogger)
	taskReportDAO := dao_task.NewTaskReportDAO(db, contextLogger)
	taskStudentsReportDao := dao_task.NewTaskStudentsReportDao(db, contextLogger)
	taskStudentDetailsDao := dao_task.NewTaskStudentDetailsDao(db, contextLogger)
	taskReportSettingDao := dao_task.NewTaskReportSettingDao(db, contextLogger)
	taskReportService := task_service.NewTaskStatService(taskReportDAO, taskStudentsReportDao, taskStudentDetailsDao, taskReportSettingDao, contextLogger)
	taskAnswerService := task_service.NewTaskAnswerService(taskStudentDetailsDao, contextLogger)
	v2, cleanup4, err := dao.NewClickHouseRWClient(data, contextLogger)
	if err != nil {
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	behaviorDAO := behavior.NewBehaviorDAO(v2, contextLogger)
	taskReportHandler := task.NewTaskReportHandler(taskService, taskResourceService, taskAssignService, taskReportService, ucenterClient, client, apiRdbClient, taskAnswerService, behaviorDAO, contextLogger)
	behaviorHandler := behavior2.NewBehaviorHandler(behaviorDAO, apiRdbClient, contextLogger)
	kafkaProducerClient, cleanup5, err := kafka.NewKafkaProducerClient(ctx, data, contextLogger)
	if err != nil {
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	behaviorProducer := behavior2.NewBehaviorProducer(behaviorHandler, kafkaProducerClient, contextLogger)
	taskReportController := controller_task.NewTaskReportController(taskReportHandler, teacherMiddleware, contextLogger, behaviorProducer)
	teacherController := teacher.NewTeacherController(contextLogger, ucenterClient, teacherMiddleware)
	resourceFavoriteDAO := providers3.ResourceFavoriteDAOProvider(db)
	resourceFavoriteService := resource_favorite.NewResourceFavoriteService(resourceFavoriteDAO)
	resourceFavoriteController := resource_favorite2.NewResourceFavoriteController(resourceFavoriteService, teacherMiddleware, contextLogger)
	sessionMessageHandler := behavior2.NewSessionMessageHandler(behaviorDAO, apiRdbClient, contextLogger)
	behaviorController := behavior3.NewBehaviorController(behaviorHandler, sessionMessageHandler, behaviorProducer, teacherMiddleware, contextLogger)
	scheduleCacheService := schedule.NewScheduleCacheService(apiRdbClient, contextLogger, config)
	scheduleController := schedule2.NewScheduleController(scheduleCacheService, contextLogger, teacherMiddleware)
	httpRouter := route.NewHttpRouter(dbTestController, uploadController, taskController, tempSelectionController, taskReportController, teacherController, resourceFavoriteController, behaviorController, scheduleController, teacherMiddleware)
	httpServer := server.NewGinHttpServer(cnf, contextLogger, httpRouter, middlewareMiddleware)
	app := server.NewServer(cnf, grpcServer, httpServer, contextLogger)
	return app, func() {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
	}, nil
}

// wire.go:

// newElasticsearchClient creates a new Elasticsearch client.
func NewElasticsearchClient(config *conf.Elasticsearch) (*elasticsearch.Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{config.EsURL},
		Username:  config.Username,
		Password:  config.Password,
	}
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}
