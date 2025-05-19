package response

import (
	"net/http"
	"time"

	"gil_teacher/app/third_party/errorx"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status       int         `json:"status"`
	Code         int         `json:"code"`
	Message      string      `json:"message"`
	ResponseTime int64       `json:"response_time"`
	Data         interface{} `json:"data"`
}

// NewResponse new a default response
func NewResponse() *Response {
	return &Response{
		Status:       200,
		Code:         0,
		Message:      "SUCCESS",
		ResponseTime: time.Now().Unix(),
	}
}

// Json return response by json
func (resp *Response) Json(c *gin.Context) {

	c.JSON(http.StatusOK, resp)
}

// JsonRaw return response by json
func (resp *Response) JsonRaw(c *gin.Context, v interface{}) {
	resp.Data = v

	c.JSON(http.StatusOK, resp)
}

// Error return response by errorx.Errors interface
func (resp *Response) Error(c *gin.Context, err errorx.Errors) {
	resp.Code = int(err.Code())
	resp.Message = err.Message()
	resp.Status = 400
	c.JSON(http.StatusOK, resp)
}

// ErrorRaw return response by error code and message
func (resp *Response) ErrorRaw(c *gin.Context, code int, msg string) {
	resp.Code = code
	resp.Message = msg
	resp.Status = 400
	c.JSON(http.StatusOK, resp)
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	resp := NewResponse()
	resp.Data = data
	c.JSON(http.StatusOK, resp)
}

// Error 返回错误响应
func Error(c *gin.Context, httpStatus int, message string) {
	resp := NewResponse()
	resp.Status = httpStatus
	resp.Code = httpStatus
	resp.Message = message
	c.JSON(httpStatus, resp)
}
