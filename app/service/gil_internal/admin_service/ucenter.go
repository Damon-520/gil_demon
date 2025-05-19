package admin_service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gil_teacher/app/conf"
	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
	httputil "gil_teacher/app/core/http"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	"gil_teacher/app/model/itl"
	"gil_teacher/app/utils"
)

// UcenterClient Ucenter API客户端
type UcenterClient struct {
	host   string            // API服务地址
	client *http.Client      // HTTP客户端
	cache  *dao.ApiRdbClient // 缓存客户端
	log    *logger.ContextLogger
}

// NewUcenterClient 创建Ucenter API客户端
func NewUcenterClient(config *conf.Conf, cache *dao.ApiRdbClient, l *logger.ContextLogger) (*UcenterClient, error) {
	return &UcenterClient{
		host: config.Config.GilAdminAPI.UcenterHost,
		client: &http.Client{
			Timeout: consts.TeacherDefaultAPITimeout,
		},
		cache: cache,
		log:   l,
	}, nil
}

// GetTeacherDetail 获取教师详细信息
func (c *UcenterClient) GetTeacherDetail(ctx context.Context, token string, schoolID string) (*itl.TeacherDetailData, *response.Response) {
	url := fmt.Sprintf("%s%s", c.host, consts.GetTeacherDetailAPI.Path)
	c.log.Debug(ctx, "请求运营平台 API: %s, token: %s", url, token)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, consts.GetTeacherDetailAPI.Method, url, nil)
	if err != nil {
		c.log.Error(ctx, "创建请求失败: %v", err)
		return nil, &response.ERR_GIL_ADMIN
	}

	// 设置请求头
	httputil.SetAuthorization(req, token)
	req.Header.Set(consts.UcenterCustomHeaderUserTypeID, consts.UcenterCustomHeaderUserTypeIDValue)
	req.Header.Set(consts.UcenterCustomHeaderOrganizationID, schoolID)

	// 实现重试逻辑
	var resp *http.Response
	var lastErr error

	for i := 0; i <= consts.DefaultMaxRetries; i++ {
		if i > 0 {
			time.Sleep(consts.TeacherDefaultRetryInterval)
			c.log.Warn(ctx, "正在进行第%d次重试请求, URL: %s", i, url)
		}

		resp, err = c.client.Do(req)
		if err == nil {
			break
		}
		lastErr = err
	}

	if err != nil {
		c.log.Error(ctx, "发送请求失败（重试%d次后）: %v", consts.DefaultMaxRetries, lastErr)
		return nil, &response.ERR_GIL_ADMIN
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		// 区分 401 未授权错误
		if resp.StatusCode == http.StatusUnauthorized {
			c.log.Warn(ctx, "运营平台 API 返回未授权错误: %d %s", resp.StatusCode, resp.Status)
			return nil, &response.ERR_UNAUTHORIZED
		}
		c.log.Error(ctx, "运营平台 API 返回错误状态码: %d %s", resp.StatusCode, resp.Status)
		return nil, &response.ERR_GIL_ADMIN
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error(ctx, "运营平台 API 读取响应失败: %v", err)
		return nil, &response.ERR_GIL_ADMIN
	}

	// 记录原始响应内容用于调试
	bodyStr := string(body)
	c.log.Debug(ctx, "收到运营平台 API 响应, URL: %s, 状态码: %d, 请求ID: %s, 响应内容: %s",
		url, resp.StatusCode, req.Header.Get("X-Request-ID"), bodyStr)

	// 解析响应
	var apiResp itl.TeacherDetailResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		c.log.Error(ctx, "运营平台 API 解析响应失败: %v, 响应内容: %s", err, bodyStr)
		return nil, &response.ERR_GIL_ADMIN
	}

	// 检查API响应状态
	if apiResp.Code != 0 {
		c.log.Error(ctx, "运营平台 API 返回错误: %s", apiResp.Message)
		return nil, &response.ERR_GIL_ADMIN
	}

	return &apiResp.Data, nil
}

// GetSchoolMaterial 获取学校的学科教材
func (c *UcenterClient) GetSchoolMaterial(ctx context.Context, schoolID int64) ([]itl.GradeMaterial, *response.Response) {
	url := fmt.Sprintf("%s%s?schoolId=%d", c.host, consts.GetSchoolMaterialAPI.Path, schoolID)
	c.log.Debug(ctx, "请求运营平台 API: %s, schoolID: %d", url, schoolID)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, consts.GetSchoolMaterialAPI.Method, url, nil)
	if err != nil {
		c.log.Error(ctx, "创建请求失败: %v", err)
		return nil, &response.ERR_GIL_ADMIN
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error(ctx, "发送请求失败: %v", err)
		return nil, &response.ERR_GIL_ADMIN
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		// 区分 401 未授权错误
		if resp.StatusCode == http.StatusUnauthorized {
			c.log.Warn(ctx, "运营平台 API 返回未授权错误: %d %s", resp.StatusCode, resp.Status)
			return nil, &response.ERR_UNAUTHORIZED
		}
		c.log.Error(ctx, "运营平台 API 返回错误状态码: %d %s", resp.StatusCode, resp.Status)
		return nil, &response.ERR_GIL_ADMIN
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error(ctx, "运营平台 API 读取响应失败: %v", err)
		return nil, &response.ERR_GIL_ADMIN
	}

	// 记录原始响应内容用于调试
	bodyStr := string(body)
	c.log.Debug(ctx, "收到运营平台 API 响应, URL: %s, 状态码: %d, 请求ID: %s, 响应内容: %s",
		url, resp.StatusCode, req.Header.Get("X-Request-ID"), bodyStr)

	// 解析响应
	var apiResp itl.GetSchoolMaterialResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		c.log.Error(ctx, "运营平台 API 解析响应失败: %v, 响应内容: %s", err, bodyStr)
		return nil, &response.ERR_GIL_ADMIN
	}

	// 检查API响应状态
	if apiResp.Code != 0 {
		c.log.Error(ctx, "运营平台 API 返回错误: %s", apiResp.Message)
		return nil, &response.ERR_GIL_ADMIN
	}

	return apiResp.Data, nil
}

// GetClassStudent 获取班级学生
//
//	map[classID]*itl.ClassInfo
func (c *UcenterClient) GetClassStudent(ctx context.Context, schoolID int64, classIDs []int64) (map[int64]*itl.ClassInfo, error) {
	schoolClassMap := make(map[int64]*itl.ClassInfo)
	// 从缓存中获取数据，按学校班级存储该班全部学生数据，如果数据不存在，则从运营平台获取
	for _, classID := range classIDs {
		// TODO 开发阶段，不读缓存
		if true {
			continue
		}

		class, students, err := c.getClassCache(ctx, schoolID, classID)
		if err != nil {
			c.log.Error(ctx, "获取班级缓存数据失败: %v", err)
			continue
		}

		schoolClassMap[classID] = &itl.ClassInfo{
			ClassID:   class.ID,
			ClassName: class.Name,
			Students:  students,
		}
	}

	// 如果所有班级数据都从缓存中获取到了，直接返回
	if len(schoolClassMap) == len(classIDs) {
		return schoolClassMap, nil
	}

	// 找出缓存中不存在的班级ID
	missingClassIDs := make([]int64, 0)
	for _, classID := range classIDs {
		if _, exists := schoolClassMap[classID]; !exists {
			missingClassIDs = append(missingClassIDs, classID)
		}
	}

	// 将 []int64 转换为 []string
	strIDs := make([]string, len(missingClassIDs))
	for i, id := range missingClassIDs {
		strIDs[i] = fmt.Sprintf("%d", id)
	}
	url := fmt.Sprintf("%s%s?classIDs=%s", c.host, consts.GetClassStudentAPI.Path, strings.Join(strIDs, ","))
	c.log.Debug(ctx, "请求运营平台 API: %s", url)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, consts.GetClassStudentAPI.Method, url, nil)
	if err != nil {
		c.log.Error(ctx, "创建请求失败: %v", err)
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error(ctx, "发送请求失败: %v", err)
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Error(ctx, "运营平台 API 返回错误状态码: %d %s", resp.StatusCode, resp.Status)
		return nil, fmt.Errorf("运营平台 API 返回错误状态码: %d %s", resp.StatusCode, resp.Status)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error(ctx, "运营平台 API 读取响应失败: %v", err)
		return nil, fmt.Errorf("运营平台 API 读取响应失败: %w", err)
	}

	// 记录原始响应内容用于调试
	bodyStr := string(body)
	c.log.Debug(ctx, "收到运营平台 API 响应, URL: %s, 状态码: %d, 请求ID: %s, 响应内容: %s",
		url, resp.StatusCode, req.Header.Get("X-Request-ID"), bodyStr)

	// 解析响应
	var apiResp itl.GetClassStudentResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		c.log.Error(ctx, "运营平台 API 解析响应失败: %v, 响应内容: %s", err, bodyStr)
		return nil, fmt.Errorf("运营平台 API 解析响应失败: %w, 响应内容: %s", err, bodyStr)
	}

	// 检查API响应状态
	if apiResp.Code != 0 {
		c.log.Error(ctx, "运营平台 API 返回错误: %s", apiResp.Message)
		return nil, fmt.Errorf("运营平台 API 返回错误: %s", apiResp.Message)
	}

	schoolClassMap, err = c.setClassCache(ctx, schoolID, apiResp.Data)
	if err != nil {
		c.log.Error(ctx, "写入缓存失败: %v", err)
		return nil, fmt.Errorf("写入缓存失败: %w", err)
	}

	return schoolClassMap, nil
}

// 从缓存中读取班级信息和班级学生信息
func (c *UcenterClient) getClassCache(ctx context.Context, schoolID int64, classID int64) (*itl.Class, []*itl.StudentInfo, error) {
	// 班级信息
	class := &itl.Class{}
	classKey := consts.ClassInfoKey(schoolID, classID)
	exists, err := c.cache.Get(ctx, classKey, class)
	if err != nil {
		c.log.Error(ctx, "获取班级缓存数据失败: %v", err)
		return nil, nil, err
	}
	if !exists {
		c.log.Debug(ctx, "班级缓存数据不存在: %s", classKey)
		return nil, nil, nil
	}

	// 学生信息
	students := make([]*itl.StudentInfo, 0)
	classStudentsKey := consts.ClassStudentKey(schoolID, classID)
	exists, err = c.cache.HGetAll(ctx, classStudentsKey, &students)
	if err != nil {
		c.log.Error(ctx, "获取学生缓存数据失败: %v", err)
		return nil, nil, err
	}
	if !exists {
		c.log.Debug(ctx, "学生缓存数据不存在: %s", classStudentsKey)
		return nil, nil, nil
	}

	return class, students, nil
}

// 班级信息和学生信息写缓存
func (c *UcenterClient) setClassCache(ctx context.Context, schoolID int64, classInfos []itl.ClassInfo) (map[int64]*itl.ClassInfo, error) {
	schoolClassMap := make(map[int64]*itl.ClassInfo)
	if len(classInfos) == 0 {
		return schoolClassMap, nil
	}

	// 写入缓存
	for _, classInfo := range classInfos {
		schoolClassMap[classInfo.ClassID] = &classInfo

		// 写入班级信息缓存
		class := &itl.Class{
			ID:   classInfo.ClassID,
			Name: classInfo.ClassName,
		}

		classKey := consts.ClassInfoKey(schoolID, classInfo.ClassID)
		if err := c.cache.Set(ctx, classKey, class, consts.ClassStudentExpire); err != nil {
			c.log.Error(ctx, "写入班级缓存失败: %v", err)
			// 继续处理其他班级，不中断流程
			continue
		}

		// 写入学生信息缓存
		classStudentsKey := consts.ClassStudentKey(schoolID, classInfo.ClassID)
		if err := c.cache.HSet(ctx, classStudentsKey, classInfo.Students, consts.ClassStudentExpire); err != nil {
			c.log.Error(ctx, "写入学生缓存失败: %v", err)
			// 继续处理其他班级，不中断流程
			continue
		}
	}

	return schoolClassMap, nil
}

// GetGradeClassInfo 获取年级班级信息
func (c *UcenterClient) GetGradeClassInfo(ctx context.Context, schoolID int64, gradeID ...int64) ([]itl.GradeClass, error) {
	// 构建基础URL
	url := fmt.Sprintf("%s%s?schoolID=%d", c.host, consts.GetGradeClassInfoAPI.Path, schoolID)

	// 如果有年级ID参数，则添加到URL中
	if len(gradeID) > 0 {
		url = fmt.Sprintf("%s&gradeIDs=%s", url, utils.Int64SliceToString(gradeID))
	}

	c.log.Debug(ctx, "请求运营平台 API: %s", url)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, consts.GetGradeClassInfoAPI.Method, url, nil)
	if err != nil {
		c.log.Error(ctx, "创建请求失败: %v", err)
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error(ctx, "发送请求失败: %v", err)
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Error(ctx, "运营平台 API 返回错误状态码: %d %s", resp.StatusCode, resp.Status)
		return nil, fmt.Errorf("运营平台 API 返回错误状态码: %d %s", resp.StatusCode, resp.Status)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error(ctx, "运营平台 API 读取响应失败: %v", err)
		return nil, fmt.Errorf("运营平台 API 读取响应失败: %w", err)
	}

	// 记录原始响应内容用于调试
	bodyStr := string(body)
	c.log.Debug(ctx, "收到运营平台 API 响应, URL: %s, 状态码: %d, 请求ID: %s, 响应内容: %s",
		url, resp.StatusCode, req.Header.Get("X-Request-ID"), bodyStr)

	// 解析响应
	var apiResp itl.GetGradeClassInfoResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		c.log.Error(ctx, "运营平台 API 解析响应失败: %v, 响应内容: %s", err, bodyStr)
		return nil, fmt.Errorf("运营平台 API 解析响应失败: %w, 响应内容: %s", err, bodyStr)
	}

	// 检查API响应状态
	if apiResp.Code != 0 {
		c.log.Error(ctx, "运营平台 API 返回错误: %s", apiResp.Message)
		return nil, fmt.Errorf("运营平台 API 返回错误: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

// GetStudentInfoByID 通过学生ID查询学生信息，返回 map[学生ID]itl.StudentInfoData
func (c *UcenterClient) GetStudentInfoByID(ctx context.Context, schoolID int64, studentIDs []int64) (map[int64]itl.StudentInfoData, error) {
	url := fmt.Sprintf("%s%s?organizationId=%d&studentIDs=%s", c.host, consts.GetStudentInfoByIDAPI.Path, schoolID, utils.Int64SliceToString(studentIDs))
	c.log.Debug(ctx, "请求运营平台 API: %s", url)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, consts.GetStudentInfoByIDAPI.Method, url, nil)
	if err != nil {
		c.log.Error(ctx, "创建请求失败: %v", err)
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error(ctx, "发送请求失败: %v", err)
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Error(ctx, "运营平台 API 返回错误状态码: %d %s", resp.StatusCode, resp.Status)
		return nil, fmt.Errorf("运营平台 API 返回错误状态码: %d %s", resp.StatusCode, resp.Status)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error(ctx, "运营平台 API 读取响应失败: %v", err)
		return nil, fmt.Errorf("运营平台 API 读取响应失败: %w", err)
	}

	// 记录原始响应内容用于调试
	bodyStr := string(body)
	c.log.Debug(ctx, "收到运营平台 API 响应, URL: %s, 状态码: %d, 请求ID: %s, 响应内容: %s",
		url, resp.StatusCode, req.Header.Get("X-Request-ID"), bodyStr)

	// 解析响应
	var apiResp itl.GetStudentInfoByIDResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		c.log.Error(ctx, "运营平台 API 解析响应失败: %v, 响应内容: %s", err, bodyStr)
		return nil, fmt.Errorf("运营平台 API 解析响应失败: %w, 响应内容: %s", err, bodyStr)
	}

	// 检查API响应状态
	if apiResp.Code != 0 {
		c.log.Error(ctx, "运营平台 API 返回错误: %s", apiResp.Message)
		return nil, fmt.Errorf("运营平台 API 返回错误: %s", apiResp.Message)
	}

	studentInfoMap := make(map[int64]itl.StudentInfoData)
	for studentIDStr, studentInfo := range apiResp.Data {
		studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
		if err != nil {
			c.log.Error(ctx, "解析学生ID失败: %v", err)
			return nil, fmt.Errorf("解析学生ID失败: %w", err)
		}
		studentInfoMap[studentID] = studentInfo
	}

	return studentInfoMap, nil
}
