package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gil_teacher/app/conf"
	"gil_teacher/app/consts"
	httputil "gil_teacher/app/core/http"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	"gil_teacher/app/model/api"
	"gil_teacher/app/utils"
)

// ScheduleCacheService 课程表缓存服务
type ScheduleCacheService struct {
	redisClient *dao.ApiRdbClient
	logger      *logger.ContextLogger
	config      *conf.Config
}

// NewScheduleCacheService 创建新的课程表缓存服务
func NewScheduleCacheService(redisClient *dao.ApiRdbClient, logger *logger.ContextLogger, cfg *conf.Config) *ScheduleCacheService {
	return &ScheduleCacheService{
		redisClient: redisClient,
		logger:      logger,
		config:      cfg,
	}
}

// getActualDate 获取实际日期
// 如果date是1或2位数字（表示星期几），尝试从dates映射中获取实际日期
// 如果映射不存在，则返回fallbackDate（如果提供）或原始date
func (s *ScheduleCacheService) getActualDate(ctx context.Context, date string, dates map[string]string, fallbackDate string) string {
	if len(date) == 1 || len(date) == 2 { // 如果是数字字符串（如"1", "2"等表示星期几）
		if mappedDate, ok := dates[date]; ok {
			return mappedDate // 使用映射的实际日期
		} else {
			// 如果找不到映射且提供了fallback日期，使用fallback
			if fallbackDate != "" {
				s.logger.Debug(ctx, "无法找到日期映射，键: %s，使用备用日期: %s", date, fallbackDate)
				return fallbackDate
			}
			// 否则记录日志并返回原始日期
			s.logger.Debug(ctx, "无法找到日期映射，键: %s", date)
		}
	}
	return date
}

// buildScheduleAPIURL 构建课程表API的URL
func (s *ScheduleCacheService) buildScheduleAPIURL(ctx context.Context, req *api.FetchScheduleRequest) (string, error) {
	// 构建基础URL
	var path string
	// if isInternal {
	// 	// 使用内部接口路径（不需要token）
	path = consts.GetInternalClassRoomAPI.Path
	// } else {
	// 	// 使用外部接口路径（需要token）
	// 	path = consts.GetClassRoomAPI.Path
	// }

	baseURL := s.config.GilAdminAPI.UcenterHost + path
	u, err := url.Parse(baseURL)
	if err != nil {
		s.logger.Error(ctx, "解析基础URL失败: %v", err)
		return "", err
	}

	// 构建查询参数
	q := u.Query()
	q.Set("teacher_id", strconv.FormatInt(req.TeacherID, 10))
	q.Set("school_id", strconv.FormatInt(req.SchoolID, 10))
	q.Set("school_year_id", strconv.FormatInt(req.SchoolYearID, 10))
	q.Set("start_date", req.StartDate)
	q.Set("end_date", req.EndDate)
	u.RawQuery = q.Encode()

	// 返回完整的URL字符串
	return u.String(), nil
}

// FetchScheduleFromAPI 从GilAdminAPI获取课程表数据
func (s *ScheduleCacheService) FetchScheduleFromAPI(ctx context.Context, req *api.FetchScheduleRequest) (*api.ScheduleResponse, error) {
	// // 判断是否使用内部接口（无需token）
	// isInternal := token == ""

	// 构建API URL
	apiURL, err := s.buildScheduleAPIURL(ctx, req)
	if err != nil {
		return nil, err
	}

	// if isInternal {
	// 	s.logger.Info(ctx, "使用内部API URL（无需token）: %s", apiURL)
	// } else {
	// 	s.logger.Info(ctx, "使用外部API URL（需要token）: %s", apiURL)
	// }

	// 创建GET请求
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		s.logger.Error(ctx, "创建HTTP请求失败: %v", err)
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httputil.SetJSONHeaders(httpReq)

	// 只有外部接口需要设置token
	// if !isInternal && token != "" {
	// 	s.logger.Debug(ctx, "设置Bearer token认证")
	// 	httputil.SetBearerToken(httpReq, token)
	// }

	// 创建HTTP客户端并设置超时时间
	client := &http.Client{
		Timeout: consts.TeacherDefaultAPITimeout,
	}

	// 发送请求
	s.logger.Info(ctx, "正在发送课程表请求, URL: %s", apiURL)
	resp, err := client.Do(httpReq)
	if err != nil {
		s.logger.Error(ctx, "发送HTTP请求失败: %v", err)
		return nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		s.logger.Error(ctx, "API返回错误状态码: %d", resp.StatusCode)
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error(ctx, "读取响应体失败: %v", err)
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	s.logger.Debug(ctx, "收到API响应: %s", string(body))

	// 解析完整的API响应
	var response api.ApiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		s.logger.Error(ctx, "解析API响应失败: %v, 响应内容: %s", err, string(body))
		return nil, fmt.Errorf("解析API响应失败: %w", err)
	}

	// 当API返回record not found错误时，返回空的课程表
	if strings.Contains(response.Message, "record not found") {
		s.logger.Info(ctx, "没有找到课程表数据，返回空数据")
		return &api.ScheduleResponse{
			Schedule: make(map[string][]api.Schedule),
		}, nil
	}

	// 检查API响应状态
	if response.Code != 0 && response.Code != int64(http.StatusOK) {
		s.logger.Error(ctx, "API返回错误: %s", response.Message)
		return nil, fmt.Errorf("API返回错误: %s", response.Message)
	}

	// 如果API返回的数据为空，初始化一个空的map
	if response.Data.Schedule == nil {
		response.Data.Schedule = make(map[string][]api.Schedule)
	}

	// 计算每个课程的IsInTimeRange字段
	for date, schedules := range response.Data.Schedule {
		// 检查是否日期格式，如果不是，尝试从Dates映射中获取实际日期
		actualDate := s.getActualDate(ctx, date, response.Data.Dates, "")
		if actualDate != date && len(response.Data.Dates) > 0 && !strings.Contains(actualDate, "-") {
			// 如果找不到映射，跳过这个日期
			continue
		}

		for i, schedule := range schedules {
			isInRange, err := utils.IsTimeInRange(actualDate, schedule.ScheduleTplPeriodStartTime, schedule.ScheduleTplPeriodEndTime)
			if err != nil {
				s.logger.Error(ctx, "计算时间范围失败: %v", err)
				continue
			}
			schedules[i].IsInTimeRange = isInRange
		}
		response.Data.Schedule[date] = schedules
	}

	s.logger.Info(ctx, "成功从API获取课程表数据")
	return &response.Data, nil
}

// SaveScheduleToRedis 保存课程表数据到Redis
func (s *ScheduleCacheService) SaveScheduleToRedis(ctx context.Context, req *api.FetchScheduleRequest, scheduleResp *api.ScheduleResponse) error {
	// 获取当前北京时间
	now := time.Now().In(consts.LocationShanghai)
	currentDate := now.Format(consts.TimeFormatDate)
	currentTime := now.Format(consts.TimeFormatHHMMSS)

	// 计算每个课程的IsInTimeRange字段和IsInClass字段
	for date, schedules := range scheduleResp.Schedule {
		// 检查是否日期格式，如果不是，尝试从Dates映射中获取实际日期
		actualDate := s.getActualDate(ctx, date, scheduleResp.Dates, currentDate)

		for i, schedule := range schedules {
			// 检查当前时间是否在课程时间范围内（可进入课程）
			isInRange, err := utils.IsTimeInRange(actualDate, schedule.ScheduleTplPeriodStartTime, schedule.ScheduleTplPeriodEndTime)
			if err != nil {
				s.logger.Error(ctx, "计算时间范围失败: %v", err)
				continue
			}
			schedules[i].IsInTimeRange = isInRange

			// 检查当前时间是否在课程时间段内（需要同时判断日期和时间）
			if actualDate == currentDate && currentTime >= schedule.ScheduleTplPeriodStartTime && currentTime <= schedule.ScheduleTplPeriodEndTime {
				schedules[i].IsInClass = true
			} else {
				schedules[i].IsInClass = false
			}

			// 初始化ClassroomID为schedule_id
			schedules[i].ClassroomID = schedule.ScheduleID
			// 如果是临时课，则修改ClassroomID
			if schedule.IsTmp == 1 {
				schedules[i].ClassroomID = consts.GenerateTempClassroomID(schedule.TmpScheduleID)
			}
		}
		scheduleResp.Schedule[date] = schedules
	}

	// 生成Redis键，使用学校ID和教师ID
	key := consts.GetTeacherScheduleKey(req.SchoolID, req.TeacherID)

	// 检查Redis中是否已存在数据
	var existingDataStr string
	exists, err := s.redisClient.Get(ctx, key, &existingDataStr)
	if err != nil && exists {
		s.logger.Error(ctx, "检查Redis中是否存在数据失败: %v", err)
		// 出错但仍继续执行，视为不存在数据
	}

	// 如果存在数据，则需要合并而不是覆盖
	if exists && existingDataStr != "" {
		// 解析现有数据
		var existingScheduleResp api.ScheduleResponse
		if err := json.Unmarshal([]byte(existingDataStr), &existingScheduleResp); err != nil {
			s.logger.Error(ctx, "解析现有课程表数据失败: %v", err)
			// 解析失败，继续使用新数据
		} else {
			s.logger.Info(ctx, "Redis中存在现有数据，将进行合并而非覆盖")

			// 合并日期映射
			if existingScheduleResp.Dates != nil {
				if scheduleResp.Dates == nil {
					scheduleResp.Dates = make(map[string]string)
				}
				for k, v := range existingScheduleResp.Dates {
					if _, exists := scheduleResp.Dates[k]; !exists {
						scheduleResp.Dates[k] = v
					}
				}
			}

			// 合并课程表数据
			if existingScheduleResp.Schedule != nil {
				for date, existingSchedules := range existingScheduleResp.Schedule {
					if _, ok := scheduleResp.Schedule[date]; !ok {
						// 新数据中不存在该日期的课程，直接使用现有数据
						scheduleResp.Schedule[date] = existingSchedules
					} else {
						// 新数据中存在该日期的课程，需要检查是否有重复并合并
						existingScheduleMap := make(map[int64]bool)

						// 标记现有课程的ID
						for _, schedule := range scheduleResp.Schedule[date] {
							existingScheduleMap[schedule.ScheduleID] = true
						}

						// 添加不重复的课程
						for _, schedule := range existingSchedules {
							if !existingScheduleMap[schedule.ScheduleID] {
								scheduleResp.Schedule[date] = append(scheduleResp.Schedule[date], schedule)
							}
						}
					}
				}
			}

			s.logger.Info(ctx, "成功合并课程表数据，合并后包含 %d 个日期的课程", len(scheduleResp.Schedule))
		}
	}

	// 过滤掉空数组日期
	for date, schedules := range scheduleResp.Schedule {
		if len(schedules) == 0 {
			s.logger.Debug(ctx, "过滤掉空数组日期: %s", date)
			delete(scheduleResp.Schedule, date)
		}
	}
	s.logger.Debug(ctx, "过滤空数组后包含 %d 个日期的课程", len(scheduleResp.Schedule))

	// 将合并后的课程表数据转换为JSON
	dataBytes, err := json.Marshal(scheduleResp)
	if err != nil {
		s.logger.Error(ctx, "序列化课程表数据失败: %v", err)
		return err
	}

	// 保存到Redis，设置过期时间
	err = s.redisClient.Set(ctx, key, string(dataBytes), consts.TeacherScheduleExpiration)
	if err != nil {
		s.logger.Error(ctx, "保存课程表数据到Redis失败: %v", err)
		return err
	}

	s.logger.Info(ctx, "成功保存课程表数据到Redis, key: %s", key)
	return nil
}

// CheckTeachingStatus 检查老师当前是否在上课
func (s *ScheduleCacheService) CheckTeachingStatus(ctx context.Context, req *api.CheckTeachingStatusRequest) (*api.TeacherTeachingStatus, error) {
	// 生成Redis键，使用学校ID和教师ID
	key := consts.GetTeacherScheduleKey(req.SchoolID, req.TeacherID)

	// 从Redis获取课程表数据
	var dataStr string
	exists, err := s.redisClient.Get(ctx, key, &dataStr)
	if err != nil {
		if !exists {
			s.logger.Info(ctx, "Redis中不存在课程表数据, key: %s", key)
			// 尝试获取该教师的其他可用课程表
			var keys []string
			cacheKey := consts.GetTeacherScheduleSearchPattern(req.TeacherID)
			exists, err = s.redisClient.Keys(ctx, cacheKey, &keys)
			if err != nil || !exists || len(keys) == 0 {
				return &api.TeacherTeachingStatus{
					IsTeaching: false,
				}, nil
			}

			// 使用找到的第一个课程表
			exists, err = s.redisClient.Get(ctx, keys[0], &dataStr)
			if err != nil || !exists {
				return &api.TeacherTeachingStatus{
					IsTeaching: false,
				}, nil
			}

			s.logger.Info(ctx, "找到替代课程表数据: %s", keys[0])
		} else {
			s.logger.Error(ctx, "从Redis获取课程表数据失败: %v", err)
			return &api.TeacherTeachingStatus{
				IsTeaching: false,
			}, nil
		}
	}

	// 反序列化数据
	var scheduleResp api.ScheduleResponse
	if err := json.Unmarshal([]byte(dataStr), &scheduleResp); err != nil {
		s.logger.Error(ctx, "反序列化课程表数据失败: %v", err)
		return &api.TeacherTeachingStatus{
			IsTeaching: false,
		}, nil
	}

	// 使用前端传入的时间，而不是从当前时间获取
	currentTime := req.CheckTime

	// 查找当前时间是否在某个课程的上课时间段内
	var currentSchedule *api.Schedule

	// 遍历所有日期的课程
	for date, schedules := range scheduleResp.Schedule {
		// 检查是否日期格式，如果不是，尝试从Dates映射中获取实际日期
		actualDate := s.getActualDate(ctx, date, scheduleResp.Dates, "")
		if actualDate != date && len(scheduleResp.Dates) > 0 && !strings.Contains(actualDate, "-") {
			// 如果找不到映射，跳过这个日期
			continue
		}

		for i, schedule := range schedules {
			// 检查当前时间是否在课程时间范围内（可进入课程）
			isInRange, err := utils.IsTimeInRange(actualDate, schedule.ScheduleTplPeriodStartTime, schedule.ScheduleTplPeriodEndTime)
			if err != nil {
				s.logger.Error(ctx, "检查时间范围失败: %v", err)
				continue
			}

			// 设置课程的IsInTimeRange字段
			schedules[i].IsInTimeRange = isInRange

			// 检查当前时间是否在课程时间段内
			if currentTime >= schedule.ScheduleTplPeriodStartTime && currentTime <= schedule.ScheduleTplPeriodEndTime {
				schedules[i].IsInClass = true
				scheduleCopy := schedules[i] // 创建副本避免指针问题
				currentSchedule = &scheduleCopy
				s.logger.Debug(ctx, "找到当前正在上课: %s, %s-%s, 课程: %s",
					actualDate,
					schedule.ScheduleTplPeriodStartTime,
					schedule.ScheduleTplPeriodEndTime,
					schedule.ClassScheduleCourse)
			} else {
				schedules[i].IsInClass = false
			}
		}
	}

	// 构造返回结果
	result := &api.TeacherTeachingStatus{
		IsTeaching: currentSchedule != nil,
		Schedule:   currentSchedule,
	}

	s.logger.Info(ctx, "当前教师上课状态: 是否上课=%v", result.IsTeaching)

	return result, nil
}

// IsTeaching 检查老师是否在上课（简化版本）
func (s *ScheduleCacheService) IsTeaching(ctx context.Context, teacherID, schoolID int64, checkDateTime time.Time) (bool, error) {
	// 格式化日期为 YYYY-MM-DD
	checkDate := checkDateTime.Format(consts.TimeFormatDate)
	checkTime := checkDateTime.Format(consts.TimeFormatTimeOnly)

	// 从Redis获取课程表数据
	key := consts.GetTeacherScheduleKey(schoolID, teacherID)

	// 从Redis获取课程表数据
	var dataStr string
	exists, err := s.redisClient.Get(ctx, key, &dataStr)
	if err != nil {
		if !exists {
			s.logger.Info(ctx, "Redis中不存在课程表数据, key: %s", key)
			return false, nil
		}
		s.logger.Error(ctx, "从Redis获取课程表数据失败: %v", err)
		return false, err
	}

	// 反序列化数据
	var scheduleResp api.ScheduleResponse
	if err := json.Unmarshal([]byte(dataStr), &scheduleResp); err != nil {
		s.logger.Error(ctx, "反序列化课程表数据失败: %v", err)
		return false, err
	}

	// 检查当前日期的课程
	// 先尝试直接使用checkDate作为键查找课程
	if schedules, ok := scheduleResp.Schedule[checkDate]; ok {
		// 直接按日期查找到课程
		s.logger.Debug(ctx, "直接找到日期 %s 的课程", checkDate)
		if isTeaching := s.checkSchedulesForTeachingStatus(schedules, checkDate, checkTime); isTeaching {
			return true, nil
		}
	} else {
		// 如果直接用日期找不到，尝试通过Dates映射查找
		found := false
		for weekdayStr, dateStr := range scheduleResp.Dates {
			if dateStr == checkDate {
				// 找到了映射的日期，使用星期几作为键查找课程
				if weekdaySchedules, ok := scheduleResp.Schedule[weekdayStr]; ok {
					s.logger.Debug(ctx, "通过星期 %s 找到日期 %s 的课程", weekdayStr, checkDate)
					if isTeaching := s.checkSchedulesForTeachingStatus(weekdaySchedules, checkDate, checkTime); isTeaching {
						return true, nil
					}
					found = true
				}
			}
		}

		// 如果没有找到映射关系，尝试遍历所有课程
		if !found {
			s.logger.Debug(ctx, "没有找到日期 %s 的直接映射，尝试遍历所有课程", checkDate)
			for dateKey, dateSchedules := range scheduleResp.Schedule {
				// 对于数字键（如"1"、"2"等表示星期几）需要使用Dates映射获取实际日期
				actualDate := s.getActualDate(ctx, dateKey, scheduleResp.Dates, "")
				if actualDate != dateKey && len(scheduleResp.Dates) > 0 && !strings.Contains(actualDate, "-") {
					// 如果找不到映射，跳过
					continue
				}

				// 检查实际日期是否匹配
				if actualDate == checkDate {
					s.logger.Debug(ctx, "找到匹配日期 %s 的课程，键为 %s", checkDate, dateKey)
					if isTeaching := s.checkSchedulesForTeachingStatus(dateSchedules, checkDate, checkTime); isTeaching {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

// GetScheduleFromRedis 从Redis获取课程表数据
func (s *ScheduleCacheService) GetScheduleFromRedis(ctx context.Context, teacherID, schoolID int64, date string) ([]api.Schedule, error) {
	// 生成Redis键
	key := consts.GetTeacherScheduleKey(schoolID, teacherID)

	// 检查键是否存在
	exists, err := s.redisClient.KeyExists(ctx, key)
	if err != nil {
		s.logger.Error(ctx, "检查Redis键是否存在失败: %v", err)
		return nil, err
	}

	if !exists {
		s.logger.Warn(ctx, "Redis中不存在该教师的课程表数据: %s", key)
		return nil, nil
	}

	// 从Redis获取数据
	var value string
	exists, err = s.redisClient.Get(ctx, key, &value)
	if err != nil {
		if !exists {
			s.logger.Info(ctx, "Redis中不存在该键: %s", key)
			return nil, nil
		}
		s.logger.Error(ctx, "从Redis获取数据失败: %v", err)
		return nil, err
	}

	// 解析JSON数据
	var scheduleResp api.ScheduleResponse
	if err := json.Unmarshal([]byte(value), &scheduleResp); err != nil {
		s.logger.Error(ctx, "解析课程表数据失败: %v", err)
		return nil, err
	}

	// 如果指定了日期，则只返回该日期的课程表数据
	if date != "" {
		if schedules, ok := scheduleResp.Schedule[date]; ok {
			return schedules, nil
		}
		return []api.Schedule{}, nil
	}

	// 否则返回整个课程表数据
	return nil, nil
}

// SaveTeacherToList 将教师ID和学校ID保存到Redis的教师列表中
func (s *ScheduleCacheService) SaveTeacherToList(ctx context.Context, teacherID, schoolID int64) error {
	// 创建教师信息
	teacherInfo := consts.GetTeacherInfoKey(schoolID, teacherID)

	// 添加到集合中（使用SADD命令，确保不会有重复）
	err := s.redisClient.SAdd(ctx, consts.GetTeacherListKey(schoolID, teacherID), teacherInfo, consts.TeacherListExpiration)
	if err != nil {
		s.logger.Error(ctx, "将教师添加到Redis列表失败: %v", err)
		return fmt.Errorf("将教师添加到Redis列表失败: %w", err)
	}

	s.logger.Info(ctx, "成功将教师(ID=%d, 学校ID=%d)添加到Redis列表", teacherID, schoolID)
	return nil
}

// GetTeacherList 从Redis获取教师列表 TODO
func (s *ScheduleCacheService) GetTeacherList(ctx context.Context) ([]map[string]int64, error) {
	// 从Redis获取教师列表
	var members []string
	exists, err := s.redisClient.SMembers(ctx, consts.TeacherListKey, &members) // key 有问题 TODO
	if err != nil {
		s.logger.Error(ctx, "从Redis获取教师列表失败: %v", err)
		return nil, err
	}

	if !exists {
		s.logger.Info(ctx, "Redis中没有教师列表数据")
		return []map[string]int64{}, nil
	}

	if len(members) == 0 {
		s.logger.Info(ctx, "Redis中没有教师列表数据")
		return []map[string]int64{}, nil
	}

	// 解析教师信息
	var teacherList []map[string]int64
	for _, member := range members {
		parts := strings.Split(member, ":")
		if len(parts) != 2 {
			s.logger.Warn(ctx, "教师信息格式不正确: %s", member)
			continue
		}

		teacherID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			s.logger.Warn(ctx, "教师ID格式不正确: %s", parts[0])
			continue
		}

		schoolID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			s.logger.Warn(ctx, "学校ID格式不正确: %s", parts[1])
			continue
		}

		teacherInfo := map[string]int64{
			"teacher_id": teacherID,
			"school_id":  schoolID,
		}
		teacherList = append(teacherList, teacherInfo)
	}

	return teacherList, nil
}

// checkTimeInRange 检查当前时间是否在课程时间范围内
func (s *ScheduleCacheService) checkTimeInRange(ctx context.Context, date string, startTime, endTime string) bool {
	isInRange, err := utils.IsTimeInRange(date, startTime, endTime)
	if err != nil {
		s.logger.Error(ctx, "解析时间失败: %v", err)
		return false
	}
	return isInRange
}

// GetScheduleByClassroomID 通过教室ID从Redis获取课程表数据
func (s *ScheduleCacheService) GetScheduleByClassroomID(ctx context.Context, classroomID, teacherID, schoolID int64, date string) (*api.Schedule, error) {
	// 获取当前北京时间
	now := time.Now().In(consts.LocationShanghai)
	currentDate := now.Format(consts.TimeFormatDate)
	currentTime := now.Format(consts.TimeFormatHHMMSS)

	// 如果没有指定日期，使用当前日期
	if date == "" {
		date = currentDate
	}

	// 获取教师课程表数据
	key := consts.GetTeacherScheduleKey(schoolID, teacherID)
	s.logger.Info(ctx, "使用key: %s", key)

	// 从Redis获取数据
	var value string
	exists, err := s.redisClient.Get(ctx, key, &value)
	if err != nil || !exists {
		s.logger.Warn(ctx, "获取Redis key失败: %s, error: %v", key, err)

		// 尝试查找该教师的其他课程表
		var keys []string
		pattern := consts.GetTeacherScheduleSearchPattern(teacherID)
		exists, err = s.redisClient.Keys(ctx, pattern, &keys)
		if err != nil || !exists || len(keys) == 0 {
			s.logger.Warn(ctx, "找不到教师的任何课程表: pattern=%s", pattern)
			return nil, nil
		}

		s.logger.Info(ctx, "找到教师的其他课程表: %v", keys)

		// 尝试每个找到的课程表
		for _, altKey := range keys {
			exists, err = s.redisClient.Get(ctx, altKey, &value)
			if !exists || err != nil {
				continue
			}

			// 尝试解析找到的课程表
			var scheduleResp api.ScheduleResponse
			if err := json.Unmarshal([]byte(value), &scheduleResp); err != nil {
				s.logger.Error(ctx, "解析替代课程表数据失败: %v, key: %s", err, altKey)
				continue
			}

			// 在这个课程表中查找教室ID
			matchingSchedule := s.findClassroomInSchedule(ctx, classroomID, &scheduleResp, currentDate, currentTime)
			if matchingSchedule != nil {
				s.logger.Info(ctx, "在替代课程表 %s 中找到了教室ID %d", altKey, classroomID)
				return matchingSchedule, nil
			}
		}

		// 所有替代方案都失败了
		return nil, nil
	}

	// 解析JSON数据
	var scheduleResp api.ScheduleResponse
	if err := json.Unmarshal([]byte(value), &scheduleResp); err != nil {
		s.logger.Error(ctx, "解析课程表数据失败: %v", err)
		return nil, err
	}

	// 查找匹配的教室
	return s.findClassroomInSchedule(ctx, classroomID, &scheduleResp, currentDate, currentTime), nil
}

// findClassroomInSchedule 在课程表中查找指定的教室ID
func (s *ScheduleCacheService) findClassroomInSchedule(ctx context.Context, classroomID int64, scheduleResp *api.ScheduleResponse, currentDate, currentTime string) *api.Schedule {
	// 先查找所有日期中是否有匹配的教室ID
	var matchingSchedule *api.Schedule
	var matchingDate string

	// 遍历所有日期的课程表
	for dateKey, schedules := range scheduleResp.Schedule {
		for i, schedule := range schedules {
			// 检查是否是目标教室ID
			if schedule.ClassroomID == classroomID ||
				(schedule.IsTmp == 1 && consts.GenerateTempClassroomID(schedule.TmpScheduleID) == classroomID) {
				// 找到匹配的课程
				matchingSchedule = &schedules[i]
				matchingDate = dateKey
				s.logger.Debug(ctx, "找到匹配的课程: 日期=%s, 教室ID=%d, 班级=%s, 年级=%s",
					dateKey, classroomID, schedule.ClassName, schedule.GradeName)
				break
			}
		}
		if matchingSchedule != nil {
			break
		}
	}

	// 如果没有找到匹配的课程，返回nil
	if matchingSchedule == nil {
		s.logger.Debug(ctx, "未找到教室ID=%d的课程信息", classroomID)
		return nil
	}

	// 设置IsInTimeRange（是否在课程时间范围内）
	isInRange, err := utils.IsTimeInRange(matchingDate, matchingSchedule.ScheduleTplPeriodStartTime, matchingSchedule.ScheduleTplPeriodEndTime)
	if err != nil {
		s.logger.Error(ctx, "检查时间范围失败: %v", err)
	}
	matchingSchedule.IsInTimeRange = isInRange

	// 设置IsInClass（是否正在上课中）- 必须是当天且在课程时间段内
	matchingSchedule.IsInClass = utils.IsInClass(matchingDate, currentDate, currentTime,
		matchingSchedule.ScheduleTplPeriodStartTime,
		matchingSchedule.ScheduleTplPeriodEndTime)

	// 确保ClassroomID正确设置
	if matchingSchedule.IsTmp == 1 {
		matchingSchedule.ClassroomID = consts.GenerateTempClassroomID(matchingSchedule.TmpScheduleID)
	} else {
		matchingSchedule.ClassroomID = matchingSchedule.ScheduleID
	}

	s.logger.Debug(ctx, "找到课程信息: classroomID=%d, date=%s, className=%s, gradeName=%s, isInClass=%v",
		classroomID, matchingDate, matchingSchedule.ClassName, matchingSchedule.GradeName, matchingSchedule.IsInClass)

	return matchingSchedule
}

// checkSchedulesForTeachingStatus 检查指定课程列表中是否有正在进行的课程
// 参数：
// - schedules: 课程列表
// - checkDate: 检查日期
// - checkTime: 检查时间
// 返回：是否有正在进行的课程
func (s *ScheduleCacheService) checkSchedulesForTeachingStatus(schedules []api.Schedule, checkDate, checkTime string) bool {
	for i, schedule := range schedules {
		// 检查当前时间是否在课程时间范围内（可进入课程）
		isInRange, err := utils.IsTimeInRange(checkDate, schedule.ScheduleTplPeriodStartTime, schedule.ScheduleTplPeriodEndTime)
		if err != nil {
			s.logger.Error(context.Background(), "检查时间范围失败: %v", err)
			continue
		}

		// 设置课程的IsInTimeRange字段（虽然这里不需要保存回去）
		schedules[i].IsInTimeRange = isInRange

		// 检查当前时间是否在课程时间段内（需要同时判断日期和时间）
		if checkTime >= schedule.ScheduleTplPeriodStartTime && checkTime <= schedule.ScheduleTplPeriodEndTime {
			schedules[i].IsInClass = true
			return true
		} else {
			schedules[i].IsInClass = false
		}
	}
	return false
}
