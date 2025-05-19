package controller_task

// 注意：此文件需要 Go 1.22 或更高版本
// 主要依赖：
// - encoding/csv: 用于 CSV 文件生成
// - net/url: 用于文件名 URL 编码
// - bytes: 用于缓冲区操作
// - fmt: 用于字符串格式化

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/domain/behavior"
	"gil_teacher/app/domain/task"
	"gil_teacher/app/middleware"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils"

	"github.com/gin-gonic/gin"
)

// TaskReportController 作业报告控制器
type TaskReportController struct {
	taskReportHandler *task.TaskReportHandler
	log               *logger.ContextLogger
	teacherMiddleware *middleware.TeacherMiddleware
	producer          *behavior.BehaviorProducer
}

func NewTaskReportController(
	taskReportHandler *task.TaskReportHandler,
	teacherMiddleware *middleware.TeacherMiddleware,
	log *logger.ContextLogger,
	producer *behavior.BehaviorProducer,
) *TaskReportController {
	return &TaskReportController{
		taskReportHandler: taskReportHandler,
		teacherMiddleware: teacherMiddleware,
		log:               log,
		producer:          producer,
	}
}

// GetSubjectClassList 获取教师查看作业报告时具备的学科和班级列表
func (c *TaskReportController) GetSubjectClassList(ctx *gin.Context) {
	// 一个用户可以具备多个角色，取每个角色的学科和班级的并集
	// 校长：具备全部学科、全部班级的权限
	// 年级主任：具备全部学科、指定年级下的全部班级的权限
	// 学科组长：具备指定学科、指定年级下的全部班级的权限
	// 学科教师：具备指定学科、指定年级下的指定班级的权限
	// 班主任：具备全部学科、指定班级的权限
	subjects := c.teacherMiddleware.GetTaskReportSubjects(ctx)
	gradeClasses, err := c.teacherMiddleware.GetTaskReportGradeClasses(ctx)
	if err != nil {
		c.log.Error(ctx, "获取运营平台年级班级列表失败 error:%v", err)
		response.Err(ctx, response.ERR_GIL_ADMIN)
		return
	}

	// 转换学科列表
	subjectsList := make([]api.Subject, len(subjects))
	for i, subject := range subjects {
		subjectsList[i] = api.Subject{
			SubjectKey:  subject,
			SubjectName: consts.SubjectNameMap[subject],
		}
	}
	res := &api.GetSubjectClassListResponse{
		Subjects:     subjectsList,
		GradeClasses: gradeClasses,
	}
	response.Success(ctx, res)
}

// LatestReport 查询最近布置的单个作业(指定教师)
func (c *TaskReportController) LatestReport(ctx *gin.Context) {
	teacherID, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		response.SystemError(ctx)
		return
	}
	classInfo := c.teacherMiddleware.ExtractTeacherClassInfo(ctx)
	subjectID := utils.Atoi64(ctx.Query("subject"))
	query := &dto.TeacherLatestTaskReportsQuery{
		TeacherID: teacherID,
		SchoolID:  schoolID,
		Subject:   subjectID,
		ClassInfo: classInfo,
	}
	report, err := c.taskReportHandler.GetTeacherLatestTaskReport(ctx, query)
	if err != nil {
		c.log.Error(ctx, "GetLatestTeacherTaskReport error:%v", err)
		response.SystemError(ctx)
		return
	}
	response.Success(ctx, report)
}

// ListReports 查询教师布置的作业报告列表
func (c *TaskReportController) ListReports(ctx *gin.Context) {
	teacherID, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		response.SystemError(ctx)
		return
	}
	classInfo := c.teacherMiddleware.ExtractTeacherClassInfo(ctx)
	subjectID := utils.Atoi64(ctx.Query("subject")) // 需要检查科目权限 TODO

	startTime, _ := utils.DateFirstSecondTimestamp(ctx.Query("startDate"))
	endTime, _ := utils.GetDateLastSecondTimestamp(ctx.Query("endDate"))
	query := &dto.TaskAssignListQuery{
		TeacherID: teacherID,
		SchoolID:  schoolID,
		Subject:   subjectID,
		Keyword:   ctx.Query("keyword"),
		GroupType: utils.Atoi64(ctx.Query("groupType")),
		GroupID:   utils.Atoi64(ctx.Query("groupId")),
		TaskType:  utils.Atoi64(ctx.Query("taskType")),
		StartTime: startTime,
		EndTime:   endTime,
		ClassInfo: classInfo,
	}
	pageInfo := &consts.APIReqeustPageInfo{
		Page:     utils.Atoi64(ctx.Query("page")),
		PageSize: utils.Atoi64(ctx.Query("pageSize")),
		SortType: consts.SortTypeDesc,
		SortBy:   "createTime",
	}
	pageInfo.Check()
	taskReportList, err := c.taskReportHandler.GetTeacherTasksReportList(ctx, query, pageInfo)
	if err != nil {
		c.log.Error(ctx, "GetTeacherTaskReportList error:%v", err)
		response.SystemError(ctx)
		return
	}
	response.Success(ctx, taskReportList)
}

// GetReportSummaryDetail 查询布置对象的作业报告详情
func (c *TaskReportController) GetReportSummaryDetail(ctx *gin.Context) {
	// get 参数
	taskId := utils.Atoi64(ctx.Query("taskId"))
	assignId := utils.Atoi64(ctx.Query("assignId"))
	// check 参数
	if taskId == 0 || assignId == 0 {
		c.log.Error(ctx, "taskId and assignId are required")
		response.ParamError(ctx, response.ERR_EMPTY_TASK_OR_ASSIGN)
		return
	}

	apiPageInfo := &consts.APIReqeustPageInfo{
		Page:          utils.Atoi64(ctx.Query("page")),
		PageSize:      utils.Atoi64(ctx.Query("pageSize")),
		SortBy:        ctx.Query("sortBy"),
		SortType:      consts.SortType(ctx.Query("sortType")),
		ValidSortKeys: []string{"progress", "accuracyRate", "difficultyDegree", "costTime"},
	}
	apiPageInfo.Check()

	queryDto := &dto.TaskAssignReportQuery{
		TaskID:       taskId,
		AssignID:     assignId,
		ResourceID:   ctx.Query("resourceId"),
		ResourceType: utils.Atoi64(ctx.Query("resourceType")),
		StudentName:  ctx.Query("studentName"),
	}
	report, err := c.taskReportHandler.GetTaskReportSummaryDetail(ctx, queryDto, apiPageInfo)
	if err != nil {
		c.log.Error(ctx, "GetTaskReportDetail error:%v", err)
		response.Error(ctx, http.StatusBadRequest, response.Response{
			Code:    response.ERR_SYSTEM.Code,
			Message: "获取数据失败: " + err.Error(),
		})
		return
	}
	response.Success(ctx, report)
}

// GetAnswers 查询班级/小组作业答题结果
// 指定班级/小组
func (c *TaskReportController) GetAnswers(ctx *gin.Context) {
	// TODO 权限检查
	taskId := utils.Atoi64(ctx.Query("taskId"))
	assignId := utils.Atoi64(ctx.Query("assignId"))
	// 检查参数
	if taskId == 0 || assignId == 0 {
		c.log.Error(ctx, "taskId and assignId are required")
		response.ParamError(ctx, response.ERR_EMPTY_TASK_OR_ASSIGN)
		return
	}

	apiPageInfo := &consts.APIReqeustPageInfo{
		Page:          utils.Atoi64(ctx.Query("page")),
		PageSize:      utils.Atoi64(ctx.Query("pageSize")),
		SortBy:        ctx.Query("sortBy"),
		SortType:      consts.SortType(ctx.Query("sortType")),
		ValidSortKeys: []string{"answerCount", "incorrectCount"},
	}
	apiPageInfo.Check()

	query := &dto.TaskAssignAnswersQuery{
		TaskReportCommonQuery: dto.TaskReportCommonQuery{
			TaskID:       taskId,
			AssignID:     assignId,
			ResourceID:   ctx.Query("resourceId"),
			ResourceType: utils.Atoi64(ctx.Query("resourceType")),
			QuestionType: utils.Atoi64(ctx.Query("questionType")),
			Keyword:      ctx.Query("keyword"),
			AllQuestions: utils.AtoBool(ctx.Query("allQuestions")),
		},
	}

	answers, err := c.taskReportHandler.GetTaskAnswerReport(ctx, query, apiPageInfo)
	if err != nil {
		c.log.Error(ctx, "GetTaskAnswerReport error:%v", err)
		response.Error(ctx, http.StatusBadRequest, response.Response{
			Code:    response.ERR_SYSTEM.Code,
			Message: "获取作答数据失败: " + err.Error(),
		})
		return
	}
	response.Success(ctx, answers)
}

// ExportReport 导出作业报告，csv 格式
// 导出指定任务指定班级/小组的作业报告，可指定课程资源
func (c *TaskReportController) ExportReport(ctx *gin.Context) {
	// TODO 权限检查
	taskId := utils.Atoi64(ctx.Query("taskId"))
	assignId := utils.Atoi64(ctx.Query("assignId"))
	// 检查参数
	if taskId == 0 || assignId == 0 {
		c.log.Error(ctx, "taskId and assignId are required")
		response.ParamError(ctx, response.ERR_EMPTY_TASK_OR_ASSIGN)
		return
	}

	// 选择的导出字段
	var exportFields []string
	fields := ctx.Query("fields")
	if fields == "" {
		exportFields = consts.ExportFields
	} else {
		exportFields = strings.Split(fields, ",")
	}
	if !consts.IsExportFields(exportFields) {
		c.log.Error(ctx, "invalid export fields")
		response.ParamError(ctx, response.ERR_INVALID_EXPORT_FIELDS)
		return
	}

	sortBy, sortType := consts.SortHandler(ctx.Query("sortBy"), consts.SortType(ctx.Query("sortType")))
	reportName, report, err := c.taskReportHandler.ExportTaskReport(ctx, &dto.ExportTaskReportQuery{
		TaskID:       taskId,
		AssignID:     assignId,
		ResourceID:   ctx.Query("resourceId"),
		ResourceType: utils.Atoi64(ctx.Query("resourceType")),
		SortBy:       sortBy,
		SortType:     sortType,
		Fields:       exportFields,
	})

	if err != nil {
		c.log.Error(ctx, "ExportTaskReport error:%v", err)
		response.Error(ctx, http.StatusBadRequest, response.Response{
			Code:    response.ERR_SYSTEM.Code,
			Message: "导出作业报告失败: " + err.Error(),
		})
		return
	}

	// 使用 bytes.Buffer 构建 CSV 内容
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入 UTF-8 BOM
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	// 写入表头
	if err := writer.Write(report.Meta); err != nil {
		c.log.Error(ctx, "Write CSV header error:%v", err)
		response.Error(ctx, http.StatusInternalServerError, response.Response{
			Code:    response.ERR_SYSTEM.Code,
			Message: "导出作业报告失败: " + err.Error(),
		})
		return
	}

	// 写入数据行
	for _, row := range report.Data {
		if err := writer.Write(row); err != nil {
			c.log.Error(ctx, "Write CSV row error:%v", err)
			response.Error(ctx, http.StatusInternalServerError, response.Response{
				Code:    response.ERR_SYSTEM.Code,
				Message: "导出作业报告失败: " + err.Error(),
			})
			return
		}
	}

	// 刷新缓冲区
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.log.Error(ctx, "Flush CSV writer error:%v", err)
		response.Error(ctx, http.StatusInternalServerError, response.Response{
			Code:    response.ERR_SYSTEM.Code,
			Message: "导出作业报告失败: " + err.Error(),
		})
		return
	}

	// 设置响应头
	ctx.Header("Content-Type", "text/csv; charset=utf-8")

	// -- 文件名处理开始 --
	// 原始文件名 (包含中文)
	originalFilename := reportName + ".csv"
	// URL 编码后的文件名
	encodedName := url.QueryEscape(originalFilename)

	// 设置 Content-Disposition
	disposition := fmt.Sprintf(`attachment; filename="%s"`, encodedName)
	ctx.Header("Content-Disposition", disposition)
	// -- 文件名处理结束 --

	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Header("Pragma", "no-cache")
	ctx.Header("Expires", "0")

	// 写入响应
	if _, err := ctx.Writer.Write(buf.Bytes()); err != nil {
		c.log.Error(ctx, "Write response error:%v", err)
		response.Error(ctx, http.StatusInternalServerError, response.Response{
			Code:    response.ERR_SYSTEM.Code,
			Message: "导出作业报告失败: " + err.Error(),
		})
		return
	}

	ctx.Status(http.StatusOK)
}

// GetAnswerPanel 获取作业题目面板
func (c *TaskReportController) GetAnswerPanel(ctx *gin.Context) {
	// TODO 权限检查
	taskId := utils.Atoi64(ctx.Query("taskId"))
	assignId := utils.Atoi64(ctx.Query("assignId"))
	// 检查参数
	if taskId == 0 || assignId == 0 {
		c.log.Error(ctx, "taskId and assignId are required")
		response.ParamError(ctx, response.ERR_EMPTY_TASK_OR_ASSIGN)
		return
	}

	query := &dto.TaskReportCommonQuery{
		TaskID:       taskId,
		AssignID:     assignId,
		ResourceID:   ctx.Query("resourceId"),
		ResourceType: utils.Atoi64(ctx.Query("resourceType")),
		QuestionType: utils.Atoi64(ctx.Query("questionType")),
		Keyword:      ctx.Query("keyword"),
		AllQuestions: utils.AtoBool(ctx.Query("allQuestions")),
	}
	answerPanel, err := c.taskReportHandler.GetTaskAnswerPanel(ctx, query)
	if err != nil {
		c.log.Error(ctx, "GetTaskAnswerPanel error:%v", err)
		response.Error(ctx, http.StatusBadRequest, response.Response{
			Code:    response.ERR_SYSTEM.Code,
			Message: "获取题目面板数据失败: " + err.Error(),
		})
		return
	}
	response.Success(ctx, answerPanel)
}

// GetSuggestionTemplate 获取建议消息模板
func (c *TaskReportController) GetSuggestionTemplate(ctx *gin.Context) {
	template := consts.OfflineCommunicationContents
	response.Success(ctx, template)
}

// GetQuestions 查询作业提问记录
func (c *TaskReportController) GetQuestions(ctx *gin.Context) {

}

// GetKnowledgeQuestions 获取作业单个知识点的全部提问记录
func (c *TaskReportController) GetKnowledgeQuestions(ctx *gin.Context) {

}

// GetStudentAnswers 获取作业指定学生的全部答题结果
func (c *TaskReportController) GetStudentAnswers(ctx *gin.Context) {
	// TODO 权限检查
	taskId := utils.Atoi64(ctx.Query("taskId"))
	assignId := utils.Atoi64(ctx.Query("assignId"))
	studentId := utils.Atoi64(ctx.Query("studentId"))
	if studentId == 0 || taskId == 0 || assignId == 0 {
		c.log.Error(ctx, "taskId, assignId, studentId are required")
		response.ParamError(ctx, response.ERR_EMPTY_TASK_OR_ASSIGN)
		return
	}

	resourceId := ctx.Query("resourceId")
	resourceType := utils.Atoi64(ctx.Query("resourceType"))
	questionType := utils.Atoi64(ctx.Query("questionType"))
	keyword := ctx.Query("keyword")
	allQuestions := utils.AtoBool(ctx.Query("allQuestions"))

	// 检查题目类型是否存在
	if !consts.QuestionTypeExists(consts.QuestionType(questionType)) {
		c.log.Warn(ctx, "questionType is invalid")
		questionType = int64(consts.QUESTION_TYPE_ALL)
	}

	// 分页信息
	pageInfo := &consts.APIReqeustPageInfo{
		Page:     utils.Atoi64(ctx.Query("page")),
		PageSize: utils.Atoi64(ctx.Query("pageSize")),
		SortType: consts.SortType(ctx.Query("sortType")),
		SortBy:   ctx.Query("sortBy"),
	}
	pageInfo.Check()

	query := &dto.StudentTaskReportQuery{
		TaskReportCommonQuery: dto.TaskReportCommonQuery{
			TaskID:       taskId,
			AssignID:     assignId,
			ResourceID:   resourceId,
			ResourceType: resourceType,
			QuestionType: questionType,
			Keyword:      keyword,
			AllQuestions: allQuestions,
		},
	}

	report, err := c.taskReportHandler.StudentTaskReport(ctx, studentId, query, pageInfo)
	if err != nil {
		c.log.Error(ctx, "StudentTaskReport error:%v", err)
		response.SystemError(ctx)
		return
	}
	response.Success(ctx, report)
}

// HandleStudentReport 对学生作业报告的处理，目前只有点赞、提醒
func (c *TaskReportController) HandleStudentReport(ctx *gin.Context) {
	// 获取请求参数
	var req api.StudentReportHandleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "HandleStudentReport error:%v", err)
		response.ParamError(ctx)
		return
	}

	// 校验请求参数
	if err := req.Validate(); err != nil {
		c.log.Error(ctx, "HandleStudentReport error:%v", err)
		response.ParamError(ctx)
		return
	}

	teacherID, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		response.Forbidden(ctx)
		return
	}

	// 投递教师行为
	// TODO: 批量投递消费
	for _, studentID := range req.StudentIDs {
		err := c.producer.SendTeacherBehavior(ctx, &dto.TeacherBehaviorDTO{
			SchoolID:     uint64(schoolID),
			TeacherID:    uint64(teacherID),
			BehaviorType: req.BehaviorType,
			TaskID:       uint64(req.TaskID),
			AssignID:     uint64(req.AssignID),
			StudentID:    uint64(studentID),
			Context:      req.Content,
		})
		if err != nil {
			c.log.Error(ctx, "HandleStudentReport error:%v", err)
			response.Err(ctx, response.ERR_KAFKA)
			return
		}
	}
	response.Success(ctx, nil)
}

// GetStudentDetail 获取学生作业报告详情
func (c *TaskReportController) GetStudentDetail(ctx *gin.Context) {
	taskID := utils.Atoi64(ctx.Query("taskId"))
	assignID := utils.Atoi64(ctx.Query("assignId"))
	studentID := utils.Atoi64(ctx.Query("studentId"))
	schoolID := utils.Atoi64(ctx.Query("schoolId"))
	if taskID == 0 || assignID == 0 || studentID == 0 || schoolID == 0 {
		c.log.Error(ctx, "taskId, assignId, studentId, schoolId are required")
		response.ParamError(ctx)
		return
	}

	res, err := c.taskReportHandler.GetStudentDetail(ctx, schoolID, taskID, assignID, studentID)
	if err != nil {
		c.log.Error(ctx, "StudentTaskReport error:%v", err)
		response.Err(ctx, *err)
		return
	}
	response.Success(ctx, res)
}

// GetReportSetting 获取报告参数设置
func (c *TaskReportController) GetReportSetting(ctx *gin.Context) {
	// TODO 权限检查
	_, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		response.Forbidden(ctx)
		return
	}

	classId := utils.Atoi64(ctx.Query("classId"))
	subjectId := utils.Atoi64(ctx.Query("subject"))
	if classId == 0 || subjectId == 0 {
		response.ParamError(ctx, response.ERR_INVALID_CLASS_OR_SUBJECT)
		return
	}

	setting, err := c.taskReportHandler.GetTaskReportSetting(ctx, schoolID, classId, subjectId)
	if err != nil {
		c.log.Error(ctx, "GetTaskReportSetting error:%v", err)
		response.SystemError(ctx)
		return
	}
	response.Success(ctx, setting)
}

// UpdateReportSetting 更新或设置报告参数设置，只有学科教师有权限进行修改
func (c *TaskReportController) UpdateReportSetting(ctx *gin.Context) {
	// TODO 权限检查
	teacherID, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		response.Forbidden(ctx)
		return
	}

	var req *api.TaskReportSetting
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "UpdateReportSetting error:%v", err)
		response.ParamError(ctx, response.ERR_INVALID_TASK_REPORT_SETTING)
		return
	}

	if err := req.Validate(); err != nil {
		c.log.Error(ctx, "UpdateReportSetting error:%v", err)
		response.ParamError(ctx, response.ERR_INVALID_TASK_REPORT_SETTING)
		return
	}

	resp := c.taskReportHandler.UpdateTaskReportSetting(ctx, schoolID, teacherID, req)
	if resp != nil {
		c.log.Error(ctx, "UpdateTaskReportSetting error:%v", resp)
		response.Err(ctx, *resp)
		return
	}
	response.Success(ctx, nil)
}
