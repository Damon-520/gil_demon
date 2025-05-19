package upload

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gil_teacher/app/conf"
	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/dao/file"
	"gil_teacher/app/model/api"
	"gil_teacher/app/third_party/response"
	"gil_teacher/app/utils"
	ossutil "gil_teacher/app/utils/oss"

	"github.com/gin-gonic/gin"
)

// 存储OSS回调相关信息的结构体
type CallbackRecord struct {
	ObjectKey    string    // 对象键名
	FileName     string    // 文件名
	FileByteSize int64     // 文件大小(字节)
	FileType     string    // 文件类型
	FileURL      string    // 文件URL
	UserId       string    // 用户ID
	BusinessType string    // 业务类型
	UploadTime   time.Time // 上传时间
	CallbackTime time.Time // 回调时间
	Status       string    // 状态: success, failed, pending
	Remark       string    // 备注
}

// UploadController 处理文件上传相关逻辑
type UploadController struct {
	ossClient      *ossutil.OSSClient         // OSS客户端
	resourceDAO    *file.ResourceDAO          // 资源记录DAO
	callbacksMutex sync.RWMutex               // 回调记录锁
	callbacks      map[string]*CallbackRecord // 回调记录 (key: objectKey)
	logger         *logger.ContextLogger      // 日志记录器
	config         *conf.Bootstrap            // 配置信息
}

// NewUploadController 创建上传控制器
func NewUploadController(
	ossClient *ossutil.OSSClient,
	resourceDAO *file.ResourceDAO,
	logger *logger.ContextLogger,
	config *conf.Bootstrap,
) *UploadController {
	return &UploadController{
		ossClient:   ossClient,
		resourceDAO: resourceDAO,
		callbacks:   make(map[string]*CallbackRecord),
		logger:      logger,
		config:      config,
	}
}

// 全局命令行参数（可以通过配置文件或环境变量设置）
var (
	region     string // OSS区域
	bucketName string // OSS存储桶名称
	objectName string // OSS对象名称前缀
)

// GetPresignedPutURL 获取OSS预签名上传URL
func (uc *UploadController) GetPresignedPutURL(c *gin.Context) {
	// 获取参数
	requestParams := api.NewGetPresignedPutURLRequest()

	// 绑定参数
	if err := c.ShouldBindJSON(requestParams); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err))
		return
	}

	// 验证参数
	if requestParams.FileName == "" {
		response.Error(c, http.StatusBadRequest, "文件名不能为空")
		return
	}

	if requestParams.PhaseEnum <= 0 {
		response.Error(c, http.StatusBadRequest, "学段枚举不能为空")
		return
	}

	if requestParams.SubjectEnum <= 0 {
		response.Error(c, http.StatusBadRequest, "学科枚举不能为空")
		return
	}

	// 处理访问权限
	fileScope := requestParams.FileScope
	if fileScope <= 0 {
		fileScope = consts.FILE_SCOPE_PRIVATE // 使用私有作为默认值
	}
	fileScopeStr := consts.FileScopeNameMap[fileScope]

	// 获取学段和学科名称
	phaseName := consts.PhaseNameMap[requestParams.PhaseEnum]
	subjectName := consts.SubjectNameMap[requestParams.SubjectEnum]

	// 从userId字符串转换为int64
	var userID int64 = 0
	if requestParams.UserId != "" {
		userID, _ = strconv.ParseInt(requestParams.UserId, 10, 64)
	}

	// 如果schoolId未提供，默认为0
	schoolID := requestParams.SchoolId
	if schoolID <= 0 {
		schoolID = 0 // 默认值
	}

	// 生成唯一的OSS对象键
	createTime := time.Now()
	randomStr := strconv.FormatInt(rand.Int63(), 10)[:6]
	uniqueFileID := utils.GenerateUniqueFileName(userID, createTime, randomStr)

	// 获取文件扩展名
	fileExt := filepath.Ext(requestParams.FileName)
	originalFileName := requestParams.FileName

	// 构造存储路径
	basePath := fmt.Sprintf("%s/%d/%d", uc.ossClient.GetBasePath(), requestParams.PhaseEnum, requestParams.SubjectEnum)
	objectKey := utils.GenerateObjectKey(basePath, uniqueFileID, fileExt)

	// 获取OSS bucket
	ossBucketName := uc.ossClient.GetBucketName()

	// 生成预签名URL
	ctx := c.Request.Context()
	expireTime := createTime.Add(30 * time.Minute) // 30分钟有效期
	presignedURL, err := uc.ossClient.GetPresignedPutURL(objectKey, expireTime)
	if err != nil {
		uc.logger.Error(ctx, "获取预签名URL失败: %v", err)
		response.Error(c, http.StatusInternalServerError, fmt.Sprintf("获取预签名URL失败: %v", err))
		return
	}

	// 生成文件访问URL
	fileURL := uc.ossClient.GetObjectURL(objectKey)

	// 获取文件类型
	fileType := strings.TrimPrefix(fileExt, ".")
	if fileType == "" {
		response.Error(c, http.StatusBadRequest, "不支持的文件类型")
		return
	}

	// 设置资源状态
	status := int64(1) // 1:审核中, 2:已通过, 3:未通过

	// 创建资源记录
	resource := &file.Resource{
		UserID:       userID,
		SchoolID:     schoolID,
		FileName:     originalFileName, // 保存原始文件名
		OSSPath:      fileURL,
		OSSBucket:    ossBucketName,
		FileType:     fileType,
		FileByteSize: 0, // 在上传完成后会更新实际大小
		FileHash:     requestParams.FileHash,
		FileScope:    fileScope,
		Metadata:     "{}",
		Status:       status,
		CreateTime:   &createTime,
		UpdateTime:   &createTime,
	}

	err = uc.resourceDAO.CreateResource(ctx, resource)
	if err != nil {
		uc.logger.Error(ctx, "创建资源记录失败: %v", err)
	}

	// 计算回调URL
	callbackURL := requestParams.CallbackURL
	if callbackURL == "" {
		// 如果未提供回调URL，使用配置中的默认回调地址
		callbackURL = fmt.Sprintf("%s%s",
			uc.config.Upload.Callback.Host,
			uc.config.Upload.Callback.Path,
		)
	}

	// 记录日志
	uc.logger.Info(ctx, "预签名上传URL已生成: 文件=%s, 用户=%s, 业务类型=%s, 存储路径=%s, 唯一ID=%s",
		requestParams.FileName, requestParams.UserId, requestParams.BusinessType, objectKey, uniqueFileID)

	// 返回响应
	response.Success(c, api.NewGetPresignedPutURLResponse(
		presignedURL,
		objectKey,
		originalFileName,
		uniqueFileID,
		callbackURL,
		fileType,
		fileScopeStr,
		requestParams.BusinessType,
		requestParams.PhaseEnum,
		phaseName,
		requestParams.SubjectEnum,
		subjectName,
		ossBucketName,
		requestParams.UserId,
		schoolID,
		resource.ID,
		fileURL,
		createTime,
	))
}

// NotifyUploadComplete 通知预签名URL上传完成
func (uc *UploadController) NotifyUploadComplete(c *gin.Context) {
	// 解析请求数据
	var notifyData api.UploadNotifyRequest

	if err := c.ShouldBindJSON(&notifyData); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("解析通知数据失败: %v", err))
		return
	}

	ctx := c.Request.Context()
	uc.logger.Info(ctx, "收到通知数据: %+v", notifyData)

	// 检查必填参数
	if notifyData.PhaseEnum <= 0 {
		response.Error(c, http.StatusBadRequest, "学段枚举（phaseEnum）不能为空")
		return
	}

	if notifyData.SubjectEnum <= 0 {
		response.Error(c, http.StatusBadRequest, "学科枚举（subjectEnum）不能为空")
		return
	}

	// 从objectKey中获取文件名
	fileName := filepath.Base(notifyData.ObjectKey)
	uc.logger.Info(ctx, "使用objectKey中的文件名: %s", fileName)

	// 生成文件访问URL
	fileURL := uc.ossClient.GetObjectURL(notifyData.ObjectKey)

	// 从userId字符串转换为int64
	var userID int64 = 0
	if notifyData.UserID != "" {
		userID, _ = strconv.ParseInt(notifyData.UserID, 10, 64)
	}

	// 如果schoolId未提供，默认为0
	schoolID := notifyData.SchoolID
	if schoolID <= 0 {
		schoolID = 0 // 默认值
	}

	// 设置状态为正在审核 (1)
	status := int64(1) // 1:审核中, 2:已通过, 3:未通过

	// 获取文件类型
	fileType := filepath.Ext(fileName)
	if fileType != "" && fileType[0] == '.' {
		fileType = fileType[1:] // 移除开头的点
	}

	// 处理元数据，确保它是有效的JSON
	metadata := "{}"
	if len(notifyData.Metadata) > 0 {
		metadata = string(notifyData.Metadata)
	}

	// 获取OSS bucket
	ossBucketName := uc.ossClient.GetBucketName()

	// 创建资源记录
	ctx = context.Background()
	callbackTime := time.Now()

	// 创建资源记录
	resource := &file.Resource{
		UserID:       userID,
		SchoolID:     schoolID,
		FileName:     fileName, // 使用可能修改过的文件名
		OSSPath:      fileURL,
		OSSBucket:    ossBucketName,
		FileType:     fileType,
		FileByteSize: notifyData.FileByteSize,
		FileHash:     notifyData.FileHash,
		FileScope:    notifyData.FileScope,
		Metadata:     metadata,
		Status:       status,
		CreateTime:   &callbackTime,
		UpdateTime:   &callbackTime,
	}

	log.Printf("准备保存资源记录: %+v", resource)
	log.Printf("资源文件信息: 大小=%d字节, 类型=%s, 哈希=%s, 状态=%d",
		notifyData.FileByteSize, fileType, notifyData.FileHash, status)

	// 先查询是否存在记录
	existingResource, err := uc.resourceDAO.GetResourceByObjectKey(ctx, notifyData.ObjectKey)
	if err != nil {
		log.Printf("查询资源记录发生错误: %v", err)

		// 如果不存在，创建新记录
		log.Printf("尝试创建新的资源记录...")
		err = uc.resourceDAO.CreateResource(ctx, resource)
		if err != nil {
			log.Printf("创建资源记录失败: %v", err)
		} else {
			log.Printf("成功创建资源记录")
		}
	} else {
		// 更新现有记录
		log.Printf("找到已存在的资源记录: %+v", existingResource)
		existingResource.FileName = fileName // 更新为可能修改过的文件名
		existingResource.OSSPath = fileURL
		existingResource.FileByteSize = notifyData.FileByteSize
		existingResource.FileType = fileType
		existingResource.FileHash = notifyData.FileHash
		existingResource.FileScope = notifyData.FileScope
		existingResource.Metadata = metadata
		existingResource.UpdateTime = &callbackTime

		err = uc.resourceDAO.UpdateResource(ctx, existingResource)
		if err != nil {
			log.Printf("更新资源记录失败: %v", err)
		} else {
			log.Printf("成功更新资源记录")
		}
	}

	// 记录日志
	log.Printf("文件上传成功(预签名通知): 文件=%s, 大小=%d, 用户=%s, 业务类型=%s",
		fileName, notifyData.FileByteSize, notifyData.UserID, notifyData.BusinessType)

	// 返回成功响应
	response.Success(c, api.NewUploadNotifyResponse(
		fileURL,
		fileName,
		notifyData.UniqueFileID,
		notifyData.OriginalFileName,
		notifyData.FileByteSize,
		notifyData.FileHash,
		notifyData.FileScope,
		notifyData.ObjectKey,
		notifyData.PhaseEnum,
		notifyData.SubjectEnum,
		notifyData.UserID,
		schoolID,
		notifyData.BusinessType,
		ossBucketName,
		callbackTime,
	))
}

// UpdateFileName 更新文件名
func (uc *UploadController) UpdateFileName(c *gin.Context) {
	// 解析请求数据
	var requestData api.UpdateFileNameRequest

	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err))
		return
	}

	// 统一验证参数
	params := map[string]interface{}{
		"resourceId": requestData.ResourceID,
		"fileName":   requestData.FileName,
	}
	if !utils.ValidateParams(c, params) {
		return
	}

	// 获取资源记录
	ctx := c.Request.Context()
	resource, err := uc.resourceDAO.GetResource(ctx, requestData.ResourceID)
	if err != nil {
		log.Printf("获取资源记录失败: %v", err)
		response.Error(c, http.StatusNotFound, "未找到资源记录")
		return
	}

	// 获取当前文件类型
	currentExt := filepath.Ext(resource.FileName)

	// 确保新文件名具有相同的扩展名
	newExt := filepath.Ext(requestData.FileName)
	if newExt == "" && currentExt != "" {
		// 如果新文件名没有扩展名但原始文件有，则添加原始扩展名
		requestData.FileName = requestData.FileName + currentExt
	} else if newExt != "" && newExt != currentExt {
		// 如果新文件名有扩展名但与原始不同，提示用户
		log.Printf("警告：新文件名扩展名(%s)与原始扩展名(%s)不同，继续使用新扩展名", newExt, currentExt)
	}

	// 更新文件名
	oldFileName := resource.FileName
	resource.FileName = requestData.FileName
	updateTime := time.Now()
	resource.UpdateTime = &updateTime

	// 更新资源记录
	err = uc.resourceDAO.UpdateResource(ctx, resource)
	if err != nil {
		log.Printf("更新资源记录失败: %v", err)
		response.Error(c, http.StatusInternalServerError, "更新资源记录失败")
		return
	}

	log.Printf("文件名更新成功: ID=%d, 旧名称=%s, 新名称=%s",
		requestData.ResourceID, oldFileName, requestData.FileName)

	// 返回成功响应
	response.Success(c, api.NewUpdateFileNameResponse(
		resource.ID,
		resource.FileName,
		oldFileName,
		resource.OSSPath,
		resource.FileType,
		resource.FileByteSize,
		updateTime,
	))
}

// QueryResources 查询资源列表
func (uc *UploadController) QueryResources(c *gin.Context) {
	// 解析分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", strconv.Itoa(consts.API_DEFAULT_PAGE_SIZE))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, response.Response{
			Code:    400,
			Message: "无效的页码",
		})
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > consts.API_MAX_PAGE_SIZE {
		c.JSON(http.StatusBadRequest, response.Response{
			Code:    400,
			Message: "无效的每页数量",
		})
		return
	}

	offset := (page - 1) * pageSize

	// 准备查询参数
	params := make(map[string]interface{})

	// 用户ID查询
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID := utils.ParseInt64(userIDStr)
		if userID <= 0 {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("无效的用户ID: %s", userIDStr))
			return
		}
		params["userId"] = userID
	}

	// 学校ID查询
	if schoolIDStr := c.Query("school_id"); schoolIDStr != "" {
		schoolID := utils.ParseInt64(schoolIDStr)
		if schoolID <= 0 {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("无效的学校ID: %s", schoolIDStr))
			return
		}
		params["schoolId"] = schoolID
	}

	// 文件名查询
	if fileName := c.Query("file_name"); fileName != "" {
		params["fileName"] = fileName
	}

	// 文件类型查询
	if fileType := c.Query("file_type"); fileType != "" {
		// 验证文件类型是否有效
		if !isValidFileType(fileType) {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("无效的文件类型: %s", fileType))
			return
		}
		params["fileType"] = fileType
	}

	// 状态查询
	if statusStr := c.Query("status"); statusStr != "" {
		status := utils.ParseInt64(statusStr)
		if status <= 0 {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("无效的状态值: %s", statusStr))
			return
		}
		params["status"] = status
	}

	// 文件访问权限查询
	if fileScopeStr := c.Query("file_scope"); fileScopeStr != "" {
		fileScope := utils.ParseInt(fileScopeStr)
		if fileScope < 0 {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("无效的文件访问权限: %s", fileScopeStr))
			return
		}
		params["fileScope"] = fileScope
	}

	// 时间范围查询
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := utils.ParseTime(startTimeStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("无效的开始时间格式: %s", startTimeStr))
			return
		}
		params["startTime"] = startTime
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := utils.ParseTime(endTimeStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("无效的结束时间格式: %s", endTimeStr))
			return
		}
		params["endTime"] = endTime
	}

	// 排序方式
	if orderBy := c.Query("order_by"); orderBy != "" {
		// 验证排序字段是否有效
		if !isValidOrderByField(orderBy) {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("无效的排序字段: %s", orderBy))
			return
		}
		params["orderBy"] = orderBy
		if orderDir := c.Query("order_dir"); orderDir != "" {
			orderDir = strings.ToUpper(orderDir)
			if orderDir != "ASC" && orderDir != "DESC" {
				response.Error(c, http.StatusBadRequest, fmt.Sprintf("无效的排序方向: %s", orderDir))
				return
			}
			params["orderDir"] = orderDir
		}
	}

	// 执行查询
	ctx := context.Background()
	resources, total, err := uc.resourceDAO.QueryResources(ctx, params, pageSize, offset)
	if err != nil {
		log.Printf("查询资源列表失败: %v", err)
		response.Error(c, http.StatusInternalServerError, fmt.Sprintf("查询资源列表失败: %v", err))
		return
	}

	// 返回响应
	response.Success(c, api.NewQueryResourcesResponse(
		total,
		page,
		pageSize,
		resources,
		params,
	))
}

// isValidFileType 检查文件类型是否有效
func isValidFileType(fileType string) bool {
	validTypes := map[string]bool{
		"pdf":  true,
		"doc":  true,
		"docx": true,
		"xls":  true,
		"xlsx": true,
		"ppt":  true,
		"pptx": true,
		"txt":  true,
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"gif":  true,
		"mp4":  true,
		"mp3":  true,
	}
	return validTypes[strings.ToLower(fileType)]
}

// isValidOrderByField 检查排序字段是否有效
func isValidOrderByField(field string) bool {
	validFields := map[string]bool{
		"user_id":     true,
		"school_id":   true,
		"file_name":   true,
		"file_scope":  true,
		"status":      true,
		"create_time": true,
	}
	return validFields[field]
}

// DeleteResource 删除资源
func (uc *UploadController) DeleteResource(c *gin.Context) {
	// 解析请求数据
	var requestData api.DeleteResourceRequest

	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err))
		return
	}

	// 检查资源ID
	if requestData.ResourceID <= 0 {
		response.Error(c, http.StatusBadRequest, "资源ID不能为空或无效")
		return
	}

	// 检查用户ID
	if requestData.UserID <= 0 {
		response.Error(c, http.StatusBadRequest, "用户ID不能为空或无效")
		return
	}

	// 执行软删除
	ctx := context.Background()
	err := uc.resourceDAO.SoftDeleteResource(ctx, requestData.ResourceID, requestData.UserID)
	if err != nil {
		log.Printf("删除资源失败: %v", err)
		response.Error(c, http.StatusInternalServerError, fmt.Sprintf("删除资源失败: %v", err))
		return
	}

	log.Printf("资源删除成功: ID=%d, 用户ID=%d", requestData.ResourceID, requestData.UserID)

	// 返回成功响应
	response.Success(c, api.NewDeleteResourceResponse(
		requestData.ResourceID,
		requestData.UserID,
	))
}
