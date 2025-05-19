package teacher

import (
	"gil_teacher/app/controller/http_server/response"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/middleware"
	"gil_teacher/app/model/itl"
	"gil_teacher/app/service/gil_internal/admin_service"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// TeacherController 教师控制器
type TeacherController struct {
	log               *logger.ContextLogger
	ucenterService    *admin_service.UcenterClient
	teacherMiddleware *middleware.TeacherMiddleware
}

// NewTeacherController 创建教师控制器实例
func NewTeacherController(log *logger.ContextLogger, ucenterService *admin_service.UcenterClient, teacherMiddleware *middleware.TeacherMiddleware) *TeacherController {
	return &TeacherController{
		log:               log,
		ucenterService:    ucenterService,
		teacherMiddleware: teacherMiddleware,
	}
}

// GetTeacherDetail 获取教师详细信息
func (tc *TeacherController) GetTeacherDetail(c *gin.Context) {
	// 从context中获取教师信息
	teacherDetail, ok := tc.teacherMiddleware.GetTeacherDetailFromContext(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	response.Success(c, teacherDetail)
}

// GetClassStudents 获取班级学生列表
func (tc *TeacherController) GetClassStudents(c *gin.Context) {
	// 获取 query 参数
	classIDString := c.Query("classIDs")
	if classIDString == "" {
		response.Success(c, nil)
		return
	}

	// 获取教师所在学校 id
	schoolID := tc.teacherMiddleware.ExtractSchoolID(c)

	classIDList := strings.Split(classIDString, ",")
	if len(classIDList) == 0 {
		response.Success(c, nil)
		return
	}

	// 获取教师任职的班级ID列表
	classIDs := tc.teacherMiddleware.ExtractTeacherClassIDs(c)

	// 校验班级ID是否在教师任职的班级ID列表中
	var classIDListInt64 []int64
	for _, classID := range classIDList {
		classIDInt64, err := strconv.ParseInt(classID, 10, 64)
		if err == nil && slices.Contains(classIDs, classIDInt64) {
			classIDListInt64 = append(classIDListInt64, classIDInt64)
		}
	}

	// 如果班级ID列表为空，则返回空列表
	if len(classIDListInt64) == 0 {
		response.Success(c, nil)
		return
	}

	// 调用 ucenterService 获取班级学生列表
	classStudents, err := tc.ucenterService.GetClassStudent(c, schoolID, classIDListInt64)
	if err != nil {
		response.Err(c, response.ERR_GIL_ADMIN)
		return
	}

	// 将 classStudents 转换为 dto.ClassInfo 列表
	res := []*itl.ClassInfo{}
	for _, class := range classStudents {
		res = append(res, class)
	}

	response.Success(c, res)
}
