package route

import (
	"gil_teacher/app/controller/http"
	"gil_teacher/app/controller/http_server/behavior"
	"gil_teacher/app/controller/http_server/resource_favorite"
	"gil_teacher/app/controller/http_server/schedule"
	controller_task "gil_teacher/app/controller/http_server/task"
	"gil_teacher/app/controller/http_server/teacher"
	"gil_teacher/app/controller/http_server/upload"
	"gil_teacher/app/middleware"
	"gil_teacher/app/third_party/response"

	"github.com/gin-gonic/gin"
)

type HttpRouter struct {
	dbTest            *http.DBTestController
	upload            *upload.UploadController
	task              *controller_task.TaskController
	tempSelection     *controller_task.TempSelectionController
	teacher           *teacher.TeacherController
	resourceFavorite  *resource_favorite.ResourceFavoriteController
	behavior          *behavior.BehaviorController
	schedule          *schedule.ScheduleController
	taskReport        *controller_task.TaskReportController
	teacherMiddleware *middleware.TeacherMiddleware
}

func NewHttpRouter(
	dbTest *http.DBTestController,
	upload *upload.UploadController,
	task *controller_task.TaskController,
	tempSelection *controller_task.TempSelectionController,
	taskReport *controller_task.TaskReportController,
	teacher *teacher.TeacherController,
	resourceFavorite *resource_favorite.ResourceFavoriteController,
	behavior *behavior.BehaviorController,
	schedule *schedule.ScheduleController,
	teacherMiddleware *middleware.TeacherMiddleware,
) *HttpRouter {
	return &HttpRouter{
		dbTest:            dbTest,
		upload:            upload,
		task:              task,
		tempSelection:     tempSelection,
		teacher:           teacher,
		resourceFavorite:  resourceFavorite,
		behavior:          behavior,
		schedule:          schedule,
		taskReport:        taskReport,
		teacherMiddleware: teacherMiddleware,
	}
}

// heathy
func heathy(ctx *gin.Context) {
	resp := response.NewResponse()
	resp.Json(ctx)
}

// InitRouter 初始化路由配置
func (hr *HttpRouter) InitRouter(r *gin.Engine) *gin.Engine {
	r.GET("/healthz", heathy)
	// 公开接口不需要认证
	public := r.Group("/internal")
	{
		// 课程表相关路由   // 这里需要验证学生token
		internalGroup := public.Group("/api/v1")
		{
			internalGroup.GET("/is_teaching", hr.schedule.IsTeachingNow)                  // 判断老师当前是否在上课
			internalGroup.GET("/teaching_status", hr.schedule.CheckTeacherTeachingStatus) // 检查教师教学状态详情

			internalGroup.GET("/teachers", hr.schedule.GetTeacherList) // 获取教师列表
			internalGroup.GET("/store", hr.schedule.StoreScheduleData) // 直接保存课程表数据，脚本使用
		}
		// 暴露给学生端的内部接口
		{
			internalGroup.POST("/student/task/list", hr.task.GetStudentTaskList) // 查询学生的任务列表
		}
	}

	// 需要认证的接口
	authorized := r.Group("/api/v1")

	// 认证中间件，获取教师信息并存储到context中
	authorized.Use(hr.teacherMiddleware.WithTeacherContext())

	{
		// 课程表相关路由 - 需要认证
		scheduleAuthGroup := authorized.Group("/schedule")
		{
			scheduleAuthGroup.GET("/is_teaching", hr.schedule.IsTeachingNow)           // 判断老师当前是否在上课
			scheduleAuthGroup.GET("/classroom/status", hr.schedule.GetClassroomStatus) // 获取班级名称和上下课状态

			scheduleAuthGroup.GET("/store/auth", hr.schedule.StoreScheduleData)          // 直接保存课程表数据（需要认证）
			scheduleAuthGroup.POST("/store/direct", hr.schedule.StoreDirectScheduleData) // 直接从Redis获取课程表数据
		}
		// 上传相关路由
		uploadGroup := authorized.Group("/upload")
		{
			uploadGroup.POST("/presigned-url", hr.upload.GetPresignedPutURL)     // 获取预签名PUT上传URL
			uploadGroup.POST("/notify-complete", hr.upload.NotifyUploadComplete) // 通知预签名URL上传完成
			uploadGroup.POST("/update-filename", hr.upload.UpdateFileName)       // 更新文件名
			uploadGroup.GET("/query", hr.upload.QueryResources)                  // 通用资源查询
			uploadGroup.POST("/delete", hr.upload.DeleteResource)                // 删除资源
		}

		// 数据库测试路由
		hr.dbTest.RegisterRoutes(r)

		// 任务相关路由
		taskGroup := authorized.Group("/task")
		{
			taskGroup.GET("/type", hr.task.GetTaskType) // 查询老师布置任务具备权限的学科及任务类型

			// 标准作业/公共资源，教师需要先查看相关资源才能发布任务
			taskGroup.GET("/knowledge-tree/list", hr.task.GetKnowledgeTreeList)        // 获取知识点类型业务树的列表
			taskGroup.GET("/knowledge-tree/detail", hr.task.GetBizTreeDetail)          // 获取知识点类型业务树的详情
			taskGroup.GET("/chapter-tree/list", hr.task.GetChapterBizTreeList)         // 获取教材类型业务树列表
			taskGroup.GET("/chapter-tree/detail", hr.task.GetBizTreeDetail)            // 获取教材类型业务树详情
			taskGroup.GET("/course-practice/list", hr.task.GetAICourseAndPracticeList) // 获取教材章节（业务树节点）下的 AI 课和巩固练习列表
			taskGroup.GET("/question-set/detail", hr.task.GetQuestionSetDetailByID)    // 通过题集ID获取题集详情，巩固练习是题集的一种
			taskGroup.GET("/question/enums", hr.task.GetQuestionEnums)                 // 获取查询题目支持的下拉框枚举值
			taskGroup.POST("/question/list/search", hr.task.GetQuestionList)           // 搜索题目列表
			taskGroup.POST("/question/detail", hr.task.GetQuestionListByIDs)           // 通过ID列表查询题目详情

			taskGroup.GET("/management/detail", hr.task.GetTaskByID)              // 根据 ID 查询单个任务
			taskGroup.POST("/management/create", hr.task.CreateTask)              // 创建任务
			taskGroup.POST("/management/update", hr.task.UpdateTask)              // 更新任务名称
			taskGroup.POST("/management/delete", hr.task.DeleteTask)              // 删除任务
			taskGroup.POST("/management/assign/update", hr.task.UpdateTaskAssign) // 更新任务分配的时间
			taskGroup.POST("/management/assign/delete", hr.task.DeleteTaskAssign) // 删除任务分配
		}

		// 临时选择（试题篮、资源篮）相关路由
		teacherTempSelectionGroup := authorized.Group("/temp-selection")
		{
			teacherTempSelectionGroup.POST("/create", hr.tempSelection.CreateSelection)  // 创建教师临时选择
			teacherTempSelectionGroup.POST("/delete", hr.tempSelection.DeleteSelections) // 删除教师临时选择
			teacherTempSelectionGroup.GET("/list", hr.tempSelection.ListSelections)      // 查询教师临时选择列表
		}

		// 教师相关路由
		teacherGroup := authorized.Group("/teacher")
		{
			teacherGroup.GET("/detail", hr.teacher.GetTeacherDetail)         // 获取教师的学校、职务、班级、个人信息
			teacherGroup.GET("/class/students", hr.teacher.GetClassStudents) // 查询班级下的学生列表
		}

		// 资源收藏路由
		favoriteGroup := authorized.Group("/resource/favorite")
		{
			favoriteGroup.POST("/create", hr.resourceFavorite.CreateFavorite) // 创建资源收藏
			favoriteGroup.POST("/list", hr.resourceFavorite.ListFavorites)    // 获取收藏列表
			favoriteGroup.POST("/cancel", hr.resourceFavorite.CancelFavorite) // 取消收藏
		}

		// 会话相关
		sessionGroup := authorized.Group("/session")
		{
			sessionGroup.POST("/open", hr.behavior.OpenSession)                  // 创建会话
			sessionGroup.POST("/message", hr.behavior.SaveMessage)               // 记录会话内容
			sessionGroup.POST("/close", hr.behavior.CloseSession)                // 关闭会话
			sessionGroup.GET("/messages", hr.behavior.GetSessionMessages)        // 查询指定会话的全部消息
			sessionGroup.POST("/message/read", hr.behavior.MarkMessageAsRead)    // 标记消息已读（最后一条消息id）
			sessionGroup.GET("/unread-count", hr.behavior.GetUnreadMessageCount) // 获取用户指定会话的未读消息数量
			sessionGroup.GET("/unread-list", hr.behavior.GetUnreadMessageList)   // 获取用户指定会话的未读消息列表
		}
		// 行为相关路由
		behaviorGroup := authorized.Group("/behavior") // 行为相关路由
		{
			behaviorGroup.POST("/teacher", hr.behavior.RecordTeacherBehavior)                         // 记录教师行为
			behaviorGroup.POST("/student", hr.behavior.RecordStudentBehavior)                         // 记录学生行为
			behaviorGroup.POST("/student/evaluate", hr.behavior.TeacherEvaluateStudent)               // 教师评价学生
			behaviorGroup.GET("/classroom/messages", hr.behavior.GetClassroomMessages)                // 查询指定课堂的全部消息
			behaviorGroup.GET("/class/latest-behaviors", hr.behavior.GetClassLatestBehaviors)         // 获取班级学生最新行为
			behaviorGroup.GET("/student/classroom-detail", hr.behavior.GetStudentClassroomDetail)     // 获取学生课堂详情
			behaviorGroup.GET("/class/behavior-category", hr.behavior.GetClassBehaviorCategory)       // 获取课堂行为分类列表
			behaviorGroup.GET("/classroom/behavior-summary", hr.behavior.GetClassroomBehaviorSummary) // 获取课后行为汇总统计
			behaviorGroup.POST("/praise", hr.behavior.PraiseStudents)                                 // 表扬学生
			behaviorGroup.POST("/attention", hr.behavior.AttentionStudents)                           // 关注学生
			behaviorGroup.GET("/classroom/learning-scores", hr.behavior.GetClassroomLearningScores)   // 获取课堂学习分列表
		}

		// 报告相关
		taskReportGroup := authorized.Group("/task/report")
		{
			taskReportGroup.GET("/subject-class/list", hr.taskReport.GetSubjectClassList)    // 获取教师查看作业报告时具备的学科和班级列表
			taskReportGroup.GET("/latest", hr.taskReport.LatestReport)                       // 查询每个任务类型最近布置的单个作业报告(指定教师)
			taskReportGroup.GET("/list", hr.taskReport.ListReports)                          // 查询作业报告列表(指定教师)
			taskReportGroup.GET("/detail", hr.taskReport.GetReportSummaryDetail)             // 查询任务的作业汇总报告(指定班级或小组)
			taskReportGroup.GET("/answers", hr.taskReport.GetAnswers)                        // 查询任务的作业答题结果(指定班级或小组)
			taskReportGroup.GET("/export", hr.taskReport.ExportReport)                       // 导出作业报告
			taskReportGroup.GET("/answer-panel", hr.taskReport.GetAnswerPanel)               // 题目面板，每个题目的正确率，方便老师查看和直接点击跳转
			taskReportGroup.GET("/suggestion-template", hr.taskReport.GetSuggestionTemplate) // 获取建议消息模板，目前前端写死了
			taskReportGroup.GET("/student/answers", hr.taskReport.GetStudentAnswers)         // 获取任务指定学生的全部答题结果
			taskReportGroup.POST("/student/handle", hr.taskReport.HandleStudentReport)       // 对学生作业报告的处理，目前只有：点赞、提醒
			taskReportGroup.GET("/student/detail", hr.taskReport.GetStudentDetail)           // 作业/点击学生头像/学生详情
			taskReportGroup.GET("/questions", hr.taskReport.GetQuestions)                    // 查询任务的全部提问记录(指定班级或小组) TODO
			taskReportGroup.GET("/knowledge/questions", hr.taskReport.GetKnowledgeQuestions) // 获取任务单个知识点的全部提问记录 TODO
			taskReportGroup.GET("/setting", hr.taskReport.GetReportSetting)                  // 获取报告参数设置
			taskReportGroup.POST("/setting/update", hr.taskReport.UpdateReportSetting)       // 更新或设置报告参数设置
		}
	}

	return r
}
