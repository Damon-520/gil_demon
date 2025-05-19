package api

import (
	"errors"
	"gil_teacher/app/consts"
	"gil_teacher/app/model/dto"
	"slices"
)

// TeacherBehaviorRequest 教师行为请求
type TeacherBehaviorRequest struct {
	SchoolID               uint64 `json:"-"`
	TeacherID              uint64 `json:"-"`
	ClassID                uint64 `json:"classId" binding:"required"`
	ClassroomID            uint64 `json:"classroomId"`
	BehaviorType           string `json:"behaviorType" binding:"required"`
	CommunicationSessionID string `json:"communicationSessionId"`
	Context                string `json:"context"`
}

func (r *TeacherBehaviorRequest) Validate() error {
	if r.BehaviorType == "" {
		return errors.New("behavior_type is required")
	}
	if r.ClassID == 0 {
		return errors.New("class_id is required")
	}
	return nil
}

// StudentBehaviorRequest 学生行为请求
type StudentBehaviorRequest struct {
	// schooldID, ClassID, StudentID 都应该是从token中获取 TODO
	SchoolID               uint64 `json:"-"`
	StudentID              uint64 `json:"-"`
	ClassID                uint64 `json:"classId" binding:"required"`
	ClassroomID            uint64 `json:"classroomId"`
	BehaviorType           string `json:"behaviorType" binding:"required"`
	CommunicationSessionID string `json:"communicationSessionId"`
	Context                string `json:"context"`
}

func (r *StudentBehaviorRequest) Validate() error {
	if r.BehaviorType == "" {
		return errors.New("behavior_type is required")
	}
	if !slices.Contains([]string{string(consts.BehaviorTypeCommunication)}, r.BehaviorType) {
		return errors.New("behavior_type is invalid")
	}
	if r.StudentID == 0 {
		return errors.New("student_id is required")
	}
	return nil
}

// OpenSessionRequest 创建会话请求
type OpenSessionRequest struct {
	// schooldID 和 userID 都应该是从token中获取
	SchoolID    uint64 `json:"schoolId" binding:"required"`
	UserID      uint64 `json:"userId" binding:"required"`
	UserType    string `json:"userType" binding:"required"` // 用户类型，可以是老师、学生、ai等
	CourseID    uint64 `json:"courseId"`
	ClassID     uint64 `json:"classId"`
	ClassroomID uint64 `json:"classroomId"`
	SessionType string `json:"sessionType" binding:"required"` // 会话类型，可以是提问、答疑、讨论、辅导等
	TargetID    string `json:"targetId" binding:"required"`    // 目标id，可以是教学视频、音频、图片、文档等，也可以是学生id
	TargetType  string `json:"targetType" binding:"required"`  // 目标类型，可以是教学视频、音频、图片、文档等，也可以是学生id
}

func (r *OpenSessionRequest) Validate() error {
	if r.SchoolID == 0 {
		return errors.New("schoolId is required")
	}
	if r.UserType == "" {
		return errors.New("userType is required")
	}
	if !slices.Contains(
		[]string{
			string(consts.CommunicationUserTypeAI),
			string(consts.CommunicationUserTypeStudent),
			string(consts.CommunicationUserTypeTeacher),
		},
		r.UserType,
	) {
		return errors.New("userType is invalid")
	}
	if r.UserID == 0 {
		return errors.New("userId is required")
	}
	if r.SessionType == "" {
		return errors.New("sessionType is required")
	}
	if !slices.Contains(
		[]string{
			string(consts.CommunicationSessionTypeQuestion),
			string(consts.CommunicationSessionTypeAnswer),
			string(consts.CommunicationSessionTypeChat),
		},
		r.SessionType,
	) {
		return errors.New("sessionType is invalid")
	}
	if r.TargetID == "" {
		return errors.New("targetId is required")
	}
	return nil
}

// SaveMessageRequest 记录会话内容请求
type SaveMessageRequest struct {
	SchoolID       uint64 `json:"schoolId" binding:"required"`
	SessionID      string `json:"sessionId" binding:"required"`
	UserID         uint64 `json:"userId" binding:"required"`
	UserType       string `json:"userType" binding:"required"`
	MessageContent string `json:"messageContent"`
	MessageType    string `json:"messageType" binding:"required"`
	AnswerTo       string `json:"answerTo"`
}

func (r *SaveMessageRequest) Validate() error {
	if r.SchoolID == 0 {
		return errors.New("schoolId is required")
	}
	if r.SessionID == "" {
		return errors.New("sessionId is required")
	}
	if r.UserID == 0 {
		return errors.New("userId is required")
	}
	if r.UserType == "" {
		return errors.New("userType is required")
	}
	if !slices.Contains(
		[]string{
			string(consts.CommunicationUserTypeAI),
			string(consts.CommunicationUserTypeStudent),
			string(consts.CommunicationUserTypeTeacher),
		},
		r.UserType,
	) {
		return errors.New("userType is invalid")
	}
	if r.MessageType == "" {
		return errors.New("messageType is required")
	}
	return nil
}

// CloseSessionRequest 关闭会话请求，只有老师和发起人可以关闭会话
type CloseSessionRequest struct {
	SchoolID  uint64 `json:"schoolId" binding:"required"`
	SessionID string `json:"sessionId" binding:"required"`
	UserID    uint64 `json:"userId" binding:"required"`
	UserType  string `json:"userType" binding:"required"` // 用户类型，只能是老师或学生
}

func (r *CloseSessionRequest) Validate() error {
	if r.SchoolID == 0 {
		return errors.New("schoolId is required")
	}
	if r.SessionID == "" {
		return errors.New("sessionId is required")
	}
	if r.UserID == 0 {
		return errors.New("userId is required")
	}
	if r.UserType == "" {
		return errors.New("userType is required")
	}
	if !slices.Contains(
		[]string{
			string(consts.CommunicationUserTypeTeacher),
			string(consts.CommunicationUserTypeStudent),
		},
		r.UserType,
	) {
		return errors.New("userType is invalid")
	}
	return nil
}

// GetClassLatestBehaviorsRequest 获取班级学生最新行为请求
type GetClassLatestBehaviorsRequest struct {
	ClassroomID uint64 `form:"classroomId" binding:"required"` // 课堂ID
}

// StudentLatestBehaviorResponse 学生最新行为响应
type StudentLatestBehaviorResponse struct {
	StudentID      int64  `json:"studentId"`      // 学生ID
	BehaviorType   string `json:"behaviorType"`   // 行为类型
	PageName       string `json:"pageName"`       // 页面名称
	Subject        string `json:"subject"`        // 学科
	MaterialID     uint64 `json:"materialId"`     // 素材ID
	LearningType   string `json:"learningType"`   // 学习类型(任务/自学)
	VideoStatus    string `json:"videoStatus"`    // 视频状态(播放/暂停)
	StayDuration   int64  `json:"stayDuration"`   // 停留时长
	Context        string `json:"context"`        // 原始上下文
	LastUpdateTime int64  `json:"lastUpdateTime"` // 最后更新时间(UTC秒数)

	// 答题信息
	TotalQuestions int64   `json:"totalQuestions"` // 总题目数
	CorrectAnswers int64   `json:"correctAnswers"` // 正确题数
	WrongAnswers   int64   `json:"wrongAnswers"`   // 错误题数
	AccuracyRate   float64 `json:"accuracyRate"`   // 正确率(%)
}

// ClassLatestBehaviorsResponse 班级学生最新行为响应
type ClassLatestBehaviorsResponse struct {
	ClassID     uint64                          `json:"classId"`     // 班级ID
	QueryTime   int64                           `json:"queryTime"`   // 查询时间(UTC秒数)
	StudentData []StudentLatestBehaviorResponse `json:"studentData"` // 学生数据列表
}

// StudentClassroomDetailRequest 学生课堂详情请求
type StudentClassroomDetailRequest struct {
	StudentID   uint64 `json:"studentId" form:"studentId" binding:"required"`     // 学生ID
	ClassroomID uint64 `json:"classroomId" form:"classroomId" binding:"required"` // 课堂ID
}

// Validate 验证请求参数
func (r *StudentClassroomDetailRequest) Validate() error {
	if r.StudentID == 0 {
		return errors.New("学生ID不能为0")
	}
	if r.ClassroomID == 0 {
		return errors.New("课堂ID不能为0")
	}
	return nil
}

// StudentClassroomDetailResponse 学生课堂详情响应
type StudentClassroomDetailResponse struct {
	StudentID        uint64                   `json:"studentId"`          // 学生ID
	ClassroomID      uint64                   `json:"classroomId"`        // 课堂ID
	SchoolID         uint64                   `json:"schoolId"`           // 学校ID
	ClassID          uint64                   `json:"classId"`            // 班级ID
	TotalStudyTime   int64                    `json:"totalStudyTime"`     // 总学习时长(分钟)
	ClassroomScore   int64                    `json:"classroomScore"`     // 课堂学习分
	MaxCorrectStreak int64                    `json:"maxCorrectStreak"`   // 最大连对题目数
	QuestionCount    int64                    `json:"questionCount"`      // 提问次数
	AccuracyRate     float64                  `json:"accuracyRate"`       // 正确率(%)
	InteractionCount int64                    `json:"interactionCount"`   // 互动次数
	ViolationCount   int64                    `json:"violationCount"`     // 违规次数
	LearningRecords  []LearningRecordResponse `json:"learningRecords"`    // 学习记录列表
	IsEvaluated      bool                     `json:"isEvaluated"`        // 是否已评价
	EvaluateContent  string                   `json:"evaluateContent"`    // 评价内容
	IsHandled        bool                     `json:"isHandled"`          // 是否已处理
	PushTime         int64                    `json:"pushTime,omitempty"` // 推送时间(UTC秒数)
}

// LearningRecordResponse 学习记录响应
type LearningRecordResponse struct {
	RecordID     string  `json:"recordId"`     // 记录ID
	ChapterID    string  `json:"chapterId"`    // 章节ID
	ChapterName  string  `json:"chapterName"`  // 章节名称
	LearningType string  `json:"learningType"` // 学习类型（课程学习/课堂自学）
	Duration     int64   `json:"duration"`     // 学习时长(秒)
	AccuracyRate float64 `json:"accuracyRate"` // 练习正确率(%)
	Progress     float64 `json:"progress"`     // 学习进度(%)
	CreateTime   int64   `json:"createTime"`   // 记录创建时间(UTC秒数)
}

// ClassBehaviorCategoryRequest 获取课堂行为分类列表请求
type ClassBehaviorCategoryRequest struct {
	ClassroomID uint64 `form:"classroomId" binding:"required"` // 课堂ID
}

// Validate 验证请求参数
func (r *ClassBehaviorCategoryRequest) Validate() error {
	if r.ClassroomID == 0 {
		return errors.New("课堂ID不能为0")
	}
	return nil
}

// GetClassBehaviorCategoryRequest 获取课堂行为分类列表请求
type GetClassBehaviorCategoryRequest struct {
	ClassroomID uint64 `form:"classroomId" binding:"required"` // 课堂ID
}

// GetClassBehaviorCategoryResponse 获取课堂行为分类列表响应
type GetClassBehaviorCategoryResponse struct {
	Categories []StudentBehaviorCategory `json:"categories"` // 行为分类列表
}

// BehaviorTag 行为标签
type BehaviorTag struct {
	Type  string `json:"type"`  // 标签类型
	Count int64  `json:"count"` // 计数
	Text  string `json:"text"`  // 显示文本
}

// StudentBehaviorCategory 学生行为分类
type StudentBehaviorCategory struct {
	StudentID         uint64        `json:"studentId"`         // 学生ID
	StudentName       string        `json:"studentName"`       // 学生姓名
	BehaviorType      string        `json:"behaviorType"`      // 行为类型
	BehaviorDesc      string        `json:"behaviorDesc"`      // 行为描述
	ReminderCount     int64         `json:"reminderCount"`     // 已提醒次数(关注次数)
	PraiseCount       int64         `json:"praiseCount"`       // 表扬次数
	TotalQuestions    int64         `json:"totalQuestions"`    // 总题目数
	CorrectAnswers    int64         `json:"correctAnswers"`    // 正确题数
	WrongAnswers      int64         `json:"wrongAnswers"`      // 错误题数
	AccuracyRate      float64       `json:"accuracyRate"`      // 正确率(%)
	LearningProgress  float64       `json:"learningProgress"`  // 学习进度(%)
	LastUpdateTime    int64         `json:"lastUpdateTime"`    // 最后更新时间(UTC秒数)
	AvatarUrl         string        `json:"avatarUrl"`         // 头像URL
	IsHandled         bool          `json:"isHandled"`         // 是否已处理
	HandleTime        int64         `json:"handleTime"`        // 处理时间(UTC秒数)
	EarlyLearnCount   int64         `json:"earlyLearnCount"`   // 提前学习次数
	QuestionCount     int64         `json:"questionCount"`     // 提问次数
	CorrectStreak     int64         `json:"correctStreak"`     // 连对次数
	PageSwitchCount   int64         `json:"pageSwitchCount"`   // 频繁切换页面次数
	OtherContentCount int64         `json:"otherContentCount"` // 学习其他内容次数
	PauseCount        int64         `json:"pauseCount"`        // 停顿操作次数
	BehaviorTags      []BehaviorTag `json:"behaviorTags"`      // 行为标签列表
}

// ClassBehaviorCategoryResponse 获取课堂行为分类列表响应
type ClassBehaviorCategoryResponse struct {
	ClassroomID   uint64                    `json:"classroomId"`   // 课堂ID
	QueryTime     int64                     `json:"queryTime"`     // 查询时间(UTC秒数)
	PraiseList    []StudentBehaviorCategory `json:"praiseList"`    // 值得表扬学生列表
	AttentionList []StudentBehaviorCategory `json:"attentionList"` // 建议关注学生列表
	//HandledList   []StudentBehaviorCategory `json:"handledList"`   // 已处理学生列表
}

// PraiseStudentRequest 表扬学生请求
type PraiseStudentRequest struct {
	ClassroomID  uint64   `json:"classroomId" binding:"required"` // 课堂ID
	StudentIDs   []uint64 `json:"studentIds" binding:"required"`  // 学生ID列表
	BehaviorType string   `json:"behaviorType"`                   // 行为类型（系统自动决定：连续答对、提前学习、主动提问）
	TeacherID    int64    `json:"-"`                              // 教师ID
	SchoolID     int64    `json:"-"`                              // 学校ID
	ClassID      int64    `json:"-"`                              // 班级ID
}

// Validate 验证请求参数
func (r *PraiseStudentRequest) Validate() error {
	if r.ClassroomID == 0 {
		return errors.New("课堂ID不能为0")
	}
	if len(r.StudentIDs) == 0 {
		return errors.New("学生ID列表不能为空")
	}

	// 行为类型现在由后端自动决定，前端不再传入
	return nil
}

// AttentionStudentRequest 关注/提醒学生请求
type AttentionStudentRequest struct {
	ClassroomID uint64   `json:"classroomId" binding:"required"` // 课堂ID
	StudentIDs  []uint64 `json:"studentIds" binding:"required"`  // 学生ID列表
	TeacherID   int64    `json:"-"`                              // 教师ID
	SchoolID    int64    `json:"-"`                              // 学校ID
}

// Validate 验证请求参数
func (r *AttentionStudentRequest) Validate() error {
	if r.ClassroomID == 0 {
		return errors.New("课堂ID不能为0")
	}
	if len(r.StudentIDs) == 0 {
		return errors.New("学生ID列表不能为空")
	}
	return nil
}

// StudentHandleResult 学生处理结果
type StudentHandleResult struct {
	StudentID            uint64   `json:"studentId"`            // 学生ID
	StudentName          string   `json:"studentName"`          // 学生姓名
	Success              bool     `json:"success"`              // 处理是否成功
	Message              string   `json:"message"`              // 处理结果消息
	AvailablePraiseTypes []string `json:"availablePraiseTypes"` // 可用的表扬类型列表
}

// PraiseStudentResponse 表扬学生响应
type PraiseStudentResponse struct {
	ClassroomID uint64                `json:"classroomId"` // 课堂ID
	Success     bool                  `json:"success"`     // 是否成功
	Message     string                `json:"message"`     // 结果消息
	HandleTime  int64                 `json:"handleTime"`  // 处理时间(UTC秒数)
	Results     []StudentHandleResult `json:"results"`     // 各学生处理结果
}

// AttentionStudentResponse 关注学生响应结果
type AttentionStudentResponse struct {
	Success     bool                  `json:"success"`     // 操作是否成功
	Message     string                `json:"message"`     // 操作结果消息
	ClassroomID uint64                `json:"classroomId"` // 课堂ID
	HandleTime  int64                 `json:"handleTime"`  // 处理时间
	Results     []StudentHandleResult `json:"results"`     // 学生处理结果
}

// ClassroomLearningScoresRequest 获取课堂学习分列表请求
type ClassroomLearningScoresRequest struct {
	ClassroomID uint64 `form:"classroomId" binding:"required"` // 课堂ID
}

// ClassroomLearningScoresResponse 获取课堂学习分列表响应
type ClassroomLearningScoresResponse struct {
	ClassroomID uint64                         `json:"classroomId"` // 课堂ID
	QueryTime   int64                          `json:"queryTime"`   // 查询时间(UTC秒数)
	Students    []*dto.StudentLearningScoreDTO `json:"students"`    // 学生学习分列表
}

// TeacherEvaluateRequest 教师评价学生请求
type TeacherEvaluateRequest struct {
	StudentID    int64  `json:"studentId" binding:"required"`    // 学生ID
	Content      string `json:"content" binding:"required"`      // 评价内容
	EvaluateType string `json:"evaluateType" binding:"required"` // 评价类型
	AssignID     int64  `json:"assignID"`                        // 布置ID（可选）
	TaskID       int64  `json:"taskID"`                          // 任务ID（可选）
	ClassroomID  int64  `json:"classroomId"`                     // 课堂ID（可选）
	ClassID      int64  `json:"classId"`                         // 班级ID
	SchoolID     int64  `json:"-"`                               // 学校ID
	TeacherID    int64  `json:"-"`                               // 教师ID
}

// Validate 验证必填字段
func (r *TeacherEvaluateRequest) Validate() error {
	if r.StudentID == 0 {
		return errors.New("student_id is required")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	if r.EvaluateType == "" {
		return errors.New("evaluate_type is required")
	}

	// 验证评价类型是否合法
	if !slices.Contains(
		[]string{
			string(consts.BehaviorTypeAssignTask),
			string(consts.BehaviorTypeClassComment),
		},
		r.EvaluateType,
	) {
		return errors.New("evaluate_type is invalid")
	}

	// 如果是布置任务类型，验证必要的任务信息
	if r.EvaluateType == string(consts.BehaviorTypeAssignTask) {
		if r.AssignID == 0 {
			return errors.New("assign_id is required for task assignment")
		}
		if r.TaskID == 0 {
			return errors.New("task_id is required for task assignment")
		}
	}

	return nil
}

// TeacherEvaluateResponse 教师评价学生响应
type TeacherEvaluateResponse struct {
	Success            bool   `json:"success"`            // 是否成功
	Message            string `json:"message"`            // 消息
	EvaluateID         int64  `json:"evaluateId"`         // 评价ID
	IsAlreadyEvaluated bool   `json:"isAlreadyEvaluated"` // 是否已评价过
	EvaluatePrompt     string `json:"evaluatePrompt"`     // 评价提示语
	PushTime           int64  `json:"pushTime"`           // 推送时间(UTC秒数)
}

// ClassroomBehaviorSummaryRequest 课后行为汇总统计请求
type ClassroomBehaviorSummaryRequest struct {
	ClassroomID uint64 `form:"classroomId" binding:"required"` // 课堂ID
}

// Validate 验证请求参数
func (r *ClassroomBehaviorSummaryRequest) Validate() error {
	if r.ClassroomID == 0 {
		return errors.New("课堂ID不能为0")
	}
	return nil
}

// ClassroomBehaviorSummaryResponse 课后行为汇总统计响应
type ClassroomBehaviorSummaryResponse struct {
	ClassroomID   uint64                    `json:"classroomId"`   // 课堂ID
	StatTime      int64                     `json:"statTime"`      // 统计时间(UTC秒数)
	TotalStudents int64                     `json:"totalStudents"` // 课堂总人数
	AvgAccuracy   float64                   `json:"avgAccuracy"`   // 班级平均正确率(%)
	AvgProgress   float64                   `json:"avgProgress"`   // 班级平均进度(%)
	TotalDuration int64                     `json:"totalDuration"` // 班级总学习时长(秒)
	PraiseList    []StudentBehaviorCategory `json:"praiseList"`    // 值得表扬学生列表
	AttentionList []StudentBehaviorCategory `json:"attentionList"` // 建议关注学生列表
	AllStudents   []StudentBehaviorCategory `json:"allStudents"`   // 所有学生列表
}
