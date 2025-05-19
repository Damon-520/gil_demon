package api

import (
	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
)

// CreateTempSelectionRequest 教师临时选择请求
type CreateTempSelectionRequest struct {
	SchoolID     int64  `json:"-"`                               // 学校ID
	TeacherID    int64  `json:"-"`                               // 教师ID
	ResourceID   string `json:"resourceId" binding:"required"`   // 资源ID
	ResourceType int64  `json:"resourceType" binding:"required"` // 资源类型
}

func (c *CreateTempSelectionRequest) Validate() *response.Response {
	// 目前只支持题目类型
	if c.ResourceType != consts.RESOURCE_TYPE_QUESTION {
		return &response.ERR_INVALID_RESOURCE_TYPE
	}
	return nil
}

// DeleteTempSelectionRequest 删除教师临时选择请求
type DeleteTempSelectionRequest struct {
	QuestionIDs []string `json:"questionIds" binding:"required"` // 要删除的题目ID列表
}

// TempSelectionItem 教师临时选择
type TempSelectionItem struct {
	ID           int64  `json:"id"`           // 自增主键ID
	ResourceID   string `json:"resourceId"`   // 资源ID
	ResourceType int64  `json:"resourceType"` // 资源类型
	CreateTime   int64  `json:"createTime"`   // 记录创建时间
}

type ListTempSelectionResponse []TempSelectionItem
