package controller_task

import (
	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/middleware"
	"gil_teacher/app/model/api"
	service "gil_teacher/app/service/task_service"

	"github.com/gin-gonic/gin"
)

// TempSelectionController 教师临时选择控制器
type TempSelectionController struct {
	log               *logger.ContextLogger
	service           *service.TempSelectionService
	teacherMiddleware *middleware.TeacherMiddleware
}

// NewTempSelectionController 创建教师临时选择控制器
func NewTempSelectionController(
	log *logger.ContextLogger,
	service *service.TempSelectionService,
	teacherMiddleware *middleware.TeacherMiddleware,
) *TempSelectionController {
	return &TempSelectionController{
		log:               log,
		service:           service,
		teacherMiddleware: teacherMiddleware,
	}
}

// CreateSelection 创建教师临时选择
func (c *TempSelectionController) CreateSelection(ctx *gin.Context) {
	var req api.CreateTempSelectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "绑定请求参数失败: %v", err)
		response.ParamError(ctx)
		return
	}

	if err := req.Validate(); err != nil {
		c.log.Error(ctx, "验证请求参数失败: %v", err)
		response.ParamError(ctx, *err)
		return
	}

	// 从中间件获取学校 ID 和老师 ID
	schoolID := c.teacherMiddleware.ExtractSchoolID(ctx)
	teacherID := c.teacherMiddleware.ExtractTeacherID(ctx)
	req.SchoolID = schoolID
	req.TeacherID = teacherID

	if err := c.service.CreateSelection(ctx, &req); err != nil {
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}

	response.Success(ctx, nil)
}

// DeleteSelections 删除教师临时选择
func (c *TempSelectionController) DeleteSelections(ctx *gin.Context) {
	var req api.DeleteTempSelectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "绑定请求参数失败: %v", err)
		response.ParamError(ctx)
		return
	}

	if len(req.QuestionIDs) == 0 {
		response.Success(ctx, nil)
		return
	}

	// 从中间件获取老师 ID
	teacherID := c.teacherMiddleware.ExtractTeacherID(ctx)

	// 目前只支持题目类型
	if err := c.service.DeleteSelections(ctx, consts.RESOURCE_TYPE_QUESTION, req.QuestionIDs, teacherID); err != nil {
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}

	response.Success(ctx, nil)
}

// ListSelections 查询教师临时选择列表
func (c *TempSelectionController) ListSelections(ctx *gin.Context) {
	// 从中间件获取老师 ID
	teacherID := c.teacherMiddleware.ExtractTeacherID(ctx)

	selections, err := c.service.ListSelections(ctx, teacherID)
	if err != nil {
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}

	resp := make(api.ListTempSelectionResponse, len(selections))
	for i, selection := range selections {
		resp[i] = api.TempSelectionItem{
			ID:           selection.ID,
			ResourceID:   selection.ResourceID,
			ResourceType: selection.ResourceType,
			CreateTime:   selection.CreateTime,
		}
	}

	response.Success(ctx, resp)
}
