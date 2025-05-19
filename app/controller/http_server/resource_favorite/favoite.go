package resource_favorite

import (
	"gil_teacher/app/core/logger"
	"gil_teacher/app/middleware"
	"gil_teacher/app/service/resource_favorite"
	"net/http"

	"gil_teacher/app/controller/http_server/response"
	"gil_teacher/app/model/api"

	"github.com/gin-gonic/gin"
)

// ResourceFavoriteController 资源收藏控制器
type ResourceFavoriteController struct {
	favoriteService   *resource_favorite.ResourceFavoriteService
	teacherMiddleware *middleware.TeacherMiddleware
	log               *logger.ContextLogger
}

// NewResourceFavoriteController 创建资源收藏控制器
func NewResourceFavoriteController(
	favoriteService *resource_favorite.ResourceFavoriteService,
	teacherMiddleware *middleware.TeacherMiddleware,
	log *logger.ContextLogger,
) *ResourceFavoriteController {
	return &ResourceFavoriteController{
		favoriteService:   favoriteService,
		teacherMiddleware: teacherMiddleware,
		log:               log,
	}
}

// CreateFavorite 创建资源收藏
// @Summary 创建资源收藏
// @Description 创建资源收藏
// @Tags 资源收藏
// @Accept json
// @Produce json
// @Param request body api.CreateResourceFavoriteReq true "创建资源收藏请求"
// @Success 200 {object} response.Response
// @Router /api/v1/resource/favorite/create [post]
func (c *ResourceFavoriteController) CreateFavorite(ctx *gin.Context) {
	var req api.CreateResourceFavoriteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx)
		return
	}

	err := c.favoriteService.CreateFavorite(ctx, &req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, response.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	response.Success(ctx, nil)
}

// ListFavorites 获取资源收藏列表
// @Summary 获取资源收藏列表
// @Description 获取资源收藏列表
// @Tags 资源收藏
// @Accept json
// @Produce json
// @Param request body api.ListResourceFavoriteReq true "获取资源收藏列表请求"
// @Success 200 {object} response.Response{data=api.ListResourceFavoriteResp}
// @Router /api/v1/resource/favorite/list [post]
func (c *ResourceFavoriteController) ListFavorites(ctx *gin.Context) {
	// 获取教师ID
	teacherID, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		c.log.Error(ctx, "获取教师ID失败: %v", err)
		response.Unauthorized(ctx)
		return
	}

	var req api.ListResourceFavoriteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx)
		return
	}

	req.TeacherID = teacherID
	req.SchoolID = schoolID
	resp, err := c.favoriteService.ListFavorites(ctx, &req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, response.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	response.Success(ctx, resp)
}

// CancelFavorite 取消收藏
// @Summary 取消收藏
// @Description 取消收藏
// @Tags 资源收藏
// @Accept json
// @Produce json
// @Param request body api.CancelResourceFavoriteReq true "取消收藏请求"
// @Success 200 {object} response.Response
// @Router /api/v1/resource/favorite/cancel [post]
func (c *ResourceFavoriteController) CancelFavorite(ctx *gin.Context) {
	// 获取教师ID
	teacherID, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		c.log.Error(ctx, "获取教师ID失败: %v", err)
		response.Unauthorized(ctx)
		return
	}

	var req api.CancelResourceFavoriteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx)
		return
	}

	req.TeacherID = teacherID
	req.SchoolID = schoolID
	err = c.favoriteService.CancelFavorite(ctx, &req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, response.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	response.Success(ctx, nil)
}
