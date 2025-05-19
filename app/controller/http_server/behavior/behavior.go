package behavior

import (
	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/domain/behavior"
	"gil_teacher/app/middleware"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils"

	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type BehaviorController struct {
	behaviorHandler       *behavior.BehaviorHandler
	sessionMessageHandler *behavior.SessionMessageHandler
	producer              *behavior.BehaviorProducer
	teacherMiddleware     *middleware.TeacherMiddleware
	log                   *logger.ContextLogger
}

func NewBehaviorController(
	behaviorHandler *behavior.BehaviorHandler,
	sessionMessageHandler *behavior.SessionMessageHandler,
	producer *behavior.BehaviorProducer,
	teacherMiddleware *middleware.TeacherMiddleware,
	log *logger.ContextLogger,
) *BehaviorController {
	return &BehaviorController{
		behaviorHandler:       behaviorHandler,
		sessionMessageHandler: sessionMessageHandler,
		producer:              producer,
		teacherMiddleware:     teacherMiddleware,
		log:                   log,
	}
}

// RecordTeacherBehavior 记录教师行为
func (c *BehaviorController) RecordTeacherBehavior(ctx *gin.Context) {
	// 获取教师ID
	teacherID, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		c.log.Error(ctx, "获取教师ID失败: %v", err)
		response.Unauthorized(ctx)
		return
	}

	var req api.TeacherBehaviorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	req.TeacherID = uint64(teacherID)
	req.SchoolID = uint64(schoolID)
	if err := c.producer.RecordTeacherBehavior(ctx, &req); err != nil {
		c.log.Error(ctx, "记录教师行为失败: %+v", err)
		response.Err(ctx, response.ERR_KAFKA)
		return
	}

	response.Success(ctx, nil)
}

// RecordStudentBehavior 记录学生行为
func (c *BehaviorController) RecordStudentBehavior(ctx *gin.Context) {
	var req api.StudentBehaviorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	if err := c.producer.RecordStudentBehavior(ctx, &req); err != nil {
		c.log.Error(ctx, "记录学生行为失败: %v", err)
		response.Err(ctx, response.ERR_KAFKA)
		return
	}

	response.Success(ctx, nil)
}

// OpenSession 创建会话
func (c *BehaviorController) OpenSession(ctx *gin.Context) {
	var req api.OpenSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	sessionID, err := c.behaviorHandler.OpenCommunicationSession(ctx, &req)
	if err != nil {
		c.log.Error(ctx, "创建会话失败: %v", err)
		response.SystemError(ctx)
		return
	}

	response.Success(ctx, map[string]any{"session_id": sessionID})
}

// SaveMessage 记录会话内容
func (c *BehaviorController) SaveMessage(ctx *gin.Context) {
	var req api.SaveMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	messageID, err := c.producer.RecordCommunicationMessage(ctx, &req)
	if err != nil {
		c.log.Error(ctx, "记录会话内容失败: %v", err)
		response.SystemError(ctx)
		return
	}

	response.Success(ctx, map[string]any{"message_id": messageID})
}

// CloseSession 关闭会话
func (c *BehaviorController) CloseSession(ctx *gin.Context) {
	var req api.CloseSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	if err := c.behaviorHandler.CloseCommunicationSession(ctx, &req); err != nil {
		c.log.Error(ctx, "关闭会话失败: %v", err)
		response.SystemError(ctx)
		return
	}

	response.Success(ctx, nil)
}

// GetSessionMessages 查询指定会话的全部消息，分页查询
func (c *BehaviorController) GetSessionMessages(ctx *gin.Context) {
	sessionID := ctx.Query("session_id")
	page, size, err := consts.PageHandler(utils.ParseInt64(ctx.Query("page")), utils.ParseInt64(ctx.Query("size")))
	if err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}
	if sessionID == "" {
		c.log.Error(ctx, "参数解析失败, session_id 不能为空")
		response.ParamError(ctx)
		return
	}

	messages, err := c.behaviorHandler.GetCommunicationSessionMessages(ctx, sessionID, page, size)
	if err != nil {
		c.log.Error(ctx, "查询会话消息失败: %v", err)
		response.SystemError(ctx)
		return
	}

	response.Success(ctx, messages)
}

// GetClassroomMessages 查询指定课堂的全部消息
func (c *BehaviorController) GetClassroomMessages(ctx *gin.Context) {
	classroomID := ctx.Query("classroom_id")
	page, size, err := consts.PageHandler(utils.ParseInt64(ctx.Query("page")), utils.ParseInt64(ctx.Query("size")))
	if err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}
	if classroomID == "" {
		c.log.Error(ctx, "参数解析失败, classroom_id 不能为空")
		response.ParamError(ctx)
		return
	}

	userID := "" // TODO 从上下文获取
	messages, err := c.behaviorHandler.GetClassroomMessages(ctx, userID, classroomID, page, size)
	if err != nil {
		c.log.Error(ctx, "查询课堂消息失败: %v", err)
		response.SystemError(ctx)
		return
	}

	response.Success(ctx, messages)
}

// MarkMessageAsRead 标记用户消息已读
func (c *BehaviorController) MarkMessageAsRead(ctx *gin.Context) {
	sessionID := ctx.Query("session_id")
	messageID := ctx.Query("message_id")
	if sessionID == "" || messageID == "" {
		c.log.Error(ctx, "参数解析失败, session_id 和 message_id 不能为空")
		response.ParamError(ctx)
		return
	}

	userID := int64(0) // TODO 从上下文获取
	if err := c.sessionMessageHandler.MarkMessageAsRead(ctx, userID, sessionID, messageID); err != nil {
		c.log.Error(ctx, "标记消息已读失败: %v", err)
		response.SystemError(ctx)
		return
	}

	response.Success(ctx, nil)
}

// GetUnreadMessageCount 获取未读消息数量
func (c *BehaviorController) GetUnreadMessageCount(ctx *gin.Context) {
	sessionID := ctx.Query("session_id")
	if sessionID == "" {
		c.log.Error(ctx, "参数解析失败, session_id 不能为空")
		response.ParamError(ctx)
		return
	}

	userID := int64(0) // TODO 从上下文获取
	count, err := c.sessionMessageHandler.GetUnreadMessageCount(ctx, userID, sessionID)
	if err != nil {
		c.log.Error(ctx, "获取未读消息数量失败: %v", err)
		response.SystemError(ctx)
		return
	}

	response.Success(ctx, map[string]any{"count": count})
}

// GetUnreadMessageList 获取未读消息列表
func (c *BehaviorController) GetUnreadMessageList(ctx *gin.Context) {
	sessionID := ctx.Query("session_id")
	if sessionID == "" {
		c.log.Error(ctx, "参数解析失败, session_id 不能为空")
		response.ParamError(ctx)
		return
	}

	userID := int64(0) // TODO 从上下文获取
	messages, err := c.sessionMessageHandler.GetUnreadMessageList(ctx, userID, sessionID)
	if err != nil {
		c.log.Error(ctx, "获取未读消息列表失败: %v", err)
		response.SystemError(ctx)
		return
	}

	response.Success(ctx, messages)
}

// GetClassLatestBehaviors 获取班级学生最新行为
func (c *BehaviorController) GetClassLatestBehaviors(ctx *gin.Context) {
	var req api.GetClassLatestBehaviorsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.ParamError(ctx, response.ERR_INVALID_CLASSROOM)
		return
	}

	behaviors, err := c.behaviorHandler.GetClassLatestBehaviors(ctx, req.ClassroomID)
	if err != nil {
		response.Err(ctx, response.ERR_INVALID_CLASSROOM)
		return
	}

	response.Success(ctx, behaviors)
}

// GetStudentClassroomDetail 获取学生课堂详情
func (c *BehaviorController) GetStudentClassroomDetail(ctx *gin.Context) {
	// 解析请求参数
	var req api.StudentClassroomDetailRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	// 校验参数
	if err := req.Validate(); err != nil {
		c.log.Error(ctx, "参数无效: %v", err)
		response.ParamError(ctx)
		return
	}

	// 调用领域层获取数据
	detail, err := c.behaviorHandler.GetStudentClassroomDetail(ctx, &req)
	if err != nil {
		c.log.Error(ctx, "获取学生课堂详情失败: %v", err)
		response.SystemError(ctx)
		return
	}

	// 返回响应
	response.Success(ctx, detail)
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

// convertToAPIBehaviorCategoryList 批量转换学生行为分类列表
func convertToAPIBehaviorCategoryList(students []dto.StudentBehaviorCategoryDTO) []api.StudentBehaviorCategory {
	result := make([]api.StudentBehaviorCategory, len(students))
	for i, student := range students {
		result[i] = convertToAPIBehaviorCategory(student)
	}
	return result
}

// GetClassBehaviorCategory 获取课堂行为分类列表
func (c *BehaviorController) GetClassBehaviorCategory(ctx *gin.Context) {
	// 解析请求参数
	var req api.ClassBehaviorCategoryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	// 校验参数
	if err := req.Validate(); err != nil {
		c.log.Error(ctx, "参数无效: %v", err)
		response.ParamError(ctx)
		return
	}

	// 调用领域层获取数据
	categories, err := c.behaviorHandler.GetClassBehaviorCategory(ctx, req.ClassroomID)
	if err != nil {
		c.log.Error(ctx, "获取课堂行为分类列表失败: %v", err)
		response.SystemError(ctx)
		return
	}

	// 将学生行为按类型分组
	praiseList := make([]api.StudentBehaviorCategory, 0)
	attentionList := make([]api.StudentBehaviorCategory, 0)

	for _, category := range categories {
		apiCategory := convertToAPIBehaviorCategory(category)

		// 生成行为标签
		generateBehaviorTags(&apiCategory)

		// 分离表扬标签和提醒标签
		var praiseTags []api.BehaviorTag
		var attentionTags []api.BehaviorTag

		for _, tag := range apiCategory.BehaviorTags {
			switch tag.Type {
			case string(consts.BehaviorTagTypeEarlyLearn), string(consts.BehaviorTagTypeQuestion), string(consts.BehaviorTagTypeCorrectStreak):
				praiseTags = append(praiseTags, tag)
			case string(consts.BehaviorTagTypePageSwitch), string(consts.BehaviorTagTypeOtherContent), string(consts.BehaviorTagTypePause):
				attentionTags = append(attentionTags, tag)
			}
		}

		// 只要有表扬标签就加入表扬列表，但只包含表扬类型的标签
		if len(praiseTags) > 0 {
			praiseCopy := apiCategory
			praiseCopy.BehaviorTags = praiseTags
			praiseList = append(praiseList, praiseCopy)
		}

		// 只要有关注标签就加入关注列表，但只包含关注类型的标签
		if len(attentionTags) > 0 {
			attentionCopy := apiCategory
			attentionCopy.BehaviorTags = attentionTags
			attentionList = append(attentionList, attentionCopy)
		}
	}

	// 按学生ID排序表扬列表
	sort.Slice(praiseList, func(i, j int) bool {
		// 首先按IsHandled排序（false排在前面，true排在后面）
		if praiseList[i].IsHandled != praiseList[j].IsHandled {
			return !praiseList[i].IsHandled // IsHandled为false的排在前面
		}
		// 然后按学生ID排序
		return praiseList[i].StudentID < praiseList[j].StudentID
	})

	// 按学生ID排序关注列表
	sort.Slice(attentionList, func(i, j int) bool {
		// 首先按IsHandled排序（false排在前面，true排在后面）
		if attentionList[i].IsHandled != attentionList[j].IsHandled {
			return !attentionList[i].IsHandled // IsHandled为false的排在前面
		}
		// 然后按学生ID排序
		return attentionList[i].StudentID < attentionList[j].StudentID
	})

	// 构造响应数据
	responseData := &api.ClassBehaviorCategoryResponse{
		ClassroomID:   req.ClassroomID,
		QueryTime:     time.Now().Unix(),
		PraiseList:    praiseList,
		AttentionList: attentionList,
	}

	// 返回响应
	response.Success(ctx, responseData)
}

// generateBehaviorTags 生成行为标签
func generateBehaviorTags(category *api.StudentBehaviorCategory) {
	category.BehaviorTags = []api.BehaviorTag{}

	// 课堂表现很棒的标签
	if category.EarlyLearnCount > 0 {
		category.BehaviorTags = append(category.BehaviorTags, api.BehaviorTag{
			Type:  string(consts.BehaviorTagTypeEarlyLearn),
			Count: category.EarlyLearnCount,
			Text:  fmt.Sprintf(consts.BehaviorTagTextEarlyLearn, category.EarlyLearnCount),
		})
	}

	if category.QuestionCount > 0 {
		category.BehaviorTags = append(category.BehaviorTags, api.BehaviorTag{
			Type:  string(consts.BehaviorTagTypeQuestion),
			Count: category.QuestionCount,
			Text:  fmt.Sprintf(consts.BehaviorTagTextQuestion, category.QuestionCount),
		})
	}

	if category.CorrectStreak > 0 {
		category.BehaviorTags = append(category.BehaviorTags, api.BehaviorTag{
			Type:  string(consts.BehaviorTagTypeCorrectStreak),
			Count: category.CorrectStreak,
			Text:  fmt.Sprintf(consts.BehaviorTagTextCorrectStreak, category.CorrectStreak),
		})
	}

	// 需要关注的行为标签
	if category.PageSwitchCount > 0 {
		category.BehaviorTags = append(category.BehaviorTags, api.BehaviorTag{
			Type:  string(consts.BehaviorTagTypePageSwitch),
			Count: category.PageSwitchCount,
			Text:  fmt.Sprintf(consts.BehaviorTagTextPageSwitch, category.PageSwitchCount),
		})
	}

	if category.OtherContentCount > 0 {
		category.BehaviorTags = append(category.BehaviorTags, api.BehaviorTag{
			Type:  string(consts.BehaviorTagTypeOtherContent),
			Count: category.OtherContentCount,
			Text:  fmt.Sprintf(consts.BehaviorTagTextOtherContent, category.OtherContentCount),
		})
	}

	if category.PauseCount > 0 {
		category.BehaviorTags = append(category.BehaviorTags, api.BehaviorTag{
			Type:  string(consts.BehaviorTagTypePause),
			Count: category.PauseCount,
			Text:  fmt.Sprintf(consts.BehaviorTagTextPause, category.PauseCount),
		})
	}
}

// PraiseStudents 表扬学生接口
func (c *BehaviorController) PraiseStudents(ctx *gin.Context) {
	// 获取教师ID
	teacherID, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		c.log.Error(ctx, "获取教师ID失败: %v", err)
		response.Unauthorized(ctx)
		return
	}

	// 解析请求参数
	var req api.PraiseStudentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	// 校验参数
	if err := req.Validate(); err != nil {
		c.log.Error(ctx, "参数无效: %v", err)
		response.ParamError(ctx)
		return
	}

	// 调用领域层处理表扬
	req.TeacherID = teacherID
	req.SchoolID = schoolID
	result, err := c.behaviorHandler.PraiseStudents(ctx, &req)
	if err != nil {
		c.log.Error(ctx, "表扬学生失败: %v", err)
		response.SystemError(ctx)
		return
	}

	// 返回响应
	response.Success(ctx, result)
}

// AttentionStudents 关注学生接口
func (c *BehaviorController) AttentionStudents(ctx *gin.Context) {
	// 获取教师ID
	teacherID, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		c.log.Error(ctx, "获取教师ID失败: %v", err)
		response.Unauthorized(ctx)
		return
	}

	// 解析请求参数
	var req api.AttentionStudentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	req.TeacherID = teacherID
	req.SchoolID = schoolID
	// 校验参数
	if err := req.Validate(); err != nil {
		c.log.Error(ctx, "参数无效: %v", err)
		response.ParamError(ctx)
		return
	}

	// 调用领域层处理关注
	result, err := c.behaviorHandler.AttentionStudents(ctx, &req)
	if err != nil {
		c.log.Error(ctx, "关注学生失败: %v", err)
		response.SystemError(ctx)
		return
	}

	// 返回响应
	response.Success(ctx, result)
}

// GetClassroomLearningScores 获取课堂学习分列表
func (c *BehaviorController) GetClassroomLearningScores(ctx *gin.Context) {
	// 解析请求参数
	var req api.ClassroomLearningScoresRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	// 校验参数
	if req.ClassroomID == 0 {
		c.log.Error(ctx, "参数无效: 课堂ID不能为0")
		response.ParamError(ctx)
		return
	}

	// 调用领域层获取数据
	scores, err := c.behaviorHandler.GetClassroomLearningScores(ctx, req.ClassroomID)
	if err != nil {
		c.log.Error(ctx, "获取课堂学习分列表失败: %v", err)
		response.SystemError(ctx)
		return
	}

	// 返回响应
	response.Success(ctx, api.ClassroomLearningScoresResponse{
		ClassroomID: req.ClassroomID,
		QueryTime:   time.Now().Unix(),
		Students:    scores,
	})
}

// TeacherEvaluateStudent 教师评价学生
func (c *BehaviorController) TeacherEvaluateStudent(ctx *gin.Context) {
	// 获取教师ID和学校ID
	teacherID, schoolID, err := c.teacherMiddleware.GetTeacherIDInfo(ctx)
	if err != nil {
		c.log.Error(ctx, "获取教师ID失败: %v", err)
		response.Unauthorized(ctx)
		return
	}

	var req api.TeacherEvaluateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	req.SchoolID = schoolID
	req.TeacherID = teacherID
	// 验证参数
	if err := req.Validate(); err != nil {
		c.log.Error(ctx, "参数验证失败: %v", err)
		response.ParamError(ctx)
		return
	}

	// 生成推送时间，保证在整个流程中使用相同的时间戳
	pushTime := time.Now().UnixNano() / 1e6

	// 初始化响应
	evaluateResp := &api.TeacherEvaluateResponse{
		Success:            true,
		Message:            consts.BehaviorResultEvaluateSuccess,
		EvaluateID:         time.Now().UnixNano(),
		IsAlreadyEvaluated: false,
		EvaluatePrompt:     consts.EvaluatePromptDefault,
		PushTime:           pushTime,
	}

	// 检查是否已经评价过该学生
	if req.ClassroomID > 0 {
		// 使用缓存键检查是否已经评价过
		evaluateKey := consts.GetEvaluateRecordKey(req.ClassroomID, req.StudentID, req.EvaluateType)
		var evaluateTime string
		exists, _ := c.behaviorHandler.GetRedisClient().Get(ctx, evaluateKey, &evaluateTime)
		if exists {
			c.log.Warn(ctx, "该学生已经被评价过: classroomID=%d, studentID=%d, evaluateType=%s",
				req.ClassroomID, req.StudentID, req.EvaluateType)
			evaluateResp.IsAlreadyEvaluated = true
			evaluateResp.Message = consts.BehaviorResultAlreadyEvaluated
			response.Success(ctx, evaluateResp)
			return
		}

		// 记录评价时间到Redis中
		timestamp := time.Now().Format(time.RFC3339)
		if err := c.behaviorHandler.GetRedisClient().Set(ctx, evaluateKey, timestamp, consts.EvaluateRecordExpire); err != nil {
			c.log.Error(ctx, "记录评价时间失败: %v", err)
			// 继续执行，不影响主要逻辑
		}
	}

	// 构造Context
	var context string
	if req.EvaluateType == string(consts.BehaviorTypeAssignTask) {
		// 尝试解析content是否为JSON
		var contentMap map[string]interface{}
		if err := json.Unmarshal([]byte(req.Content), &contentMap); err != nil {
			// 如果不是JSON，则将其作为字符串值
			contentMap = map[string]interface{}{
				"message": req.Content,
			}
		}

		// 添加任务相关信息和推送时间
		contextMap := map[string]interface{}{
			"content":  contentMap,
			"assignId": req.AssignID,
			"taskId":   req.TaskID,
			"pushTime": pushTime, // 添加推送时间
		}
		contextBytes, err := json.Marshal(contextMap)
		if err != nil {
			c.log.Error(ctx, "构造Context失败: %v", err)
			response.SystemError(ctx)
			return
		}
		context = string(contextBytes)
	} else {
		// 尝试解析content是否为JSON
		var contentMap map[string]interface{}
		if err := json.Unmarshal([]byte(req.Content), &contentMap); err != nil {
			// 如果不是JSON，则将其作为字符串值
			contentMap = map[string]interface{}{
				"message":  req.Content,
				"pushTime": pushTime, // 添加推送时间
			}
			contentBytes, err := json.Marshal(contentMap)
			if err != nil {
				c.log.Error(ctx, "构造Content失败: %v", err)
				response.SystemError(ctx)
				return
			}
			context = string(contentBytes)
		} else {
			// 如果已经是JSON，添加推送时间
			contentMap["pushTime"] = pushTime
			contentBytes, err := json.Marshal(contentMap)
			if err != nil {
				c.log.Error(ctx, "构造Content失败: %v", err)
				response.SystemError(ctx)
				return
			}
			context = string(contentBytes)
		}
	}

	// 构造学生行为请求
	behaviorReq := &api.StudentBehaviorRequest{
		SchoolID:     uint64(req.SchoolID),
		ClassID:      uint64(req.ClassID),
		StudentID:    uint64(req.StudentID),
		ClassroomID:  uint64(req.ClassroomID),
		BehaviorType: req.EvaluateType,
		Context:      context,
	}

	// 记录学生行为
	if err := c.producer.RecordStudentBehavior(ctx, behaviorReq); err != nil {
		response.Err(ctx, response.ERR_KAFKA)
		return
	}

	// 返回成功响应
	response.Success(ctx, evaluateResp)
}

// GetClassroomBehaviorSummary 获取课后行为汇总统计
func (c *BehaviorController) GetClassroomBehaviorSummary(ctx *gin.Context) {
	// 解析请求参数
	var req api.ClassroomBehaviorSummaryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error(ctx, "参数解析失败: %v", err)
		response.ParamError(ctx)
		return
	}

	// 校验参数
	if err := req.Validate(); err != nil {
		c.log.Error(ctx, "参数无效: %v", err)
		response.ParamError(ctx)
		return
	}

	// 调用领域层获取数据
	summary, err := c.behaviorHandler.GetClassroomBehaviorSummary(ctx, req.ClassroomID)
	if err != nil {
		c.log.Error(ctx, "获取课后行为汇总统计失败: %v", err)
		response.SystemError(ctx)
		return
	}

	// 返回响应
	response.Success(ctx, summary)
}
