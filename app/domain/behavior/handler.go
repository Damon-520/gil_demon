package behavior

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"gil_teacher/app/consts"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	behaviorDao "gil_teacher/app/dao/behavior"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils"
)

// BehaviorHandler 行为消息处理器
type BehaviorHandler struct {
	behaviorDAO behaviorDao.BehaviorDAO
	redisClient *dao.ApiRdbClient
	logger      *clogger.ContextLogger
}

func NewBehaviorHandler(
	behaviorDAO behaviorDao.BehaviorDAO,
	redisClient *dao.ApiRdbClient,
	logger *clogger.ContextLogger,
) *BehaviorHandler {
	return &BehaviorHandler{
		behaviorDAO: behaviorDAO,
		redisClient: redisClient,
		logger:      logger,
	}
}

// GetRedisClient 返回Redis客户端
func (h *BehaviorHandler) GetRedisClient() *dao.ApiRdbClient {
	return h.redisClient
}

func (h *BehaviorHandler) validateTeacherBehavior(behavior *dto.TeacherBehaviorDTO) error {
	if behavior.SchoolID == 0 {
		return errors.New("学校ID不能为0")
	}
	if behavior.ClassID == 0 {
		return errors.New("班级ID不能为0")
	}
	if behavior.TeacherID == 0 {
		return errors.New("教师ID不能为0")
	}
	if behavior.BehaviorType == "" {
		return errors.New("行为类型不能为空")
	}
	if behavior.CreateTime.IsZero() {
		return errors.New("创建时间不能为0")
	}

	return nil
}

func (h *BehaviorHandler) validateStudentBehavior(behavior *dto.StudentBehaviorDTO) error {
	if behavior.SchoolID == 0 {
		return errors.New("学校ID不能为0")
	}
	if behavior.ClassID == 0 {
		return errors.New("班级ID不能为0")
	}
	if behavior.StudentID == 0 {
		return errors.New("学生ID不能为0")
	}
	if behavior.BehaviorType == "" {
		return errors.New("行为类型不能为空")
	}
	if behavior.CreateTime.IsZero() {
		return errors.New("创建时间不能为0")
	}
	return nil
}

func (h *BehaviorHandler) validateCommunication(communication *dto.CommunicationMessageDTO) error {
	if communication.SessionID == "" {
		return errors.New("会话ID不能为空")
	}
	// 发起会话的只能是自然人，不能是 ai
	if !slices.Contains([]consts.CommunicationUserType{
		consts.CommunicationUserTypeStudent,
		consts.CommunicationUserTypeTeacher,
	}, consts.CommunicationUserType(communication.UserType)) {
		return errors.New("发起会话的只能是学生或老师")
	}
	if communication.UserID == 0 {
		return errors.New("用户ID不能为0")
	}
	if communication.MessageContent == "" {
		return errors.New("消息内容不能为空")
	}
	if communication.MessageType == "" {
		return errors.New("消息类型不能为空")
	}

	return nil
}

// OpenCommunicationSession 创建会话
func (h *BehaviorHandler) OpenCommunicationSession(ctx context.Context, req *api.OpenSessionRequest) (string, error) {
	session := &dto.CommunicationSessionDTO{
		UserID:      req.UserID,
		UserType:    req.UserType,
		SchoolID:    req.SchoolID,
		CourseID:    req.CourseID,
		ClassroomID: req.ClassroomID,
		SessionType: req.SessionType,
		TargetID:    setTargetID(req.TargetID),
		StartTime:   time.Now(),
	}
	sessionID, err := h.behaviorDAO.OpenCommunicationSession(ctx, session)
	if err != nil {
		return "", errors.Wrap(err, "创建会话失败")
	}
	// 缓存会话 id，方便查找验证
	h.redisClient.Set(ctx, consts.GetCommunicationSessionKey(sessionID), sessionID, consts.CommunicationSessionExpire)
	return sessionID, nil
}

// CloseCommunicationSession 关闭会话
func (h *BehaviorHandler) CloseCommunicationSession(ctx context.Context, req *api.CloseSessionRequest) error {
	// 检查会话是否存在
	exists, err := h.checkCommunicationSession(ctx, req.SessionID)
	if err != nil {
		return errors.Wrap(err, "检查会话失败")
	}
	if !exists {
		return errors.New("会话不存在")
	}

	return h.behaviorDAO.CloseCommunicationSession(ctx, req.SessionID)
}

// GetCommunicationSessionMessages 查询指定会话的全部消息
func (h *BehaviorHandler) GetCommunicationSessionMessages(ctx context.Context, sessionID string, page, size int64) ([]*dto.CommunicationMessageDTO, error) {
	// 检查会话是否存在
	exists, err := h.checkCommunicationSession(ctx, sessionID)
	if err != nil {
		return nil, errors.Wrap(err, "检查会话失败")
	}
	if !exists {
		return nil, errors.New("会话不存在")
	}

	return h.behaviorDAO.GetCommunicationSessionMessages(ctx, sessionID, &consts.DBPageInfo{Page: page, Limit: size})
}

// GetCommunicationSession 查询指定会话
func (h *BehaviorHandler) GetCommunicationSession(ctx context.Context, sessionID string) (*dto.CommunicationSessionDTO, error) {
	// 检查会话是否存在
	exists, err := h.checkCommunicationSession(ctx, sessionID)
	if err != nil {
		return nil, errors.Wrap(err, "检查会话失败")
	}
	if !exists {
		return nil, errors.New("会话不存在")
	}
	return h.behaviorDAO.GetCommunicationSession(ctx, sessionID)
}

// GetClassroomMessages 查询指定课堂的全部消息
func (h *BehaviorHandler) GetClassroomMessages(ctx context.Context, userID string, classroomID string, page, size int64) ([]*dto.CommunicationMessageDTO, error) {
	// TODO 先要查询用户有无查看消息的权限
	return h.behaviorDAO.GetClassroomMessages(ctx, classroomID, &consts.DBPageInfo{Page: page, Limit: size})
}

// // MarkMessageAsRead 标记用户消息已读
// // 用户有会话的权限才能标记已读
// func (h *BehaviorHandler) MarkMessageAsRead(ctx context.Context, userID string, sessionID string) error {
// 	return h.behaviorDAO.MarkMessageAsRead(ctx, userID, sessionID)
// }

// // GetUnreadMessageCount 获取用户指定会话的未读消息数量
// func (h *BehaviorHandler) GetUnreadMessageCount(ctx context.Context, userID string, sessionID string) (int64, error) {
// 	// 检查会话是否存在
// 	// 检查用户是否在会话关联的班级
// 	return h.behaviorDAO.GetUnreadMessageCount(ctx, userID, sessionID)
// }

// // GetUnreadMessageList 获取用户指定会话的未读消息列表
// func (h *BehaviorHandler) GetUnreadMessageList(ctx context.Context, userID string, sessionID string) ([]*dto.CommunicationMessageDTO, error) {
// 	return h.behaviorDAO.GetUnreadMessageList(ctx, userID, sessionID)
// }

// 检查会话是否存在，先检查缓存，缓存中不存在则从数据库中查询，再检查权限，只有属于这个会话所在班级的学生和老师才能查看
func (h *BehaviorHandler) checkCommunicationSession(ctx context.Context, sessionID string) (bool, error) {
	exists, err := h.redisClient.KeyExists(ctx, consts.GetCommunicationSessionKey(sessionID))
	if err != nil {
		return false, errors.Wrap(err, "查询会话缓存失败")
	}

	if !exists {
		session, err := h.behaviorDAO.GetCommunicationSession(ctx, sessionID)
		if err != nil {
			return false, errors.Wrap(err, "查询会话失败")
		}
		if session == nil {
			return false, nil
		}
		h.redisClient.Set(ctx, consts.GetCommunicationSessionKey(sessionID), session.SessionID, consts.CommunicationSessionExpire)
		return true, nil
	}

	return true, nil
}

func setTargetID(targetID string) *string {
	if targetID == "" {
		return nil
	}
	return &targetID
}

// GetClassLatestBehaviors 获取班级学生最新行为
func (h *BehaviorHandler) GetClassLatestBehaviors(ctx context.Context, ClassroomID uint64) ([]*dto.StudentLatestBehaviorDTO, error) {
	behaviors, err := h.behaviorDAO.GetClassLatestBehaviors(ctx, ClassroomID)
	if err != nil {
		h.logger.Error(ctx, "获取班级学生最新行为失败: %v", err)
		return nil, errors.Wrap(err, "获取班级学生最新行为失败")
	}

	// 检查behaviors是否为nil
	if behaviors == nil {
		h.logger.Error(ctx, "获取班级学生最新行为数据为空")
		return []*dto.StudentLatestBehaviorDTO{}, nil
	}

	// 过滤掉nil元素
	nonNilBehaviors := make([]*dto.StudentLatestBehaviorDTO, 0, len(behaviors))
	for _, behavior := range behaviors {
		if behavior != nil {
			// 从Context中提取学生姓名和头像URL，如果为空则尝试填充
			if behavior.StudentName == "" || behavior.AvatarURL == "" {
				var contextMap map[string]interface{}
				if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
					if behavior.StudentName == "" {
						behavior.StudentName = utils.GetMapStringKey(contextMap, "student_name")
					}
					if behavior.AvatarURL == "" {
						behavior.AvatarURL = utils.GetMapStringKey(contextMap, "avatar_url")
					}
				}
			}
			nonNilBehaviors = append(nonNilBehaviors, behavior)
		}
	}

	return nonNilBehaviors, nil
}

// GetStudentClassroomDetail 获取学生课堂详情
func (h *BehaviorHandler) GetStudentClassroomDetail(ctx context.Context, req *api.StudentClassroomDetailRequest) (*api.StudentClassroomDetailResponse, error) {
	// 参数校验
	if req.StudentID == 0 {
		return nil, errors.New("学生ID不能为0")
	}
	if req.ClassroomID == 0 {
		return nil, errors.New("课堂ID不能为0")
	}

	// 调用DAO层获取数据
	detail, err := h.behaviorDAO.GetStudentClassroomDetail(ctx, req.StudentID, req.ClassroomID)
	if err != nil {
		h.logger.Error(ctx, "获取学生课堂详情失败: %v", err)
		return nil, err
	}

	// 数据有效性检查
	if detail == nil {
		return nil, errors.New("获取学生课堂详情失败: 返回数据为空")
	}

	h.logger.Debug(ctx, "从DAO获取的数据: %+v, 学习记录数量: %d", detail, len(detail.LearningRecords))

	// 转换为API响应格式
	response := &api.StudentClassroomDetailResponse{
		StudentID:        detail.StudentID,
		ClassroomID:      detail.ClassroomID,
		SchoolID:         detail.SchoolID,
		ClassID:          detail.ClassID,
		TotalStudyTime:   detail.TotalStudyTime,
		ClassroomScore:   detail.ClassroomScore,
		MaxCorrectStreak: detail.MaxCorrectStreak,
		QuestionCount:    detail.QuestionCount,
		AccuracyRate:     detail.AccuracyRate,
		InteractionCount: detail.InteractionCount,
		ViolationCount:   detail.ViolationCount,
		LearningRecords:  make([]api.LearningRecordResponse, 0, len(detail.LearningRecords)),
		IsEvaluated:      false,
		EvaluateContent:  "",
		IsHandled:        false,
	}

	// 检查是否已评价
	evaluateKey := consts.GetEvaluateRecordKey(int64(req.ClassroomID), int64(req.StudentID), string(consts.BehaviorTypeClassComment))
	var evaluateTime string
	evaluateExists, _ := h.redisClient.Get(ctx, evaluateKey, &evaluateTime)
	if evaluateExists {
		response.IsEvaluated = true

		// 从数据库中查询评价内容
		behaviors, err := h.behaviorDAO.GetStudentBehaviorsByType(ctx, req.StudentID, req.ClassroomID, string(consts.BehaviorTypeClassComment))
		if err == nil && len(behaviors) > 0 {
			// 找到最近的一条评价记录
			var latestBehavior *behaviorDao.StudentBehavior
			var latestTime time.Time

			for _, behavior := range behaviors {
				if behavior == nil {
					continue
				}

				createTime, err := time.Parse(time.RFC3339, behavior.CreateTime.Format(time.RFC3339))
				if err != nil {
					continue
				}

				if latestBehavior == nil || createTime.After(latestTime) {
					latestBehavior = behavior
					latestTime = createTime
				}
			}

			// 从最近的评价记录中提取评价内容
			if latestBehavior != nil {
				var contextMap map[string]interface{}
				if err := json.Unmarshal([]byte(latestBehavior.Context), &contextMap); err == nil {
					// 尝试获取message字段作为评价内容
					if message, ok := contextMap["message"].(string); ok {
						response.EvaluateContent = message
					} else {
						// 如果没有message字段，则将整个context作为评价内容
						response.EvaluateContent = latestBehavior.Context
					}

					// 提取推送时间
					if pushTime, ok := contextMap["pushTime"].(float64); ok {
						response.PushTime = int64(pushTime)
					}
				}
			}
		}
	}

	// 检查是否已处理
	handledKey := consts.GetClassBehaviorHandledKey(int64(req.ClassroomID))
	var handleTime int64
	handleTimeExists, _ := h.redisClient.HGetField(ctx, handledKey, fmt.Sprintf("%d", req.StudentID), &handleTime)
	if handleTimeExists && handleTime > 0 {
		response.IsHandled = true
	}

	// 如果没有违规次数数据，从行为数据中计算
	if response.ViolationCount == 0 {
		// 获取该学生的行为数据
		behaviors, err := h.behaviorDAO.GetStudentsBehaviors(ctx, []uint64{req.StudentID})
		if err == nil && len(behaviors) > 0 {
			var violationCount int64
			for _, behavior := range behaviors {
				if behavior == nil {
					continue
				}

				// 从Context中提取违规行为数据
				var contextMap map[string]interface{}
				if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
					// 统计违规行为次数
					pageSwitchCount := int64(utils.GetMapIntKey(contextMap, "page_switch_count", 0))
					otherContentCount := int64(utils.GetMapIntKey(contextMap, "other_content_count", 0))
					pauseCount := int64(utils.GetMapIntKey(contextMap, "pause_count", 0))

					violationCount += pageSwitchCount + otherContentCount + pauseCount
				}
			}
			response.ViolationCount = violationCount
		}
	}

	// 转换学习记录列表
	for _, record := range detail.LearningRecords {
		apiRecord := api.LearningRecordResponse{
			RecordID:     record.RecordID,
			ChapterID:    record.ChapterID,
			ChapterName:  record.ChapterName,
			LearningType: record.LearningType,
			Duration:     record.Duration,
			AccuracyRate: record.AccuracyRate,
			Progress:     record.Progress,
			CreateTime:   record.CreateTime,
		}
		response.LearningRecords = append(response.LearningRecords, apiRecord)
	}

	h.logger.Debug(ctx, "返回API响应: %+v, 学习记录数量: %d", response, len(response.LearningRecords))

	return response, nil
}

// GetClassBehaviorCategory 获取课堂行为分类
func (h *BehaviorHandler) GetClassBehaviorCategory(ctx context.Context, classroomID uint64) ([]dto.StudentBehaviorCategoryDTO, error) {
	// 获取课堂学生行为数据
	behaviors, err := h.behaviorDAO.GetClassLatestBehaviors(ctx, classroomID)
	if err != nil {
		h.logger.Error(ctx, "获取课堂学生行为数据失败: %v", err)
		return nil, errors.Wrap(err, "获取课堂学生行为数据失败")
	}

	// 检查behaviors是否为nil
	if behaviors == nil {
		h.logger.Error(ctx, "获取课堂学生行为数据为空")
		return []dto.StudentBehaviorCategoryDTO{}, nil
	}

	// 新增：获取班级所有行为用于聚合不同类型
	allBehaviors, err := h.behaviorDAO.GetClassAllBehaviors(ctx, classroomID)
	if err != nil {
		h.logger.Error(ctx, "获取课堂学生所有行为失败: %v", err)
		// 不要返回错误，继续使用原有的behaviors
	}

	// 获取已处理学生列表
	handledKey := consts.GetClassBehaviorHandledKey(int64(classroomID))
	handledMap := make(map[int64]int64)
	_, err = h.redisClient.HGetAll(ctx, handledKey, &handledMap)
	if err != nil {
		h.logger.Error(ctx, "获取已处理学生列表失败: %v", err)
		// 继续执行，不影响主流程
	}

	// 初始化结果
	// 按学生ID聚合行为
	studentBehaviorsMap := make(map[int64]*dto.StudentBehaviorCategoryDTO)

	// 首先处理最新行为，获取基本信息
	for _, behavior := range behaviors {
		// 确保behavior不为nil
		if behavior == nil {
			h.logger.Warn(ctx, "发现空的学生行为记录，已跳过")
			continue
		}

		studentID := behavior.StudentID

		// 转换为行为分类DTO
		student := dto.StudentBehaviorCategoryDTO{
			StudentID:      uint64(behavior.StudentID),
			BehaviorType:   behavior.BehaviorType,
			TotalQuestions: behavior.TotalQuestions,
			CorrectAnswers: behavior.CorrectAnswers,
			WrongAnswers:   behavior.WrongAnswers,
			AccuracyRate:   behavior.AccuracyRate,
			LastUpdateTime: behavior.LastUpdateTime,
			IsHandled:      false,
		}

		// 从Context中提取学生姓名和头像URL
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
			student.StudentName = utils.GetMapStringKey(contextMap, "student_name")
			student.AvatarUrl = utils.GetMapStringKey(contextMap, "avatar_url")
			student.LearningProgress = utils.GetMapFloat64Key(contextMap, "learning_progress")
		}

		// 检查是否已处理
		if handleTime, exists := handledMap[studentID]; exists && handleTime > 0 {
			student.IsHandled = true
			student.HandleTime = handleTime
		}

		// 获取提醒次数
		reminderKey := consts.GetStudentReminderCountKey(int64(classroomID), studentID)
		var reminderCount int64
		_, err = h.redisClient.Get(ctx, reminderKey, &reminderCount)
		if reminderCount > 0 {
			student.ReminderCount = reminderCount
		}

		// 获取表扬次数 - 按类型统计
		praiseTypes := []string{
			string(consts.BehaviorTagTypeCorrectStreak),
			string(consts.BehaviorTagTypeEarlyLearn),
			string(consts.BehaviorTagTypeQuestion),
		}

		var praiseCount int64
		for _, behaviorType := range praiseTypes {
			typeKey := consts.GetPraiseTypeRecordKey(int64(classroomID), studentID, behaviorType)
			var timestamp string
			typeExists, _ := h.redisClient.Get(ctx, typeKey, &timestamp)
			if typeExists {
				praiseCount++
			}
		}
		student.PraiseCount = praiseCount

		studentBehaviorsMap[studentID] = &student
	}

	// 然后处理所有行为，累加统计值
	if allBehaviors != nil && len(allBehaviors) > 0 {
		for _, behavior := range allBehaviors {
			if behavior == nil {
				continue
			}

			studentID := behavior.StudentID
			student, exists := studentBehaviorsMap[studentID]
			if !exists {
				continue // 如果不在最新行为中，跳过
			}

			// 从Context中提取各项行为计数
			var contextMap map[string]interface{}
			if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
				// 提取提前学习次数、提问次数和连对次数（表扬类型）
				earlyLearnCount := utils.GetMapInt64Key(contextMap, "early_learn_count")
				if earlyLearnCount > 0 {
					student.EarlyLearnCount += earlyLearnCount
				}

				questionCount := utils.GetMapInt64Key(contextMap, "question_count")
				if questionCount > 0 {
					student.QuestionCount += questionCount
				}

				correctStreak := utils.GetMapInt64Key(contextMap, "correct_streak")
				if correctStreak > student.CorrectStreak {
					student.CorrectStreak = correctStreak
				}

				// 提取频繁切换页面次数、学习其他内容次数和停顿操作次数（关注类型）
				pageSwitchCount := utils.GetMapInt64Key(contextMap, "page_switch_count")
				if pageSwitchCount > 0 {
					student.PageSwitchCount += pageSwitchCount
				}

				otherContentCount := utils.GetMapInt64Key(contextMap, "other_content_count")
				if otherContentCount > 0 {
					student.OtherContentCount += otherContentCount
				}

				pauseCount := utils.GetMapInt64Key(contextMap, "pause_count")
				if pauseCount > 0 {
					student.PauseCount += pauseCount
				}
			}
		}
	}

	// 转换回切片
	categories := make([]dto.StudentBehaviorCategoryDTO, 0, len(studentBehaviorsMap))
	for _, student := range studentBehaviorsMap {
		// 生成行为标签
		generateBehaviorTags(student)
		categories = append(categories, *student)
	}

	return categories, nil
}

// generateBehaviorTags 为学生生成行为标签
func generateBehaviorTags(student *dto.StudentBehaviorCategoryDTO) {
	student.BehaviorTags = []dto.BehaviorTag{}

	// 课堂表现很棒的标签
	if student.EarlyLearnCount > 0 {
		student.BehaviorTags = append(student.BehaviorTags, dto.BehaviorTag{
			Type:  "earlyLearn",
			Count: student.EarlyLearnCount,
			Text:  fmt.Sprintf("提前学%d次", student.EarlyLearnCount),
		})
	}

	if student.QuestionCount > 0 {
		student.BehaviorTags = append(student.BehaviorTags, dto.BehaviorTag{
			Type:  "question",
			Count: student.QuestionCount,
			Text:  fmt.Sprintf("提问%d次", student.QuestionCount),
		})
	}

	if student.CorrectStreak > 0 {
		student.BehaviorTags = append(student.BehaviorTags, dto.BehaviorTag{
			Type:  "correctStreak",
			Count: student.CorrectStreak,
			Text:  fmt.Sprintf("连对%d次", student.CorrectStreak),
		})
	}

	// 需要关注的行为标签
	if student.PageSwitchCount > 0 {
		student.BehaviorTags = append(student.BehaviorTags, dto.BehaviorTag{
			Type:  "pageSwitch",
			Count: student.PageSwitchCount,
			Text:  fmt.Sprintf("频繁切换页面 %d次", student.PageSwitchCount),
		})
	}

	if student.OtherContentCount > 0 {
		student.BehaviorTags = append(student.BehaviorTags, dto.BehaviorTag{
			Type:  "otherContent",
			Count: student.OtherContentCount,
			Text:  fmt.Sprintf("学习其他内容 %d次", student.OtherContentCount),
		})
	}

	if student.PauseCount > 0 {
		student.BehaviorTags = append(student.BehaviorTags, dto.BehaviorTag{
			Type:  "pause",
			Count: student.PauseCount,
			Text:  fmt.Sprintf("停顿操作 %d次", student.PauseCount),
		})
	}
}

// isPraiseWorthy 判断学生行为是否值得表扬
func (h *BehaviorHandler) isPraiseWorthy(behavior *dto.StudentLatestBehaviorDTO) bool {
	if behavior == nil {
		return false
	}

	// 获取行为上下文
	var contextMap dto.BehaviorContext
	if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err != nil {
		h.logger.Error(nil, "解析行为上下文失败", "error", err)
		return false
	}

	// 获取相关指标
	correctStreak := contextMap.CorrectStreak
	earlyLearnCount := contextMap.EarlyLearnCount
	questionCount := contextMap.QuestionCount
	correctAnswers := contextMap.CorrectAnswers
	totalQuestions := contextMap.TotalQuestions
	learningType := contextMap.LearningType

	// 1. 连续答对类型检查
	if correctStreak >= consts.PraiseCheckCorrectStreak {
		h.logger.Debug(context.Background(), "表扬检查：连续答对达标", "连续答对次数", correctStreak)
		return true
	}

	// 2. 提前学习类型检查
	if earlyLearnCount >= consts.PraiseCheckEarlyLearnCountRequired || learningType == string(consts.LearningTypeSelfStudy) {
		h.logger.Debug(context.Background(), "表扬检查：提前学习达标", "提前学习次数", earlyLearnCount)
		return true
	}

	// 3. 提问类型检查
	if questionCount >= consts.PraiseCheckQuestionCountRequired {
		h.logger.Debug(context.Background(), "表扬检查：提问达标", "提问次数", questionCount)
		return true
	}

	// 4. 答题正确率检查
	if correctAnswers >= consts.PraiseCheckMinCorrectAnswers &&
		(totalQuestions == 0 || float64(correctAnswers)/float64(totalQuestions) >= consts.PraiseCheckCorrectRateThreshold) {
		h.logger.Debug(context.Background(), "表扬检查：答题正确率达标",
			"正确答题数", correctAnswers,
			"总题数", totalQuestions)
		return true
	}

	// 5. 学习时长检查
	stayDuration := float64(behavior.StayDuration)
	videoStatus := behavior.VideoStatus

	if stayDuration >= float64(consts.PraiseCheckMinStayDurationSeconds) && videoStatus != "pause" {
		h.logger.Debug(context.Background(), "表扬检查：学习时长达标",
			"学习时长(秒)", stayDuration,
			"视频状态", videoStatus)
		return true
	}

	return false
}

// needsAttention 判断学生是否需要关注
func (h *BehaviorHandler) needsAttention(behavior *dto.StudentLatestBehaviorDTO) bool {
	// 增加空指针检查
	if behavior == nil {
		return false
	}

	// 答题正确率低于50%
	if behavior.BehaviorType == consts.BehaviorTypeAnswer {
		if behavior.AccuracyRate < 50 && behavior.TotalQuestions >= 3 {
			return true
		}
	}

	// 视频暂停时间过长
	if behavior.BehaviorType == consts.BehaviorTypeLearning && behavior.VideoStatus == "pause" {
		// 从Context中直接获取暂停时长
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
			if stayDuration, ok := contextMap["stay_duration"].(float64); ok {
				if stayDuration >= consts.AttentionRuleVideoPauseTimeSeconds { // 使用常量替代硬编码的300秒
					return true
				}
			}
		} else if behavior.StayDuration >= consts.AttentionRuleVideoPauseTimeSeconds { // 使用常量替代硬编码的300秒
			return true
		}
	}

	return false
}

// generatePraiseDesc 生成表扬描述
func (h *BehaviorHandler) generatePraiseDesc(behavior *dto.StudentLatestBehaviorDTO) string {
	if behavior == nil {
		return consts.BehaviorDescDefaultPraiseSimple
	}

	if behavior.BehaviorType == consts.BehaviorTypeAnswer {
		// 从Context中提取更详细的答题信息
		var correctAnswers int64 = behavior.CorrectAnswers

		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
			if questionsInfo, ok := contextMap["questions_info"].(map[string]interface{}); ok {
				if correct, ok := questionsInfo["correct_answers"].(float64); ok {
					correctAnswers = int64(correct)
				}
			}
		}

		// 连对题目描述
		description := consts.FormatMessage(consts.BehaviorDescAnswerCorrectStreakSimple, correctAnswers)
		return description
	}

	if behavior.BehaviorType == consts.BehaviorTypeLearning {
		// 从Context中提取更详细的学习信息
		var stayDuration int64 = behavior.StayDuration

		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
			if duration, ok := contextMap["stay_duration"].(float64); ok {
				stayDuration = int64(duration)
			}
		}

		minutes := stayDuration / 60
		if minutes > 0 {
			// 专注学习描述
			return consts.BehaviorDescFocusedLearningSimple
		}
		return "专注学习中"
	}

	if behavior.BehaviorType == consts.BehaviorTypeQuestion {
		// 从Context中提取提问内容
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
			if _, ok := contextMap["question_content"].(string); ok {
				return consts.BehaviorDescActiveQuestioningSimple
			}
		}

		return "积极提问"
	}

	return consts.BehaviorDescDefaultPraiseSimple
}

// generateAttentionDesc 生成关注描述
func (h *BehaviorHandler) generateAttentionDesc(behavior *dto.StudentLatestBehaviorDTO) string {
	// 增加空指针检查
	if behavior == nil {
		return consts.BehaviorDescNeedAttention
	}

	if behavior.BehaviorType == consts.BehaviorTypeAnswer && behavior.TotalQuestions > 0 {
		// 从Context中提取更详细的答题信息
		var accuracyRate float64 = behavior.AccuracyRate
		var totalQuestions int64 = behavior.TotalQuestions
		var correctAnswers int64 = behavior.CorrectAnswers

		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
			// 从Context中更新这些信息（如果有的话）
			if questionsInfo, ok := contextMap["questions"].(map[string]interface{}); ok {
				if correct, ok := questionsInfo["correct_answers"].(float64); ok {
					correctAnswers = int64(correct)
				}
				if total, ok := questionsInfo["total_questions"].(float64); ok {
					totalQuestions = int64(total)
				}

				// 重新计算正确率
				accuracyRate = utils.F64Percent(float64(correctAnswers), float64(totalQuestions), 2)
			}
		}

		if totalQuestions > 0 {
			// 连对题目描述
			description := consts.FormatMessage(consts.BehaviorDescLowAccuracyRate, accuracyRate)
			return description
		}
		return consts.BehaviorDescAnsweringDifficulties
	}

	if behavior.BehaviorType == consts.BehaviorTypeLearning && behavior.VideoStatus == "pause" {
		// 从Context中提取更详细的学习信息
		var stayDuration int64 = behavior.StayDuration

		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
			if duration, ok := contextMap["stay_duration"].(float64); ok {
				stayDuration = int64(duration)
			}
		}

		minutes := stayDuration / 60
		if minutes > 0 {
			// 视频暂停描述
			description := consts.FormatMessage(consts.BehaviorDescVideoPausedMinutes, minutes)
			return description
		}
		return consts.BehaviorDescVideoPaused
	}

	// 检查其他问题
	var contextMap map[string]interface{}
	if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
		var problems []string

		// 检查频繁切换页面
		if pageSwitchCount := utils.GetMapValueI64(contextMap, "page_switch_count", 0); pageSwitchCount > 0 {
			problems = append(problems, consts.FormatMessage(consts.ProblemTplFrequentPageSwitch, pageSwitchCount))
		}

		// 检查学习其它内容
		if otherContentCount := utils.GetMapValueI64(contextMap, "other_content_count", 0); otherContentCount > 0 {
			problems = append(problems, consts.FormatMessage(consts.ProblemTplOtherContent, otherContentCount))
		}

		// 检查停顿操作
		if pauseCount := utils.GetMapValueI64(contextMap, "pause_count", 0); pauseCount > 0 {
			problems = append(problems, consts.FormatMessage(consts.ProblemTplPauseOperation, pauseCount))
		}

		// 生成提醒消息
		if len(problems) > 0 {
			return strings.Join(problems, "，")
		}
	}

	return consts.BehaviorDescNeedAttention
}

// PraiseStudents 表扬学生（单个或批量）
func (h *BehaviorHandler) PraiseStudents(ctx context.Context, req *api.PraiseStudentRequest) (*api.PraiseStudentResponse, error) {
	// 空指针检查
	if req == nil {
		return nil, errors.New("请求参数不能为空")
	}

	// 初始化响应
	response := &api.PraiseStudentResponse{
		Success:     true,
		Message:     consts.BehaviorResultPraiseSuccess,
		ClassroomID: req.ClassroomID,
		HandleTime:  time.Now().Unix(),
		Results:     make([]api.StudentHandleResult, 0, len(req.StudentIDs)),
	}

	// 设置所有可用的表扬类型
	allTypes := []string{
		string(consts.BehaviorTagTypeCorrectStreak),
		string(consts.BehaviorTagTypeEarlyLearn),
		string(consts.BehaviorTagTypeQuestion),
	}
	h.logger.Debug(ctx, "支持的表扬类型: %v", allTypes)

	// 直接使用req.StudentIDs，不进行类型转换，使用原始的uint64类型
	behaviors, err := h.behaviorDAO.GetStudentsBehaviors(ctx, req.StudentIDs)
	if err != nil {
		h.logger.Error(ctx, "获取学生行为数据失败: %v", err)
		return nil, errors.Wrap(err, "获取学生行为数据失败")
	}

	// 创建学生行为信息映射，方便查找
	studentBehaviors := make(map[uint64]*dto.StudentLatestBehaviorDTO)
	for _, behavior := range behaviors {
		// 添加nil检查
		if behavior == nil {
			h.logger.Warn(ctx, "发现空的学生行为记录，已跳过")
			continue
		}

		h.logger.Debug(ctx, "学生 %d 原始行为数据: 类型=%s, 上下文=%s",
			behavior.StudentID, string(behavior.BehaviorType), behavior.Context)

		studentBehaviors[uint64(behavior.StudentID)] = behavior
	}

	// 准备学生处理结果
	studentResults := make(map[uint64]*api.StudentHandleResult)
	// 记录每个学生选择的表扬类型
	selectedTypes := make(map[uint64]string)
	// 记录每个学生已经接收过的表扬类型
	studentPraisedTypes := make(map[uint64][]string)

	h.logger.Debug(ctx, "开始处理表扬请求，课堂ID: %d, 学生数量: %d", req.ClassroomID, len(req.StudentIDs))

	for _, studentID := range req.StudentIDs {
		studentName := ""

		// 从行为数据中获取学生信息
		if behavior, ok := studentBehaviors[studentID]; ok && behavior != nil {
			// 尝试从Context中提取学生姓名
			var contextMap map[string]interface{}
			if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
				if name, ok := contextMap["student_name"].(string); ok {
					studentName = name
				}
			} else {
				// 记录错误但不中断流程
				h.logger.Error(ctx, "解析学生%d行为上下文失败: %v, context=%s", studentID, err, behavior.Context)
			}
		}

		// 获取该学生剩余可用的表扬类型和已表扬的类型数量
		availableTypes := make([]string, 0, 3)
		praisedTypes := make([]string, 0, 3)
		praisedCount := 0

		// 首先检查总的表扬记录
		praiseKey := consts.GetPraiseRecordKey(int64(req.ClassroomID), int64(studentID))
		var generalTimestamp string
		generalPraiseExists, _ := h.redisClient.Get(ctx, praiseKey, &generalTimestamp)

		if generalPraiseExists {
			h.logger.Debug(ctx, "学生 %d 在课堂 %d 中有总表扬记录", studentID, req.ClassroomID)
		}

		// 检查各个表扬类型的redis记录
		for _, behaviorType := range allTypes {
			typeKey := consts.GetPraiseTypeRecordKey(int64(req.ClassroomID), int64(studentID), behaviorType)
			var timestamp string
			typeExists, err := h.redisClient.Get(ctx, typeKey, &timestamp)

			h.logger.Debug(ctx, "检查表扬记录 - 学生ID: %d, 类型: %s, 存在: %v, 错误: %v",
				studentID, behaviorType, typeExists, err)

			if typeExists {
				praisedCount++
				praisedTypes = append(praisedTypes, behaviorType)
			} else {
				availableTypes = append(availableTypes, behaviorType)
			}
		}

		// 记录该学生已接收过的表扬类型
		studentPraisedTypes[studentID] = praisedTypes

		h.logger.Debug(ctx, "学生 %d 表扬统计 - 已表扬类型: %v (%d种), 可用类型: %v",
			studentID, praisedTypes, praisedCount, availableTypes)

		// 检查是否已达到最大表扬次数(3种类型都已表扬)
		if praisedCount >= consts.MaxDailyPraiseCount {
			h.logger.Debug(ctx, "学生 %d 已达到最大表扬次数 (%d次)", studentID, consts.MaxDailyPraiseCount)
			// 已达到最大表扬次数
			studentResults[studentID] = &api.StudentHandleResult{
				StudentID:            studentID,
				StudentName:          studentName,
				Success:              false,
				Message:              consts.FormatMessage(consts.BehaviorResultMaxPraiseReached, consts.MaxDailyPraiseCount),
				AvailablePraiseTypes: []string{}, // 没有可用类型
			}
			continue
		}

		// 检查该学生是否满足表扬条件
		behavior, hasBehavior := studentBehaviors[studentID]
		isWorthy := false
		if hasBehavior && behavior != nil {
			isWorthy = h.isPraiseWorthy(behavior)
			h.logger.Debug(ctx, "学生 %d 表扬条件检查结果: %v", studentID, isWorthy)
		}

		if !hasBehavior || behavior == nil || !isWorthy {
			h.logger.Debug(ctx, "学生 %d 不满足表扬条件", studentID)
			studentResults[studentID] = &api.StudentHandleResult{
				StudentID:            studentID,
				StudentName:          studentName,
				Success:              false,
				Message:              consts.BehaviorResultNotPraiseWorthy,
				AvailablePraiseTypes: availableTypes,
			}
			continue
		}

		// 如果没有可用类型，但还没达到3种，这是一种错误状态
		if len(availableTypes) == 0 && praisedCount < consts.MaxDailyPraiseCount {
			h.logger.Error(ctx, "学生 %d 出现异常状态: 无可用表扬类型，但praisedCount=%d < %d",
				studentID, praisedCount, consts.MaxDailyPraiseCount)
			// 这种情况理论上不应该发生，但为了代码健壮性还是做个处理
			studentResults[studentID] = &api.StudentHandleResult{
				StudentID:            studentID,
				StudentName:          studentName,
				Success:              false,
				Message:              consts.BehaviorResultSystemError,
				AvailablePraiseTypes: []string{}, // 没有可用类型
			}
			continue
		}

		// 自动选择一个表扬类型
		// 首先检查学生行为，尝试选择最适合的类型
		selectedType := h.selectBestPraiseType(studentID, studentBehaviors, availableTypes)
		// 为每个学生单独记录选择的类型
		selectedTypes[studentID] = selectedType

		h.logger.Debug(ctx, "学生 %d 选择表扬类型: %s", studentID, selectedType)

		// 可以进行表扬
		studentResults[studentID] = &api.StudentHandleResult{
			StudentID:            studentID,
			StudentName:          studentName,
			Success:              true,
			Message:              consts.BehaviorResultPraiseSuccess,
			AvailablePraiseTypes: availableTypes,
		}
	}

	// 记录已处理学生
	now := time.Now().Unix()
	// 逐个处理已处理学生记录
	handleTimeMap := make(map[int64]int64)
	for studentID, result := range studentResults {
		if result.Success {
			handleTimeMap[int64(studentID)] = now

			// 获取为该学生选择的表扬类型
			selectedType := selectedTypes[studentID]

			// 记录表扬时间按行为类型
			praiseTypeKey := consts.GetPraiseTypeRecordKey(int64(req.ClassroomID), int64(studentID), selectedType)
			err = h.redisClient.Set(ctx, praiseTypeKey, now, consts.PraiseRecordExpire)
			if err != nil {
				h.logger.Error(ctx, "记录学生表扬时间(按类型)失败: %v", err)
				// 不影响主流程，继续执行
			} else {
				h.logger.Debug(ctx, "成功记录学生 %d 表扬时间(类型=%s)", studentID, selectedType)
			}

			// 同时更新总的表扬记录（兼容旧版本）
			praiseKey := consts.GetPraiseRecordKey(int64(req.ClassroomID), int64(studentID))
			err = h.redisClient.Set(ctx, praiseKey, now, consts.PraiseRecordExpire)
			if err != nil {
				h.logger.Error(ctx, "记录学生表扬时间失败: %v", err)
				// 不影响主流程，继续执行
			}

			// 由于已成功处理，更新可用类型列表（从中移除当前类型）
			newAvailableTypes := make([]string, 0)
			for _, t := range allTypes {
				// 检查该类型是否未被使用过（既不是当前选择的类型，也不是之前已表扬的类型）
				alreadyPraised := false
				for _, pt := range studentPraisedTypes[studentID] {
					if t == pt {
						alreadyPraised = true
						break
					}
				}

				if t != selectedType && !alreadyPraised {
					newAvailableTypes = append(newAvailableTypes, t)
				}
			}
			result.AvailablePraiseTypes = newAvailableTypes

			h.logger.Debug(ctx, "学生 %d 表扬结果: 当前表扬: %s, 历史表扬: %v, 剩余可用类型: %v",
				studentID, selectedType, studentPraisedTypes[studentID], newAvailableTypes)
		}
	}
	handledKey := consts.GetClassBehaviorHandledKey(int64(req.ClassroomID))
	h.redisClient.HSet(ctx, handledKey, handleTimeMap, consts.ClassBehaviorHandledExpire)

	// 发送表扬通知前，将选择的类型保存到请求对象，以便通知处理
	h.sendPraiseNotifications(ctx, req, studentResults, studentBehaviors, selectedTypes)

	// 检查是否所有学生都表扬失败
	allFailed := true
	for _, result := range studentResults {
		if result.Success {
			allFailed = false
			break
		}
	}

	// 如果所有学生都表扬失败且至少有一个学生，设置整体响应状态为失败
	if allFailed && len(studentResults) > 0 {
		response.Success = false
		response.Message = consts.BehaviorResultAllPraiseFailed
	}

	// 填充结果列表
	for _, result := range studentResults {
		response.Results = append(response.Results, *result)
	}

	return response, nil
}

// selectBestPraiseType 根据学生行为自动选择最佳表扬类型
func (h *BehaviorHandler) selectBestPraiseType(studentID uint64, behaviors map[uint64]*dto.StudentLatestBehaviorDTO, availableTypes []string) string {
	// 如果没有可用类型，返回默认值
	if len(availableTypes) == 0 {
		h.logger.Debug(context.Background(), "学生 %d 没有可用表扬类型，返回默认类型", studentID)
		return string(consts.BehaviorTagTypeCorrectStreak) // 默认选择连续答对
	}

	// 如果学生行为数据不存在，返回第一个可用类型
	behavior, ok := behaviors[studentID]
	if !ok {
		h.logger.Debug(context.Background(), "学生 %d 没有行为数据，随机选择第一个可用类型: %s", studentID, availableTypes[0])
		return availableTypes[0]
	}

	// 解析学生行为上下文
	var contextMap map[string]interface{}
	if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err != nil {
		h.logger.Debug(context.Background(), "学生 %d 行为上下文解析失败: %v，选择第一个可用类型: %s",
			studentID, err, availableTypes[0])
		return availableTypes[0]
	}

	h.logger.Debug(context.Background(), "学生 %d 行为数据: 类型=%s, 上下文=%+v, 可用表扬类型=%v",
		studentID, behavior.BehaviorType, contextMap, availableTypes)

	// 各类型的匹配分数
	typeScores := make(map[string]int)
	for _, t := range availableTypes {
		typeScores[t] = consts.PraiseTypeInitialBaseScore // 初始化基础得分为5，避免零分状态
	}

	// 从上下文中提取所有可能的指标
	correctStreak := int(utils.GetMapFloat64Key(contextMap, "correct_streak"))
	earlyLearnCount := int(utils.GetMapFloat64Key(contextMap, "early_learn_count"))
	questionCount := int(utils.GetMapFloat64Key(contextMap, "question_count"))
	totalQuestions := int(utils.GetMapFloat64Key(contextMap, "total_questions"))
	correctAnswers := int(utils.GetMapFloat64Key(contextMap, "correct_answers"))
	stayDuration := int(behavior.StayDuration)
	learningType := utils.GetMapStringKey(contextMap, "learning_type")

	// 1. 评估连续答对类型
	if slices.Contains(availableTypes, string(consts.BehaviorTagTypeCorrectStreak)) {
		score := typeScores[string(consts.BehaviorTagTypeCorrectStreak)]

		// 连对越多分数越高
		if correctStreak > 0 {
			score += correctStreak * consts.PraiseCorrectStreakScoreWeight // 提高连对的权重
		}

		// 答题正确率高也加分
		if totalQuestions > 0 && correctAnswers > 0 {
			accuracyRate := float64(correctAnswers) / float64(totalQuestions) * 100
			if accuracyRate >= consts.PraiseAccuracyRateThresholdForBonus {
				score += consts.PraiseAccuracyRateBonusScore // 正确率高于80%加10分
			}
		}

		// 行为类型是答题也加分
		if behavior.BehaviorType == consts.BehaviorTypeAnswer {
			score += consts.PraiseAnswerTypeBonusScore
		}

		typeScores[string(consts.BehaviorTagTypeCorrectStreak)] = score
	}

	// 2. 评估主动提问类型
	if slices.Contains(availableTypes, string(consts.BehaviorTagTypeQuestion)) {
		score := typeScores[string(consts.BehaviorTagTypeQuestion)]

		// 提问次数加高分
		if questionCount > 0 {
			score += questionCount * consts.PraiseQuestionCountScoreWeight // 提高提问的权重
		}

		// 行为类型是提问也加高分
		if behavior.BehaviorType == consts.BehaviorTypeQuestion {
			score += consts.PraiseQuestionTypeBonusScore
		}

		typeScores[string(consts.BehaviorTagTypeQuestion)] = score
	}

	// 3. 评估提前学习类型
	if slices.Contains(availableTypes, string(consts.BehaviorTagTypeEarlyLearn)) {
		score := typeScores[string(consts.BehaviorTagTypeEarlyLearn)]

		// 提前学习次数加高分
		if earlyLearnCount > 0 {
			score += earlyLearnCount * consts.PraiseEarlyLearnCountScoreWeight
		}

		// 学习时间较长加分
		if stayDuration > consts.PraiseStayDurationThresholdMinutesForBonus*60 { // 超过15分钟
			score += int(stayDuration/60/consts.PraiseStayDurationBonusIntervalMinutes)*consts.PraiseStayDurationBonusScorePerInterval + consts.PraiseLongStayDurationBaseBonusScore // 每5分钟加1分，基础加10分
		}

		// 学习类型是自学加高分
		if learningType == consts.LearningTypeSelfStudy {
			score += consts.PraiseSelfStudyLearningTypeBonusScore
		}

		// 行为类型是学习也加分
		if behavior.BehaviorType == consts.BehaviorTypeLearning {
			score += consts.PraiseLearningBehaviorTypeBonusScore
		}

		typeScores[string(consts.BehaviorTagTypeEarlyLearn)] = score
	}

	// 打印每种类型的得分
	for t, score := range typeScores {
		h.logger.Debug(context.Background(), "学生 %d 表扬类型评分: 类型=%s, 得分=%d", studentID, t, score)
	}

	// 找出得分最高的类型
	var bestType string
	bestScore := -1

	// 随机打乱顺序，避免相同分数时总是选择同一个类型
	keys := make([]string, 0, len(typeScores))
	for t := range typeScores {
		keys = append(keys, t)
	}
	for _, t := range keys {
		score := typeScores[t]
		if score > bestScore {
			bestScore = score
			bestType = t
		}
	}

	// 如果没有找到明显最佳类型，返回第一个可用类型
	if bestScore <= consts.PraiseBestTypeMinScoreThreshold { // 只有初始得分
		h.logger.Debug(context.Background(), "学生 %d 没有找到明显最佳类型，选择第一个可用类型: %s", studentID, availableTypes[0])
		return availableTypes[0]
	}

	h.logger.Debug(context.Background(), "学生 %d 选择最佳表扬类型: %s, 得分=%d", studentID, bestType, bestScore)
	return bestType
}

// getBehaviorTypeDesc 获取行为类型的描述文本
func getBehaviorTypeDesc(behaviorType string) string {
	switch behaviorType {
	case string(consts.BehaviorTagTypeCorrectStreak):
		return consts.BehaviorDescTypeCorrectStreak
	case string(consts.BehaviorTagTypeEarlyLearn):
		return consts.BehaviorDescTypeEarlyLearn
	case string(consts.BehaviorTagTypeQuestion):
		return consts.BehaviorDescTypeQuestion
	default:
		return behaviorType
	}
}

// 发送表扬通知
func (h *BehaviorHandler) sendPraiseNotifications(ctx context.Context, req *api.PraiseStudentRequest,
	results map[uint64]*api.StudentHandleResult, behaviors map[uint64]*dto.StudentLatestBehaviorDTO,
	selectedTypes map[uint64]string) {

	// 构建消息队列请求
	for studentID, result := range results {
		if !result.Success {
			continue // 跳过处理失败的学生
		}

		// 获取为该学生选择的表扬类型
		selectedType := selectedTypes[studentID]
		h.logger.Debug(ctx, "为学生 %d 发送表扬通知，类型: %s", studentID, selectedType)

		// 根据表扬类型生成个性化表扬消息
		var praiseMsg string

		// 首先根据行为数据生成消息
		behavior := behaviors[studentID]
		if behavior != nil {
			// 解析行为上下文
			var contextMap map[string]interface{}
			if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
				// 根据表扬类型生成特定消息
				switch selectedType {
				case string(consts.BehaviorTagTypeCorrectStreak):
					// 连续答对表扬
					correctStreak := int(utils.GetMapFloat64Key(contextMap, "correct_streak"))
					if correctStreak >= consts.PraiseCheckCorrectStreak { // 使用配置项
						praiseMsg = fmt.Sprintf("太棒了！你已经连续答对 %d 题了！", correctStreak)
					} else {
						praiseMsg = consts.BehaviorDescPraiseGeneralCorrectAnswer // 使用常量
					}

				case string(consts.BehaviorTagTypeEarlyLearn):
					// 提前学习表扬
					earlyLearnCount := int(utils.GetMapFloat64Key(contextMap, "early_learn_count"))
					if earlyLearnCount > 0 {
						praiseMsg = fmt.Sprintf(consts.BehaviorDescPraiseEarlyLearnSpecific, earlyLearnCount)
					} else {
						praiseMsg = consts.BehaviorDescPraiseEarlyLearnGeneral
					}

				case string(consts.BehaviorTagTypeQuestion):
					// 提问表扬
					questionCount := int(utils.GetMapFloat64Key(contextMap, "question_count"))
					if questionCount > 0 {
						praiseMsg = consts.BehaviorDescPraiseQuestionSpecific
					} else {
						praiseMsg = consts.BehaviorDescPraiseQuestionGeneral
					}

				default:
					// 默认表扬文案
					praiseMsg = consts.BehaviorDescDefaultPraiseSimple
				}
			}
		}

		// 如果没有根据类型生成特定消息，则使用通用方法
		if praiseMsg == "" {
			praiseMsg = h.generatePraiseDesc(behavior)
		}

		// 记录表扬行为
		h.logger.Debug(ctx, "向学生%d发送表扬通知: %s, 类型: %s", studentID, praiseMsg, selectedType)

		contextData := map[string]interface{}{
			"studentId":    studentID,
			"message":      praiseMsg,
			"behaviorType": selectedType,
			"praiseType":   selectedType, // 额外加入表扬类型，便于前端显示
		}
		contextBytes, err := json.Marshal(contextData)
		if err != nil {
			contextBytes = []byte("{}")
		}

		// 创建教师行为记录
		behaviorReq := &api.TeacherBehaviorRequest{
			SchoolID:     uint64(req.SchoolID),
			TeacherID:    uint64(req.TeacherID),
			ClassID:      uint64(req.ClassID),
			ClassroomID:  uint64(req.ClassroomID),
			BehaviorType: string(consts.BehaviorTypePraise),
			Context:      string(contextBytes), // 使用序列化后的JSON字符串
		}

		// 使用行为DAO直接保存行为记录
		teacherBehavior := &dto.TeacherBehaviorDTO{
			SchoolID:     behaviorReq.SchoolID,
			ClassID:      behaviorReq.ClassID,
			ClassroomID:  &behaviorReq.ClassroomID,
			TeacherID:    behaviorReq.TeacherID,
			BehaviorType: consts.BehaviorType(behaviorReq.BehaviorType),
			Context:      behaviorReq.Context,
			CreateTime:   time.Now(),
		}

		// 保存教师行为
		err = h.behaviorDAO.SaveTeacherBehavior(ctx, []*dto.TeacherBehaviorDTO{teacherBehavior})
		if err != nil {
			h.logger.Error(ctx, "保存表扬行为失败: %v", err)
		} else {
			h.logger.Debug(ctx, "成功保存学生 %d 的表扬行为记录", studentID)
		}
	}
}

// AttentionStudents 关注学生（单个或批量）
func (h *BehaviorHandler) AttentionStudents(ctx context.Context, req *api.AttentionStudentRequest) (*api.AttentionStudentResponse, error) {
	// 空指针检查
	if req == nil {
		return nil, errors.New("请求参数不能为空")
	}

	// 初始化响应
	response := &api.AttentionStudentResponse{
		ClassroomID: req.ClassroomID,
		Success:     true,
		Message:     consts.BehaviorResultAttentionSuccess,
		HandleTime:  time.Now().Unix(),
		Results:     make([]api.StudentHandleResult, 0, len(req.StudentIDs)),
	}

	// 获取学生行为数据
	behaviors, err := h.behaviorDAO.GetStudentsBehaviors(ctx, req.StudentIDs)
	if err != nil {
		h.logger.Error(ctx, "获取学生行为数据失败: %v", err)
		return nil, errors.Wrap(err, "获取学生行为数据失败")
	}

	// 创建学生行为信息映射，方便查找
	studentBehaviors := make(map[uint64]*dto.StudentLatestBehaviorDTO)
	for _, behavior := range behaviors {
		if behavior == nil {
			h.logger.Warn(ctx, "发现空的学生行为记录，已跳过")
			continue
		}

		// 创建DTO结构
		behaviorDTO := &dto.StudentLatestBehaviorDTO{
			StudentID:      int64(behavior.StudentID),
			BehaviorType:   consts.BehaviorType(behavior.BehaviorType),
			Context:        behavior.Context,
			LastUpdateTime: time.Now().Unix(), // 使用当前时间代替CreateTime
		}

		// 解析上下文数据
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
			// 只在成功解析JSON时提取字段
			behaviorDTO.StudentName = utils.GetMapStringKey(contextMap, "student_name")
			behaviorDTO.AvatarURL = utils.GetMapStringKey(contextMap, "avatar_url")
		} else {
			h.logger.Error(ctx, "解析行为上下文失败: %v, context: %s", err, behavior.Context)
		}

		studentBehaviors[uint64(behaviorDTO.StudentID)] = behaviorDTO
	}

	// 可用的关注类型列表
	allTypes := []string{
		string(consts.BehaviorTagTypePageSwitch),
		string(consts.BehaviorTagTypeOtherContent),
		string(consts.BehaviorTagTypePause),
	}

	// 准备学生处理结果
	studentResults := make(map[uint64]*api.StudentHandleResult)
	// 记录每个学生选择的关注类型
	selectedTypes := make(map[uint64]string)

	h.logger.Debug(ctx, "开始处理关注请求，课堂ID: %d, 学生数量: %d", req.ClassroomID, len(req.StudentIDs))

	// 记录已处理学生
	handledKey := consts.GetClassBehaviorHandledKey(int64(req.ClassroomID))
	now := time.Now().Unix()
	handleTimeMap := make(map[int64]int64)

	// 处理每个学生
	for _, studentID := range req.StudentIDs {
		studentName := ""

		// 从行为数据中获取学生信息
		behavior, ok := studentBehaviors[studentID]
		if ok && behavior != nil {
			studentName = behavior.StudentName
		}

		result := &api.StudentHandleResult{
			StudentID:   studentID,
			StudentName: studentName,
			Success:     true,
		}

		// 选择最合适的关注类型（仅用于消息生成，不限制多次提醒）
		selectedType := h.selectBestAttentionType(ctx, studentID, studentBehaviors, allTypes)
		selectedTypes[studentID] = selectedType

		// 获取提醒次数
		reminderKey := consts.GetStudentReminderCountKey(int64(req.ClassroomID), int64(studentID))
		reminderCount, _ := h.redisClient.Incr(ctx, reminderKey, consts.StudentReminderCountExpire)

		// 生成关注消息
		if ok && behavior != nil {
			// 从类型选择生成对应的提醒消息
			attentionDesc := h.generateAttentionDescForType(behavior, selectedType)
			result.Message = consts.FormatMessage(consts.MsgTplKeepFocusedWithProblems, attentionDesc, strconv.FormatInt(reminderCount, 10))
		} else {
			result.Message = consts.FormatMessage(consts.MsgTplKeepFocused, strconv.FormatInt(reminderCount, 10))
		}

		// 设置处理结果
		studentResults[studentID] = result

		// 记录处理时间
		handleTimeMap[int64(studentID)] = now

		response.Results = append(response.Results, *result)
	}

	// 更新处理时间
	h.redisClient.HSet(ctx, handledKey, handleTimeMap, consts.ClassBehaviorHandledExpire)

	// 发送教师行为记录
	h.sendAttentionNotifications(ctx, req, studentResults, studentBehaviors, selectedTypes)

	return response, nil
}

// selectBestAttentionType 根据学生行为自动选择最佳关注类型
func (h *BehaviorHandler) selectBestAttentionType(ctx context.Context, studentID uint64, behaviors map[uint64]*dto.StudentLatestBehaviorDTO, availableTypes []string) string {
	// 如果没有可用类型，返回默认值
	if len(availableTypes) == 0 {
		h.logger.Debug(ctx, "学生 %d 没有可用关注类型，返回默认类型", studentID)
		return string(consts.BehaviorTagTypePageSwitch) // 默认选择频繁切换页面
	}

	// 如果学生行为数据不存在，返回第一个可用类型
	behavior, ok := behaviors[studentID]
	if !ok {
		h.logger.Debug(ctx, "学生 %d 没有行为数据，随机选择第一个可用类型: %s", studentID, availableTypes[0])
		return availableTypes[0]
	}

	// 解析学生行为上下文
	var contextMap map[string]interface{}
	if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err != nil {
		h.logger.Debug(ctx, "学生 %d 行为上下文解析失败: %v，选择第一个可用类型: %s",
			studentID, err, availableTypes[0])
		return availableTypes[0]
	}

	h.logger.Debug(ctx, "学生 %d 行为数据: 类型=%s, 上下文=%+v, 可用关注类型=%v",
		studentID, behavior.BehaviorType, contextMap, availableTypes)

	// 各类型的匹配分数
	typeScores := make(map[string]int)
	for _, t := range availableTypes {
		typeScores[t] = 0 // 初始化得分
	}

	// 1. 评估频繁切换页面类型
	if slices.Contains(availableTypes, string(consts.BehaviorTagTypePageSwitch)) {
		score := 0

		// 如果有页面切换记录，加分
		if pageSwitchCount, ok := contextMap["page_switch_count"].(float64); ok && pageSwitchCount > 0 {
			score += consts.AttentionRulePageSwitchBaseScore
			// 切换次数越多分数越高
			score += int(pageSwitchCount) * consts.AttentionRulePageSwitchScorePerCount
		}

		typeScores[string(consts.BehaviorTagTypePageSwitch)] = score
	}

	// 2. 评估学习其他内容类型
	if slices.Contains(availableTypes, string(consts.BehaviorTagTypeOtherContent)) {
		score := 0

		// 如果有学习其他内容记录，加分
		if otherContentCount, ok := contextMap["other_content_count"].(float64); ok && otherContentCount > 0 {
			score += consts.AttentionRuleOtherContentBaseScore
			// 次数越多分数越高
			score += int(otherContentCount) * consts.AttentionRuleOtherContentScorePerCount
		}

		typeScores[string(consts.BehaviorTagTypeOtherContent)] = score
	}

	// 3. 评估停顿操作类型
	if slices.Contains(availableTypes, string(consts.BehaviorTagTypePause)) {
		score := 0

		// 如果是学习行为且视频处于暂停状态，加高分
		if behavior.BehaviorType == consts.BehaviorTypeLearning && behavior.VideoStatus == "pause" {
			score += consts.AttentionRuleVideoPauseLearningScore
		}

		// 如果有暂停记录，加分
		if pauseCount, ok := contextMap["pause_count"].(float64); ok && pauseCount > 0 {
			score += consts.AttentionRulePauseCountBaseScore
			// 暂停次数越多分数越高
			score += int(pauseCount) * consts.AttentionRulePauseCountScorePerCount
		}

		// 如果有停留时间记录且较长，加分
		if stayDuration, ok := contextMap["stay_duration"].(float64); ok && stayDuration > consts.AttentionRuleVideoPauseTimeSeconds {
			score += consts.AttentionRuleBaseScoreForLongPause
			// 停留时间越长分数越高
			score += int(stayDuration/60) * consts.AttentionRuleScorePerMinutePaused
		}

		typeScores[string(consts.BehaviorTagTypePause)] = score
	}

	// 打印每种类型的得分
	for t, score := range typeScores {
		h.logger.Debug(ctx, "学生 %d 关注类型评分: 类型=%s, 得分=%d", studentID, t, score)
	}

	// 找出得分最高的类型
	var bestType string
	bestScore := -1

	for t, score := range typeScores {
		if score > bestScore {
			bestScore = score
			bestType = t
		}
	}

	// 如果没有找到明显最佳类型，返回第一个可用类型
	if bestScore <= 0 {
		h.logger.Debug(ctx, "学生 %d 没有找到明显最佳关注类型，选择第一个可用类型: %s", studentID, availableTypes[0])
		return availableTypes[0]
	}

	h.logger.Debug(ctx, "学生 %d 选择最佳关注类型: %s, 得分=%d", studentID, bestType, bestScore)
	return bestType
}

// generateAttentionDescForType 为特定关注类型生成关注描述
func (h *BehaviorHandler) generateAttentionDescForType(behavior *dto.StudentLatestBehaviorDTO, attentionType string) string {
	if behavior == nil {
		return consts.BehaviorDescNeedAttention
	}

	var contextMap map[string]interface{}
	if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err != nil {
		return consts.BehaviorDescNeedAttention
	}

	switch attentionType {
	case string(consts.BehaviorTagTypePageSwitch):
		if pageSwitchCount, ok := contextMap["page_switch_count"].(float64); ok && pageSwitchCount > 0 {
			return fmt.Sprintf("频繁切换页面 %d 次，请保持专注", int(pageSwitchCount))
		}
		return consts.AttentionDescFrequentSwitchFocus

	case string(consts.BehaviorTagTypeOtherContent):
		if otherContentCount, ok := contextMap["other_content_count"].(float64); ok && otherContentCount > 0 {
			return fmt.Sprintf("浏览其他内容 %d 次，请回到学习页面", int(otherContentCount))
		}
		return consts.AttentionDescIrrelevantContentFocus

	case string(consts.BehaviorTagTypePause):
		if behavior.BehaviorType == consts.BehaviorTypeLearning && behavior.VideoStatus == "pause" {
			if stayDuration, ok := contextMap["stay_duration"].(float64); ok && stayDuration > 0 {
				minutes := int(stayDuration) / 60
				if minutes > 0 {
					return fmt.Sprintf("视频已暂停 %d 分钟，需要帮助吗？", minutes)
				}
			}
			return consts.AttentionDescVideoPausedNeedHelp
		}
		if pauseCount, ok := contextMap["pause_count"].(float64); ok && pauseCount > 0 {
			return fmt.Sprintf("多次暂停操作 %d 次，请专注学习", int(pauseCount))
		}
		return consts.AttentionDescFrequentPausesFocus
	}

	return consts.BehaviorDescNeedAttention
}

// getAttentionTypeDesc 获取关注类型的描述文本
func getAttentionTypeDesc(attentionType string) string {
	switch attentionType {
	case string(consts.BehaviorTagTypePageSwitch):
		return consts.BehaviorDescTypePageSwitch
	case string(consts.BehaviorTagTypeOtherContent):
		return consts.BehaviorDescTypeOtherContent
	case string(consts.BehaviorTagTypePause):
		return consts.BehaviorDescTypePause
	default:
		return attentionType
	}
}

// 发送关注通知
func (h *BehaviorHandler) sendAttentionNotifications(ctx context.Context, req *api.AttentionStudentRequest,
	results map[uint64]*api.StudentHandleResult, behaviors map[uint64]*dto.StudentLatestBehaviorDTO,
	selectedTypes map[uint64]string) {

	// 构建消息队列请求
	for studentID, result := range results {
		if !result.Success {
			continue // 跳过处理失败的学生
		}

		// 获取提醒次数
		reminderKey := consts.GetStudentReminderCountKey(int64(req.ClassroomID), int64(studentID))
		var reminderCount int64
		h.redisClient.Get(ctx, reminderKey, &reminderCount)

		// 记录关注行为
		h.logger.Debug(ctx, "向学生%d发送关注通知: %s", studentID, result.Message)

		attentionType, ok := selectedTypes[studentID]
		if !ok {
			attentionType = "" // 若未指定类型，使用空字符串
		}

		contextData := map[string]interface{}{
			"studentId":     studentID,
			"message":       result.Message,
			"attentionType": attentionType, // 使用为该学生选择的类型
			"reminderCount": reminderCount,
		}
		contextBytes, err := json.Marshal(contextData)
		if err != nil {
			contextBytes = []byte("{}")
		}

		// 创建教师行为记录
		teacherBehavior := &dto.TeacherBehaviorDTO{
			SchoolID:     uint64(req.SchoolID),
			TeacherID:    uint64(req.TeacherID),
			ClassroomID:  &req.ClassroomID,
			BehaviorType: consts.BehaviorTypeAttention,
			Context:      string(contextBytes), // 使用序列化后的JSON字符串
			CreateTime:   time.Now(),
		}

		// 保存教师行为
		err = h.behaviorDAO.SaveTeacherBehavior(ctx, []*dto.TeacherBehaviorDTO{teacherBehavior})
		if err != nil {
			h.logger.Error(ctx, "保存关注行为失败: %v", err)
		}
	}
}

// GetClassroomLearningScores 获取单节课程的学习分列表
func (h *BehaviorHandler) GetClassroomLearningScores(ctx context.Context, classroomID uint64) ([]*dto.StudentLearningScoreDTO, error) {
	// 参数校验
	if classroomID == 0 {
		return nil, errors.New("课堂ID不能为0")
	}

	// 获取学生行为数据
	behaviors, err := h.behaviorDAO.GetClassLatestBehaviors(ctx, classroomID)
	if err != nil {
		h.logger.Error(ctx, "获取课堂学习分列表失败: %v", err)
		return nil, errors.Wrap(err, "获取课堂学习分列表失败")
	}

	// 检查behaviors是否为nil
	if behaviors == nil {
		h.logger.Error(ctx, "获取课堂学习分数据为空")
		return []*dto.StudentLearningScoreDTO{}, nil
	}

	// 定义结果
	results := make([]*dto.StudentLearningScoreDTO, 0, len(behaviors))

	// 遍历行为数据，生成学习分数据
	for _, behavior := range behaviors {
		// 确保behavior不为nil
		if behavior == nil {
			h.logger.Warn(ctx, "发现空的学生行为记录，已跳过")
			continue
		}

		// 计算学习分数：学习时间占30%，答题正确率占70%
		timeScore := utils.F64Min(float64(behavior.StayDuration)/60, 30)
		accuracyScore := utils.F64Percent(float64(behavior.CorrectAnswers), float64(behavior.TotalQuestions), 2) * 70
		totalScore := int64(timeScore + accuracyScore)

		// 获取头像URL和学生姓名
		studentName := ""
		avatarURL := ""
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
			if name, ok := contextMap["student_name"].(string); ok {
				studentName = name
			}
			if avatar, ok := contextMap["avatar_url"].(string); ok {
				avatarURL = avatar
			}
		}

		results = append(results, &dto.StudentLearningScoreDTO{
			StudentID:     uint64(behavior.StudentID),
			StudentName:   studentName,
			AvatarURL:     avatarURL,
			LearningScore: totalScore,
			LearningTime:  uint64(behavior.StayDuration),
			CorrectCount:  uint64(behavior.CorrectAnswers),
			TotalCount:    uint64(behavior.TotalQuestions),
		})
	}

	// 根据学习分数对结果进行降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].LearningScore > results[j].LearningScore
	})

	h.logger.Debug(ctx, "获取课堂(%d)学习分列表成功，共%d条记录", classroomID, len(results))
	return results, nil
}

// GetClassroomBehaviorSummary 获取课后行为汇总统计
func (h *BehaviorHandler) GetClassroomBehaviorSummary(ctx context.Context, classroomID uint64) (*api.ClassroomBehaviorSummaryResponse, error) {
	// 参数校验
	if classroomID == 0 {
		return nil, errors.New("课堂ID不能为0")
	}

	// 获取课堂学生行为数据
	behaviors, err := h.behaviorDAO.GetClassAllBehaviors(ctx, classroomID)
	if err != nil {
		h.logger.Error(ctx, "获取课堂学生行为数据失败: %v", err)
		return nil, errors.Wrap(err, "获取课堂学生行为数据失败")
	}

	// 检查behaviors是否为nil
	if behaviors == nil {
		h.logger.Error(ctx, "获取课堂行为数据为空")
		// 返回空结果
		return &api.ClassroomBehaviorSummaryResponse{
			ClassroomID:   classroomID,
			StatTime:      time.Now().Unix(),
			TotalStudents: 0,
			AvgAccuracy:   0,
			AvgProgress:   0,
			TotalDuration: 0,
			PraiseList:    []api.StudentBehaviorCategory{},
			AttentionList: []api.StudentBehaviorCategory{},
			AllStudents:   []api.StudentBehaviorCategory{},
		}, nil
	}

	// 按学生ID对行为进行分组
	studentBehaviorsMap := make(map[int64][]*dto.StudentLatestBehaviorDTO)
	for _, behavior := range behaviors {
		// 确保behavior不为nil
		if behavior == nil {
			h.logger.Warn(ctx, "发现空的学生行为记录，已跳过")
			continue
		}
		studentBehaviorsMap[behavior.StudentID] = append(studentBehaviorsMap[behavior.StudentID], behavior)
	}

	// 计算班级总数据
	totalStudents := int64(len(studentBehaviorsMap))
	var totalAccuracy float64
	var totalProgress float64
	var totalDuration int64
	var accuracyCount int // 有答题数据的学生数量

	// 转换为行为分类DTO
	studentCategories := make([]dto.StudentBehaviorCategoryDTO, 0, totalStudents)

	for studentID, studentBehaviors := range studentBehaviorsMap {
		// 确保学生行为列表不为空
		if len(studentBehaviors) == 0 {
			h.logger.Warn(ctx, "学生 %d 的行为记录为空，已跳过", studentID)
			continue
		}

		// 对每个学生的行为进行汇总
		var studentCategory dto.StudentBehaviorCategoryDTO
		studentCategory.StudentID = uint64(studentID)

		// 取最新行为的一些基本信息
		latestBehavior := studentBehaviors[0]
		if latestBehavior == nil {
			h.logger.Warn(ctx, "学生 %d 的最新行为记录为空，已跳过", studentID)
			continue
		}

		studentCategory.LastUpdateTime = latestBehavior.LastUpdateTime

		// 从Context中提取学生信息
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(latestBehavior.Context), &contextMap); err == nil {
			studentCategory.StudentName = utils.GetMapStringKey(contextMap, "student_name")
			studentCategory.AvatarUrl = utils.GetMapStringKey(contextMap, "avatar_url")
			studentCategory.LearningProgress = utils.GetMapFloat64Key(contextMap, "learning_progress")
		}

		// 统计各类行为次数
		var totalQuestions int64
		var correctAnswers int64
		var stayDuration int64

		for _, behavior := range studentBehaviors {
			// 确保behavior不为nil
			if behavior == nil {
				h.logger.Warn(ctx, "学生 %d 存在空的行为记录，已跳过", studentID)
				continue
			}

			// 答题行为
			if behavior.TotalQuestions > 0 {
				totalQuestions += behavior.TotalQuestions
				correctAnswers += behavior.CorrectAnswers
			}

			// 停留时长
			stayDuration += behavior.StayDuration

			// 提取特殊行为计数
			var contextMap map[string]interface{}
			if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
				// 累加行为计数
				studentCategory.EarlyLearnCount += int64(utils.GetMapIntKey(contextMap, "early_learn_count", 0))
				studentCategory.QuestionCount += int64(utils.GetMapIntKey(contextMap, "question_count", 0))
				studentCategory.PageSwitchCount += int64(utils.GetMapIntKey(contextMap, "page_switch_count", 0))
				studentCategory.OtherContentCount += int64(utils.GetMapIntKey(contextMap, "other_content_count", 0))
				studentCategory.PauseCount += int64(utils.GetMapIntKey(contextMap, "pause_count", 0))

				// 对于连对次数，取最大值
				correctStreak := int64(utils.GetMapIntKey(contextMap, "correct_streak", 0))
				if correctStreak > studentCategory.CorrectStreak {
					studentCategory.CorrectStreak = correctStreak
				}
			}
		}

		// 设置行为统计数据
		studentCategory.TotalQuestions = totalQuestions
		studentCategory.CorrectAnswers = correctAnswers
		studentCategory.WrongAnswers = totalQuestions - correctAnswers

		// 计算正确率
		if totalQuestions > 0 {
			studentCategory.AccuracyRate = utils.F64Percent(float64(correctAnswers), float64(totalQuestions)*100, 2)
			totalAccuracy += studentCategory.AccuracyRate
			accuracyCount++
		}

		// 生成行为标签
		generateBehaviorTags(&studentCategory)

		// 添加到列表
		studentCategories = append(studentCategories, studentCategory)

		// 累加到班级总计
		totalProgress += studentCategory.LearningProgress
		totalDuration += stayDuration
	}

	// 计算班级平均值
	avgAccuracy := utils.F64Div(totalAccuracy, float64(accuracyCount), 2)
	avgProgress := utils.F64Div(totalProgress, float64(totalStudents), 2)

	// 将学生行为按类型分组
	praiseList := make([]api.StudentBehaviorCategory, 0)
	attentionList := make([]api.StudentBehaviorCategory, 0)
	allStudents := make([]api.StudentBehaviorCategory, len(studentCategories))

	for i, student := range studentCategories {
		apiStudent := convertToAPIBehaviorCategory(student)
		allStudents[i] = apiStudent

		// 分离表扬标签和提醒标签
		var praiseTags []dto.BehaviorTag
		var attentionTags []dto.BehaviorTag

		for _, tag := range student.BehaviorTags {
			switch tag.Type {
			case string(consts.BehaviorTagTypeEarlyLearn), string(consts.BehaviorTagTypeQuestion), string(consts.BehaviorTagTypeCorrectStreak):
				praiseTags = append(praiseTags, tag)
			case string(consts.BehaviorTagTypePageSwitch), string(consts.BehaviorTagTypeOtherContent), string(consts.BehaviorTagTypePause):
				attentionTags = append(attentionTags, tag)
			}
		}

		// 只要有表扬标签就加入表扬列表，但只包含表扬类型的标签
		if len(praiseTags) > 0 {
			praiseCopy := student
			praiseCopy.BehaviorTags = praiseTags
			praiseList = append(praiseList, convertToAPIBehaviorCategory(praiseCopy))
		}

		// 只要有关注标签就加入关注列表，但只包含关注类型的标签
		if len(attentionTags) > 0 {
			attentionCopy := student
			attentionCopy.BehaviorTags = attentionTags
			attentionList = append(attentionList, convertToAPIBehaviorCategory(attentionCopy))
		}
	}

	// 构造响应
	response := &api.ClassroomBehaviorSummaryResponse{
		ClassroomID:   classroomID,
		StatTime:      time.Now().Unix(),
		TotalStudents: totalStudents,
		AvgAccuracy:   avgAccuracy,
		AvgProgress:   avgProgress,
		TotalDuration: totalDuration,
		PraiseList:    praiseList,
		AttentionList: attentionList,
		AllStudents:   allStudents,
	}

	return response, nil
}

// convertToAPIBehaviorCategory 将DTO转换为API响应格式
func convertToAPIBehaviorCategory(student dto.StudentBehaviorCategoryDTO) api.StudentBehaviorCategory {
	return api.StudentBehaviorCategory{
		StudentID:         student.StudentID,
		StudentName:       student.StudentName,
		BehaviorType:      string(student.BehaviorType),
		BehaviorDesc:      student.BehaviorDesc,
		ReminderCount:     student.ReminderCount,
		PraiseCount:       student.PraiseCount,
		TotalQuestions:    student.TotalQuestions,
		CorrectAnswers:    student.CorrectAnswers,
		WrongAnswers:      student.WrongAnswers,
		AccuracyRate:      student.AccuracyRate,
		LearningProgress:  student.LearningProgress,
		LastUpdateTime:    student.LastUpdateTime,
		AvatarUrl:         student.AvatarUrl,
		IsHandled:         student.IsHandled,
		HandleTime:        student.HandleTime,
		EarlyLearnCount:   student.EarlyLearnCount,
		QuestionCount:     student.QuestionCount,
		CorrectStreak:     student.CorrectStreak,
		PageSwitchCount:   student.PageSwitchCount,
		OtherContentCount: student.OtherContentCount,
		PauseCount:        student.PauseCount,
	}
}
