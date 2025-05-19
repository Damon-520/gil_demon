package api

import (
	"encoding/json"
	"gil_teacher/app/dao/file"
	"gil_teacher/app/utils"
	"time"
)

// UploadNotifyRequest 上传通知请求结构
type UploadNotifyRequest struct {
	ObjectKey        string          `json:"objectKey" binding:"required"` // OSS对象键
	FileName         string          `json:"fileName"`                     // 显示的文件名（可修改）
	OriginalFileName string          `json:"originalFileName"`             // 原始上传文件名
	UniqueFileID     string          `json:"uniqueFileId"`                 // 唯一文件ID
	FileByteSize     int64           `json:"fileByteSize"`                 // 文件大小(字节)
	FileHash         string          `json:"fileHash"`                     // 文件哈希
	FileScope        int             `json:"fileScope"`                    // 文件访问权限
	Metadata         json.RawMessage `json:"metadata"`                     // 元数据
	UserID           string          `json:"userId"`                       // 用户ID
	BusinessType     string          `json:"businessType"`                 // 业务类型
	PhaseEnum        int64           `json:"phaseEnum"`                    // 学段枚举
	SubjectEnum      int64           `json:"subjectEnum"`                  // 学科枚举
	SchoolID         int64           `json:"schoolId"`                     // 学校ID
}

// UploadNotifyResponse 上传通知响应结构
type UploadNotifyResponse struct {
	FileURL          string    `json:"fileUrl"`          // 文件访问URL
	FileName         string    `json:"fileName"`         // 文件名
	UniqueFileID     string    `json:"uniqueFileId"`     // 唯一文件ID
	OriginalFileName string    `json:"originalFileName"` // 原始文件名
	FileByteSize     int64     `json:"fileByteSize"`     // 文件大小(字节)
	FileHash         string    `json:"fileHash"`         // 文件哈希
	FileScope        int       `json:"fileScope"`        // 文件访问权限
	ObjectKey        string    `json:"objectKey"`        // OSS对象键
	PhaseEnum        int64     `json:"phaseEnum"`        // 学段枚举
	SubjectEnum      int64     `json:"subjectEnum"`      // 学科枚举
	UserID           string    `json:"userId"`           // 用户ID
	SchoolID         int64     `json:"schoolId"`         // 学校ID
	BusinessType     string    `json:"businessType"`     // 业务类型
	OSSBucket        string    `json:"ossBucket"`        // OSS存储桶
	CallbackTime     time.Time `json:"callbackTime"`     // 回调时间
	Status           string    `json:"status"`           // 状态
}

// NewUploadNotifyResponse 创建上传通知响应
func NewUploadNotifyResponse(
	fileURL, fileName, uniqueFileID, originalFileName string,
	fileByteSize int64, fileHash string, fileScope int,
	objectKey string, phaseEnum, subjectEnum int64,
	userID string, schoolID int64, businessType, ossBucket string,
	callbackTime time.Time,
) *UploadNotifyResponse {
	return &UploadNotifyResponse{
		FileURL:          fileURL,
		FileName:         fileName,
		UniqueFileID:     uniqueFileID,
		OriginalFileName: originalFileName,
		FileByteSize:     fileByteSize,
		FileHash:         fileHash,
		FileScope:        fileScope,
		ObjectKey:        objectKey,
		PhaseEnum:        phaseEnum,
		SubjectEnum:      subjectEnum,
		UserID:           userID,
		SchoolID:         schoolID,
		BusinessType:     businessType,
		OSSBucket:        ossBucket,
		CallbackTime:     callbackTime,
		Status:           "success",
	}
}

// UpdateFileNameRequest 更新文件名请求结构体
type UpdateFileNameRequest struct {
	ResourceID int64  `json:"resourceId" binding:"required"` // 资源ID
	FileName   string `json:"fileName" binding:"required"`   // 新文件名
}

// DeleteResourceRequest 删除资源请求结构体
type DeleteResourceRequest struct {
	ResourceID int64 `json:"resourceId" binding:"required"` // 资源ID
	UserID     int64 `json:"userId" binding:"required"`     // 用户ID
}

// UpdateFileNameResponse 更新文件名响应结构
type UpdateFileNameResponse struct {
	ResourceID   int64  `json:"resourceId"`   // 资源ID
	FileName     string `json:"fileName"`     // 新文件名
	OriginalName string `json:"originalName"` // 原始文件名
	OSSPath      string `json:"ossPath"`      // OSS路径
	FileType     string `json:"fileType"`     // 文件类型
	FileByteSize int64  `json:"fileByteSize"` // 文件大小
	UpdateTime   string `json:"updateTime"`   // 更新时间
	Status       string `json:"status"`       // 状态
}

// NewUpdateFileNameResponse 创建更新文件名响应
func NewUpdateFileNameResponse(
	resourceID int64,
	fileName string,
	originalName string,
	ossPath string,
	fileType string,
	fileByteSize int64,
	updateTime time.Time,
) *UpdateFileNameResponse {
	return &UpdateFileNameResponse{
		ResourceID:   resourceID,
		FileName:     fileName,
		OriginalName: originalName,
		OSSPath:      ossPath,
		FileType:     fileType,
		FileByteSize: fileByteSize,
		UpdateTime:   updateTime.Format(time.RFC3339),
		Status:       "success",
	}
}

// DeleteResourceResponse 删除资源响应结构
type DeleteResourceResponse struct {
	ResourceID int64  `json:"resourceId"` // 资源ID
	UserID     int64  `json:"userId"`     // 用户ID
	Status     string `json:"status"`     // 状态
	Message    string `json:"message"`    // 消息
}

// NewDeleteResourceResponse 创建删除资源响应
func NewDeleteResourceResponse(resourceID, userID int64) *DeleteResourceResponse {
	return &DeleteResourceResponse{
		ResourceID: resourceID,
		UserID:     userID,
		Status:     "success",
		Message:    "资源删除成功",
	}
}

// ResourceListItem 资源列表项结构
type ResourceListItem struct {
	ResourceID   int64  `json:"resourceId"`   // 资源ID
	UserID       int64  `json:"userId"`       // 用户ID
	SchoolID     int64  `json:"schoolId"`     // 学校ID
	FileName     string `json:"fileName"`     // 文件名
	FileType     string `json:"fileType"`     // 文件类型
	FileByteSize int64  `json:"fileByteSize"` // 文件大小
	FileScope    int    `json:"fileScope"`    // 文件访问权限
	OSSPath      string `json:"ossPath"`      // OSS路径
	Status       int64  `json:"status"`       // 状态
	CreateTime   string `json:"createTime"`   // 创建时间
	UpdateTime   string `json:"updateTime"`   // 更新时间
}

// QueryResourcesResponse 查询资源列表响应结构
type QueryResourcesResponse struct {
	Total      int64              `json:"total"`      // 总记录数
	Page       int                `json:"page"`       // 当前页码
	PageSize   int                `json:"pageSize"`   // 每页数量
	TotalPages int64              `json:"totalPages"` // 总页数
	Resources  []ResourceListItem `json:"resources"`  // 资源列表
	Params     interface{}        `json:"params"`     // 查询参数
}

// NewResourceListItem 创建资源列表项
func NewResourceListItem(resource *file.Resource) ResourceListItem {
	return ResourceListItem{
		ResourceID:   resource.ID,
		UserID:       resource.UserID,
		SchoolID:     resource.SchoolID,
		FileName:     resource.FileName,
		FileType:     resource.FileType,
		FileByteSize: resource.FileByteSize,
		FileScope:    resource.FileScope,
		OSSPath:      resource.OSSPath,
		Status:       resource.Status,
		CreateTime:   resource.CreateTime.Format(time.RFC3339),
		UpdateTime:   resource.UpdateTime.Format(time.RFC3339),
	}
}

// NewQueryResourcesResponse 创建查询资源列表响应
func NewQueryResourcesResponse(
	total int64,
	page, pageSize int,
	resources []*file.Resource,
	params interface{},
) *QueryResourcesResponse {
	// 转换资源列表
	resourceList := make([]ResourceListItem, len(resources))
	for i, resource := range resources {
		resourceList[i] = NewResourceListItem(resource)
	}

	return &QueryResourcesResponse{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: utils.I64Div(total+int64(pageSize)-1, int64(pageSize)),
		Resources:  resourceList,
		Params:     params,
	}
}

// GetPresignedPutURLRequest 获取预签名上传URL请求参数
type GetPresignedPutURLRequest struct {
	FileName     string `json:"fileName" binding:"required"`     // 文件名，必填
	FileByteSize int64  `json:"fileByteSize"`                    // 文件大小（字节）
	FileHash     string `json:"fileHash"`                        // 文件哈希值
	FileScope    int    `json:"fileScope"`                       // 访问权限
	BusinessType string `json:"businessType" binding:"required"` // 业务类型，必填
	UserId       string `json:"userId"`                          // 用户ID
	PhaseEnum    int64  `json:"phaseEnum" binding:"required"`    // 学段枚举，必填
	SubjectEnum  int64  `json:"subjectEnum" binding:"required"`  // 学科枚举，必填
	SchoolId     int64  `json:"schoolId"`                        // 学校ID
	CallbackURL  string `json:"callbackUrl"`                     // 回调URL
}

// NewGetPresignedPutURLRequest 创建获取预签名上传URL请求参数
func NewGetPresignedPutURLRequest() *GetPresignedPutURLRequest {
	return &GetPresignedPutURLRequest{}
}

// GetPresignedPutURLResponse 获取预签名URL响应结构
type GetPresignedPutURLResponse struct {
	PresignedURL     string `json:"presignedUrl"`     // 预签名URL
	ExpiresIn        int    `json:"expiresIn"`        // 过期时间(秒)
	ObjectKey        string `json:"objectKey"`        // OSS对象键
	OriginalFileName string `json:"originalFileName"` // 原始文件名
	UniqueFileID     string `json:"uniqueFileId"`     // 唯一文件ID
	CallbackURL      string `json:"callbackUrl"`      // 回调URL
	FileName         string `json:"fileName"`         // 文件名(与前端兼容)
	FileType         string `json:"fileType"`         // 文件类型
	FileScope        string `json:"fileScope"`        // 文件访问权限
	BusinessType     string `json:"businessType"`     // 业务类型
	PhaseEnum        int64  `json:"phaseEnum"`        // 学段枚举
	PhaseName        string `json:"phaseName"`        // 学段名称
	SubjectEnum      int64  `json:"subjectEnum"`      // 学科枚举
	SubjectName      string `json:"subjectName"`      // 学科名称
	OSSBucket        string `json:"ossBucket"`        // OSS存储桶
	UserID           string `json:"userId"`           // 用户ID
	SchoolID         int64  `json:"schoolId"`         // 学校ID
	ResourceID       int64  `json:"resourceId"`       // 资源ID
	ResourcePath     string `json:"resourcePath"`     // 资源路径
	GeneratedTime    string `json:"generatedTime"`    // 生成时间
}

// NewGetPresignedPutURLResponse 创建获取预签名URL响应
func NewGetPresignedPutURLResponse(
	presignedURL string,
	objectKey string,
	originalFileName string,
	uniqueFileID string,
	callbackURL string,
	fileType string,
	fileScope string,
	businessType string,
	phaseEnum int64,
	phaseName string,
	subjectEnum int64,
	subjectName string,
	ossBucket string,
	userID string,
	schoolID int64,
	resourceID int64,
	resourcePath string,
	createTime time.Time,
) *GetPresignedPutURLResponse {
	return &GetPresignedPutURLResponse{
		PresignedURL:     presignedURL,
		ExpiresIn:        1800, // 30分钟，单位秒
		ObjectKey:        objectKey,
		OriginalFileName: originalFileName,
		UniqueFileID:     uniqueFileID,
		CallbackURL:      callbackURL,
		FileName:         originalFileName, // 保持与前端兼容
		FileType:         fileType,
		FileScope:        fileScope,
		BusinessType:     businessType,
		PhaseEnum:        phaseEnum,
		PhaseName:        phaseName,
		SubjectEnum:      subjectEnum,
		SubjectName:      subjectName,
		OSSBucket:        ossBucket,
		UserID:           userID,
		SchoolID:         schoolID,
		ResourceID:       resourceID,
		ResourcePath:     resourcePath,
		GeneratedTime:    createTime.Format(time.RFC3339),
	}
}
