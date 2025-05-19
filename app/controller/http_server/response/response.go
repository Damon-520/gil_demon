package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int    `json:"code"`    // 业务码
	Message string `json:"message"` // 提示信息
	Data    any    `json:"data"`    // 数据
}

// 按照约定，前端只处理 http status 为 401、403、500 的情况，此时会忽略 body 内容
// 业务类型的错误时，http status 为 200，此时会看 body 的内容
var (
	// 成功
	SUCCESS = Response{Code: 0, Message: "success"}

	// 错误码格式：前三位为 http 状态码，后四位为自定义错误码
	ERR_PARAM        = Response{Code: 4000001, Message: "请检查输入参数是否正确"}
	ERR_UNAUTHORIZED = Response{Code: 4010001, Message: "请先登录"}
	ERR_FORBIDDEN    = Response{Code: 4030001, Message: "未授权"}
	ERR_NOT_FOUND    = Response{Code: 4040001, Message: "资源不存在"}
	ERR_SYSTEM       = Response{Code: 5000001, Message: "系统开小差，请稍后再试"}

	// 底层模块错误，区分错误码，但错误信息不对外暴露过多
	ERR_POSTGRESQL   = Response{Code: 2000001, Message: "系统开小差，请稍后再试"}
	ERR_KAFKA        = Response{Code: 2000002, Message: "系统开小差，请稍后再试"}
	ERR_REDIS        = Response{Code: 2000003, Message: "系统开小差，请稍后再试"}
	ERR_GIL_QUESTION = Response{Code: 2000004, Message: "系统开小差，请稍后再试"}
	ERR_GIL_ADMIN    = Response{Code: 2000005, Message: "系统开小差，请稍后再试"}
	ERR_VOLC_AI      = Response{Code: 2000006, Message: "系统开小差，请稍后再试"}
	ERR_CLICKHOUSE   = Response{Code: 2000007, Message: "系统开小差，请稍后再试"}

	// 业务模块错误
	ERR_INVALID_PAGE                = Response{Code: 2001001, Message: "请选择正确的页码或每页数量"}
	ERR_SUBJECT                     = Response{Code: 2001002, Message: "请选择正确的学科"}
	ERR_BIZ_TREE                    = Response{Code: 2001003, Message: "请选择正确的教材或章节"}
	ERR_EMPTY_TASK_NAME             = Response{Code: 2001004, Message: "请填写任务名称"}
	ERR_INVALID_TASK_TYPE           = Response{Code: 2001005, Message: "请输入正确的任务类型"}
	ERR_EMPTY_RESOURCE              = Response{Code: 2001006, Message: "请选择布置的资源"}
	ERR_INVALID_RESOURCE            = Response{Code: 2001006, Message: "请输入正确的资源"}
	ERR_INVALID_RESOURCE_TYPE       = Response{Code: 2001007, Message: "请输入正确的资源类型"}
	ERR_EMPTY_STUDENT               = Response{Code: 2001008, Message: "请选择班级或学生"}
	ERR_INVALID_TIME                = Response{Code: 2001009, Message: "请输入正确的时间"}
	ERR_DUP_CLASS                   = Response{Code: 2001010, Message: "请不要选择重复的班级"}
	ERR_MULTI_CUSTOM_STU_GROUP      = Response{Code: 2001011, Message: "只允许一个自定义学生分组"}
	ERR_MIXED_GROUP                 = Response{Code: 2001012, Message: "班级和自定义学生不能同时存在"}
	ERR_EMPTY_TASK_NAME_COMMENT     = Response{Code: 2001013, Message: "请输入任务名称或老师留言"}
	ERR_INVALID_TASK                = Response{Code: 2001014, Message: "请选择正确的任务"}
	ERR_INVALID_ASSIGN              = Response{Code: 2001015, Message: "请选择正确的任务分配班级或群组"}
	ERR_INVALID_ORDER_BY            = Response{Code: 2001016, Message: "请选择正确的排序方式"}
	ERR_INVALID_QUESTION_SORT       = Response{Code: 2001017, Message: "排序字段只能为 createTime 或 useCount"}
	ERR_CQC                         = Response{Code: 2001018, Message: "请输入合规的文字内容"}
	ERR_EMPTY_TASK_OR_ASSIGN        = Response{Code: 2001019, Message: "请输入任务 ID 或任务分配 ID"}
	ERR_INVALID_EXPORT_FIELDS       = Response{Code: 2001020, Message: "不支持的导出字段"}
	ERR_INVALID_CLASS_OR_SUBJECT    = Response{Code: 2001021, Message: "班级或学科错误"}
	ERR_INVALID_TASK_REPORT_SETTING = Response{Code: 2001022, Message: "任务报告配置错误"}
	ERR_NO_PERMISSION_TO_MODIFY     = Response{Code: 2001023, Message: "无权限修改配置"}
	ERR_INVALID_STUDENT             = Response{Code: 2001024, Message: "请选择正确的学生"}

	// 课堂相关错误
	ERR_INVALID_CLASSROOM   = Response{Code: 2002001, Message: "请选择正确的课堂"}
	ERR_EMPTY_CLASSROOM     = Response{Code: 2002002, Message: "请选择课堂"}
	ERR_CLASSROOM_NOT_FOUND = Response{Code: 2002003, Message: "课堂不存在"}
	ERR_CLASSROOM_ID_ZERO   = Response{Code: 2002004, Message: "课堂ID不能为0"}
)

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS.Code,
		Message: SUCCESS.Message,
		Data:    data,
	})
}

// Err 错误响应
func Err(c *gin.Context, res Response) {
	// 总共七位，提取 res 前三位作为 http status
	statusCode := res.Code / 10000
	c.JSON(statusCode, res)
}

// SystemError 500 系统错误响应
func SystemError(c *gin.Context, errRes ...Response) {
	response := ERR_SYSTEM
	if len(errRes) != 0 {
		response = errRes[0]
	}
	c.JSON(http.StatusInternalServerError, response)
}

// ParamError 参数错误响应，业务类型错误，http status 返回 200
func ParamError(c *gin.Context, args ...Response) {
	response := ERR_PARAM
	if len(args) > 0 {
		response = args[0]
	}
	c.JSON(http.StatusOK, response)
}

// Unauthorized 401 未授权响应
func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, ERR_UNAUTHORIZED)
}

// Forbidden 403 禁止访问响应
func Forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, ERR_FORBIDDEN)
}

// Error 自定义错误响应
func Error(c *gin.Context, statusCode int, body Response) {
	c.JSON(statusCode, body)
}
