package schedule

import (
	"encoding/json"
	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/middleware"
	"gil_teacher/app/model/api"
	"gil_teacher/app/service/schedule"
	"gil_teacher/app/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ScheduleController 课程表控制器
type ScheduleController struct {
	scheduleService   *schedule.ScheduleCacheService
	logger            *logger.ContextLogger
	teacherMiddleware *middleware.TeacherMiddleware
}

// NewScheduleController 初始化课程表功能控制器（注：教师端不创建课表，只处理课表数据）
func NewScheduleController(scheduleService *schedule.ScheduleCacheService, logger *logger.ContextLogger, teacherMiddleware *middleware.TeacherMiddleware) *ScheduleController {
	return &ScheduleController{
		scheduleService:   scheduleService,
		logger:            logger,
		teacherMiddleware: teacherMiddleware,
	}
}

// SaveScheduleToRedis 将课程表数据保存到Redis
func (c *ScheduleController) SaveScheduleToRedis(ctx *gin.Context) {
	// 解析请求参数
	var req api.SaveScheduleToRedisRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error(ctx, "解析请求参数失败: %v", err)
		response.ParamError(ctx)
		return
	}

	// 解析课程表数据
	var scheduleResp api.ScheduleResponse
	if err := json.Unmarshal([]byte(req.Data), &scheduleResp); err != nil {
		c.logger.Error(ctx, "解析课程表数据失败: %v", err)
		response.ParamError(ctx)
		return
	}

	// 准备请求参数
	fetchReq := &api.FetchScheduleRequest{
		TeacherID:    req.TeacherID,
		SchoolID:     req.SchoolID,
		SchoolYearID: req.SchoolYearID,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
	}

	// 保存到Redis
	if err := c.scheduleService.SaveScheduleToRedis(ctx, fetchReq, &scheduleResp); err != nil {
		c.logger.Error(ctx, "保存课程表数据到Redis失败: %v", err)
		response.Err(ctx, response.ERR_REDIS)
		return
	}

	response.Success(ctx, "保存课程表数据成功")
}

// StoreScheduleData 保存课程表数据
func (c *ScheduleController) StoreScheduleData(ctx *gin.Context) {
	// 从URL参数获取请求数据并转换类型
	teacherID, err := strconv.Atoi(ctx.Query("teacher_id"))
	if err != nil {
		c.logger.Error(ctx, "teacher_id参数类型错误: %v", err)
		response.ParamError(ctx)
		return
	}

	schoolID, err := strconv.Atoi(ctx.Query("school_id"))
	if err != nil {
		c.logger.Error(ctx, "school_id参数类型错误: %v", err)
		response.ParamError(ctx)
		return
	}

	schoolYearID, err := strconv.Atoi(ctx.Query("school_year_id"))
	if err != nil {
		c.logger.Error(ctx, "school_year_id参数类型错误: %v", err)
		response.ParamError(ctx)
		return
	}

	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	if startDate == "" || endDate == "" {
		c.logger.Error(ctx, "缺少必要参数 start_date 或 end_date")
		response.ParamError(ctx)
		return
	}

	// 创建请求对象
	req := &api.FetchScheduleRequest{
		TeacherID:    int64(teacherID),
		SchoolID:     int64(schoolID),
		SchoolYearID: int64(schoolYearID),
		StartDate:    startDate,
		EndDate:      endDate,
	}

	c.processScheduleData(ctx, req)
}

// StoreScheduleDataForScript 保存课程表数据（不需要认证，供脚本使用）
// 此方法已弃用，保留是为了兼容历史代码
func (c *ScheduleController) StoreScheduleDataForScript(ginCtx *gin.Context) {
	// 简单转发到StoreScheduleData方法，由其根据路径判断是否需要token
	c.logger.Info(ginCtx.Request.Context(), "调用弃用方法StoreScheduleDataForScript，转发到StoreScheduleData")
	c.StoreScheduleData(ginCtx)
}

// processScheduleData 处理课程表数据的获取和保存
func (c *ScheduleController) processScheduleData(ctx *gin.Context, req *api.FetchScheduleRequest) {
	// 从API获取课程表数据
	scheduleResp, err := c.scheduleService.FetchScheduleFromAPI(ctx, req)
	if err != nil {
		c.logger.Error(ctx, "从API获取课程表数据失败: %v", err)
		response.SystemError(ctx)
		return
	}

	// 计算每一天对应的日期
	dates, err := utils.CalculateWeekDates(req.StartDate)
	if err != nil {
		c.logger.Error(ctx, "计算周一到周日日期失败: %v", err)
		response.ParamError(ctx)
		return
	}

	// 将日期信息添加到响应中
	scheduleResp.Dates = dates

	// 将 schedule 的 key 从周几改为对应的日期
	scheduleWithDates := make(map[string][]api.Schedule)
	for weekday, date := range dates {
		if schedules, ok := scheduleResp.Schedule[weekday]; ok {
			scheduleWithDates[date] = schedules
		}
	}

	// 用于返回给客户端的响应，只包含当前周的数据
	clientResponse := &api.ScheduleResponse{
		Schedule: make(map[string][]api.Schedule),
		Dates:    dates,
	}

	// 只保留当前日期范围内的课程，用于返回给客户端
	validDates := make(map[string]bool)
	for _, date := range dates {
		validDates[date] = true
		if schedules, ok := scheduleWithDates[date]; ok && len(schedules) > 0 {
			// 只包含非空课程列表
			clientResponse.Schedule[date] = schedules
		}
	}

	// 设置完整数据用于保存到Redis
	scheduleResp.Schedule = scheduleWithDates

	// 保存完整课程表数据到Redis
	err = c.scheduleService.SaveScheduleToRedis(ctx, req, scheduleResp)
	if err != nil {
		c.logger.Error(ctx, "保存课程表数据到Redis失败: %v", err)
		response.Err(ctx, response.ERR_REDIS)
		return
	}

	// 保存教师ID和学校ID到Redis列表
	err = c.scheduleService.SaveTeacherToList(ctx, req.TeacherID, req.SchoolID)
	if err != nil {
		c.logger.Error(ctx, "保存教师ID和学校ID到Redis列表失败: %v", err)
		// 不影响主流程，只记录日志，不返回错误
	}

	// 过滤掉scheduleResp中的空数组日期
	for date, schedules := range scheduleResp.Schedule {
		if len(schedules) == 0 {
			c.logger.Debug(ctx, "过滤掉空数组日期: %s", date)
			delete(scheduleResp.Schedule, date)
		}
	}
	c.logger.Info(ctx, "过滤空数组后包含 %d 个日期的课程", len(scheduleResp.Schedule))

	// 返回成功响应
	response.Success(ctx, scheduleResp)
}

// IsTeachingNow 判断老师当前是否在上课
func (c *ScheduleController) IsTeachingNow(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	// 从参数中获取教师ID和学校ID
	teacherID, err := strconv.Atoi(ginCtx.Query("teacher_id"))
	if err != nil || teacherID <= 0 {
		response.ParamError(ginCtx)
		return
	}

	// 获取学校ID
	schoolID, err := strconv.Atoi(ginCtx.Query("school_id"))
	if err != nil || schoolID <= 0 {
		response.ParamError(ginCtx)
		return
	}

	// 获取日期和时间，默认使用当前时间
	checkDate := ginCtx.DefaultQuery("date", time.Now().Format(consts.TimeFormatDate))
	checkTime := ginCtx.DefaultQuery("time", time.Now().Format(consts.TimeFormatTimeOnly))

	// 解析日期和时间
	checkDateTime, err := time.Parse(consts.TimeFormatSecond, checkDate+" "+checkTime)
	if err != nil {
		c.logger.Error(ctx, "解析日期时间失败: %v", err)
		response.ParamError(ginCtx)
		return
	}

	// 检查是否在上课
	isTeaching, err := c.scheduleService.IsTeaching(ctx, int64(teacherID), int64(schoolID), checkDateTime)
	if err != nil {
		c.logger.Error(ctx, "检查上课状态失败: %v", err)
		response.SystemError(ginCtx)
		return
	}

	// 返回结果
	response.Success(ginCtx, gin.H{
		"is_teaching": isTeaching,
	})
}

// StoreDirectScheduleData 直接从Redis获取课程表数据
func (c *ScheduleController) StoreDirectScheduleData(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	var req api.DirectScheduleRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		c.logger.Error(ctx, "请求参数解析失败: %v", err)
		response.ParamError(ginCtx)
		return
	}

	// 验证必要参数
	if req.TeacherID <= 0 || req.SchoolID <= 0 || req.Date == "" {
		c.logger.Error(ctx, "缺少必要参数 teacher_id, school_id 或 date")
		response.ParamError(ginCtx)
		return
	}

	// 直接从Redis获取课程表数据
	scheduleData, err := c.scheduleService.GetScheduleFromRedis(ctx, req.TeacherID, req.SchoolID, req.Date)
	if err != nil {
		c.logger.Error(ctx, "从Redis获取课程表数据失败: %v", err)
		response.Err(ginCtx, response.ERR_REDIS)
		return
	}

	if scheduleData == nil || len(scheduleData) == 0 {
		response.Success(ginCtx, api.ScheduleDayResponse{
			Date:     req.Date,
			Schedule: []api.Schedule{},
		})
		return
	}

	// 返回成功响应
	response.Success(ginCtx, api.ScheduleDayResponse{
		Date:     req.Date,
		Schedule: scheduleData,
	})
}

// GetTeacherList 获取教师列表 脚本专用
func (c *ScheduleController) GetTeacherList(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	// 从Redis获取教师列表
	teacherList, err := c.scheduleService.GetTeacherList(ctx)
	if err != nil {
		c.logger.Error(ctx, "获取教师列表失败: %v", err)
		response.Err(ginCtx, response.ERR_REDIS)
		return
	}

	if len(teacherList) == 0 {
		response.Success(ginCtx, api.TeacherListResponse{
			Count:       0,
			TeacherList: []api.TeacherInfo{},
		})
		return
	}

	// 将map转换为TeacherInfo数组
	var teachers []api.TeacherInfo
	for _, item := range teacherList {
		teachers = append(teachers, api.TeacherInfo{
			TeacherID: item["teacher_id"],
			SchoolID:  item["school_id"],
		})
	}

	// 返回成功响应
	response.Success(ginCtx, api.TeacherListResponse{
		Count:       int64(len(teachers)),
		TeacherList: teachers,
	})
}

// CheckTeacherTeachingStatus 检查教师当前教学状态（详细版本）
func (c *ScheduleController) CheckTeacherTeachingStatus(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	// 从参数中获取教师ID和学校ID
	teacherID, err := strconv.Atoi(ginCtx.Query("teacher_id"))
	if err != nil || teacherID <= 0 {
		response.ParamError(ginCtx)
		return
	}

	// 获取学校ID
	schoolID, err := strconv.Atoi(ginCtx.Query("school_id"))
	if err != nil || schoolID <= 0 {
		response.ParamError(ginCtx)
		return
	}

	// 获取学年ID
	schoolYearID, err := strconv.Atoi(ginCtx.DefaultQuery("school_year_id", "0"))
	if err != nil {
		schoolYearID = 0 // 默认值
	}

	// 获取日期和时间参数
	startDate := ginCtx.DefaultQuery("start_date", time.Now().Format(consts.TimeFormatDate))
	endDate := ginCtx.DefaultQuery("end_date", time.Now().Format(consts.TimeFormatDate))
	checkDate := ginCtx.DefaultQuery("date", time.Now().Format(consts.TimeFormatDate))
	checkTime := ginCtx.DefaultQuery("time", time.Now().Format(consts.TimeFormatTimeOnly))

	// 解析日期
	checkDateTime, err := time.Parse(consts.TimeFormatSecond, checkDate+" "+checkTime)
	if err != nil {
		c.logger.Error(ctx, "解析日期时间失败: %v", err)
		response.ParamError(ginCtx)
		return
	}

	// 获取星期几（1-7，周一到周日）
	weekdayStr := ginCtx.DefaultQuery("weekday", "0")
	weekday, err := strconv.Atoi(weekdayStr)
	if err != nil || weekday < 1 || weekday > 7 {
		// 如果没有提供weekday或无效，则从checkDateTime计算
		weekday = int(checkDateTime.Weekday())
		if weekday == 0 {
			weekday = 7 // 将周日从0调整为7
		}
	}

	// 创建请求对象
	req := &api.CheckTeachingStatusRequest{
		TeacherID:    int64(teacherID),
		SchoolID:     int64(schoolID),
		SchoolYearID: int64(schoolYearID),
		StartDate:    startDate,
		EndDate:      endDate,
		Now:          checkDateTime.Unix(),
		CheckTime:    checkTime,
		Weekday:      int64(weekday),
	}

	// 调用服务方法获取教学状态
	result, err := c.scheduleService.CheckTeachingStatus(ctx, req)
	if err != nil {
		c.logger.Error(ctx, "检查教师教学状态失败: %v", err)
		response.SystemError(ginCtx)
		return
	}

	// 返回结果
	response.Success(ginCtx, result)
}

// extractAuthToken 从请求中提取认证令牌
func (c *ScheduleController) extractAuthToken(ginCtx *gin.Context) string {
	// 首先尝试从URL参数获取token
	token := ginCtx.Query("token")
	if token != "" {
		return token
	}

	// 如果URL参数中没有token，尝试从Authorization头获取
	authHeader := ginCtx.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// convertToSubjectInfos 将科目ID列表转换为SubjectInfo结构体数组
// 如果传入的科目列表为空，则返回默认的科目列表（语文和数学）
func (c *ScheduleController) convertToSubjectInfos(subjectIDs []int64) []api.SubjectInfo {
	// 如果教师科目为空，添加默认科目
	if len(subjectIDs) == 0 {
		return []api.SubjectInfo{
			{SubjectKey: consts.SUBJECT_CHINESE, SubjectName: consts.SUBJECT_CHINESE_NAME},
			{SubjectKey: consts.SUBJECT_MATH, SubjectName: consts.SUBJECT_MATH_NAME},
		}
	}

	// 正常情况下转换教师的科目
	subjectInfos := make([]api.SubjectInfo, 0, len(subjectIDs))
	for _, subjectID := range subjectIDs {
		subjectName := consts.SubjectNameMap[subjectID]
		if subjectName != "" {
			subjectInfos = append(subjectInfos, api.SubjectInfo{
				SubjectKey:  subjectID,
				SubjectName: subjectName,
			})
		}
	}

	// 如果转换后仍然为空（可能是因为科目ID无效），返回默认科目
	if len(subjectInfos) == 0 {
		return []api.SubjectInfo{
			{SubjectKey: consts.SUBJECT_CHINESE, SubjectName: consts.SUBJECT_CHINESE_NAME},
			{SubjectKey: consts.SUBJECT_MATH, SubjectName: consts.SUBJECT_MATH_NAME},
		}
	}

	return subjectInfos
}

// GetClassroomStatus 获取班级名称和上下课状态
func (c *ScheduleController) GetClassroomStatus(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	// 解析请求参数
	var req api.ClassroomStatusRequest
	if err := ginCtx.ShouldBindQuery(&req); err != nil {
		c.logger.Error(ctx, "解析请求参数失败: %v", err)
		response.ParamError(ginCtx)
		return
	}

	// 验证请求参数
	if err := req.Validate(); err != nil {
		c.logger.Error(ctx, "验证请求参数失败: %v", err)
		response.ParamError(ginCtx, response.ERR_CLASSROOM_ID_ZERO)
		return
	}

	teacherID := c.teacherMiddleware.ExtractTeacherID(ginCtx)
	schoolID := c.teacherMiddleware.ExtractSchoolID(ginCtx)
	teacherSubjects := c.teacherMiddleware.ExtractTeacherSubjects(ginCtx)

	c.logger.Debug(ctx, "获取到的教师科目: %v", teacherSubjects)

	// 获取当前日期
	now := time.Now().In(consts.LocationShanghai)
	currentDate := now.Format(consts.TimeFormatDate)

	// 通过教室ID获取课程信息
	schedule, err := c.scheduleService.GetScheduleByClassroomID(ctx, req.ClassroomID, teacherID, schoolID, currentDate)
	if err != nil {
		c.logger.Error(ctx, "获取课程信息失败: %v", err)
		response.SystemError(ginCtx)
		return
	}

	if schedule == nil {
		c.logger.Warn(ctx, "未找到课堂信息: classroomId=%d", req.ClassroomID)
		// 返回空响应，但仍然包含classroomID
		response.Success(ginCtx, api.ClassroomStatusResponse{
			ClassroomID: req.ClassroomID,
			IsInClass:   false,
			GradeName:   consts.ClassroomNotFoundGradeName,
			ClassName:   consts.ClassroomNotFoundClassName,
			Subjects:    c.convertToSubjectInfos(nil), // 使用默认科目
			Subject:     "",                           // 空学科内容
		})
		return
	}

	// 日志记录找到的课程信息
	c.logger.Debug(ctx, "找到课堂信息: classroomId=%d, gradeName=%s, className=%s, isInClass=%v",
		req.ClassroomID, schedule.GradeName, schedule.ClassName, schedule.IsInClass)

	// 确保班级名称和年级名称不为空
	gradeName := schedule.GradeName
	if gradeName == "" {
		gradeName = consts.DefaultGradeName
		c.logger.Warn(ctx, "课程信息中年级名称为空: classroomId=%d", req.ClassroomID)
	}

	className := schedule.ClassName
	if className == "" {
		className = consts.DefaultClassName
		c.logger.Warn(ctx, "课程信息中班级名称为空: classroomId=%d", req.ClassroomID)
	}

	// 使用辅助方法转换科目信息
	subjectInfos := c.convertToSubjectInfos(teacherSubjects)
	c.logger.Debug(ctx, "转换后的科目信息: %+v", subjectInfos)

	// 返回课堂状态
	resp := api.ClassroomStatusResponse{
		ClassroomID: req.ClassroomID,
		GradeName:   gradeName,
		ClassName:   className,
		IsInClass:   schedule.IsInClass,
		Subject:     schedule.ClassScheduleCourse,
		Subjects:    subjectInfos,
	}
	c.logger.Debug(ctx, "返回的响应数据: %+v", resp)

	// 返回响应
	response.Success(ginCtx, resp)
}
