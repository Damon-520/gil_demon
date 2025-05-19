package dto

import (
	"context"
	"encoding/json"
	"gil_teacher/app/consts"
	"time"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeTeacherBehavior MessageType = "teacher_behavior"
	MessageTypeStudentBehavior MessageType = "student_behavior"
	MessageTypeCommunication   MessageType = "communication"
)

// BehaviorMessage 行为消息
type BehaviorMessage struct {
	Type      MessageType     `json:"type"`
	Content   json.RawMessage `json:"content"`
	Timestamp time.Time       `json:"timestamp"`
	Version   string          `json:"version"`
}

// TeacherBehaviorDTO 教师行为 DTO
type TeacherBehaviorDTO struct {
	SchoolID               uint64              `json:"schoolId"`
	ClassID                uint64              `json:"classId"`
	ClassroomID            *uint64             `json:"classroomId,omitempty"`
	TeacherID              uint64              `json:"teacherId"`
	BehaviorType           consts.BehaviorType `json:"behaviorType"`
	CommunicationSessionID *string             `json:"communicationSessionId,omitempty"`
	Context                string              `json:"context,omitempty"`
	TaskID                 uint64              `json:"taskId,omitempty"`
	AssignID               uint64              `json:"assignId,omitempty"`
	StudentID              uint64              `json:"studentId,omitempty"`
	CreateTime             time.Time           `json:"createTime"`
}

// StudentBehaviorDTO 学生行为 DTO
type StudentBehaviorDTO struct {
	SchoolID               uint64              `json:"schoolId"`
	ClassID                uint64              `json:"classId"`
	ClassroomID            *uint64             `json:"classroomId"`
	StudentID              uint64              `json:"studentId"`
	BehaviorType           consts.BehaviorType `json:"behaviorType"`
	CommunicationSessionID *string             `json:"communicationSessionId,omitempty"`
	Context                string              `json:"context,omitempty"`
	CreateTime             time.Time           `json:"createTime"`
}

// CommunicationSessionDTO 沟通会话 DTO
type CommunicationSessionDTO struct {
	SessionID   string     `json:"sessionId"`
	UserID      uint64     `json:"userId"`
	UserType    string     `json:"userType"`
	SchoolID    uint64     `json:"schoolId"`
	CourseID    uint64     `json:"courseId,omitempty"`
	ClassID     uint64     `json:"classId,omitempty"`
	ClassroomID uint64     `json:"classroomId,omitempty"`
	SessionType string     `json:"sessionType"`
	TargetID    *string    `json:"targetId,omitempty"`
	Closed      bool       `json:"closed,omitempty"`
	StartTime   time.Time  `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
}

// CommunicationMessageDTO 沟通消息 DTO
type CommunicationMessageDTO struct {
	MessageID      string    `json:"messageId"`
	SessionID      string    `json:"sessionId"`
	UserID         uint64    `json:"userId"`
	UserType       string    `json:"userType"`
	MessageContent string    `json:"messageContent"`
	MessageType    string    `json:"messageType"`
	AnswerTo       string    `json:"answerTo,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	// SchoolID       uint64    `json:"schoolId"`
	// CourseID       uint64    `json:"courseId,omitempty"`
	// ClassroomID    uint64    `json:"classroomId,omitempty"`
}

// StudentLatestBehaviorDTO 学生最新行为DTO
type StudentLatestBehaviorDTO struct {
	StudentID      int64               `json:"studentId"`      // 学生ID
	StudentName    string              `json:"studentName"`    // 学生姓名
	AvatarURL      string              `json:"avatarUrl"`      // 头像URL
	BehaviorType   consts.BehaviorType `json:"behaviorType"`   // 行为类型
	PageName       string              `json:"pageName"`       // 页面名称
	Subject        string              `json:"subject"`        // 学科
	MaterialID     uint64              `json:"materialId"`     // 素材ID
	LearningType   string              `json:"learningType"`   // 学习类型(任务/自学)
	VideoStatus    string              `json:"videoStatus"`    // 视频状态(播放/暂停)
	StayDuration   int64               `json:"stayDuration"`   // 停留时长
	Context        string              `json:"context"`        // 原始上下文
	LastUpdateTime int64               `json:"lastUpdateTime"` // 最后更新时间(UTC秒数)

	// 答题信息
	TotalQuestions int64   `json:"totalQuestions"` // 总题目数
	CorrectAnswers int64   `json:"correctAnswers"` // 正确题数
	WrongAnswers   int64   `json:"wrongAnswers"`   // 错误题数
	AccuracyRate   float64 `json:"accuracyRate"`   // 正确率(%)
}

// StudentClassroomDetailDTO 学生课堂详情DTO
type StudentClassroomDetailDTO struct {
	StudentID        uint64              `json:"studentId"`        // 学生ID
	ClassroomID      uint64              `json:"classroomId"`      // 课堂ID
	SchoolID         uint64              `json:"schoolId"`         // 学校ID
	ClassID          uint64              `json:"classId"`          // 班级ID
	TotalStudyTime   int64               `json:"totalStudyTime"`   // 总学习时长(分钟)
	ClassroomScore   int64               `json:"classroomScore"`   // 课堂学习分
	MaxCorrectStreak int64               `json:"maxCorrectStreak"` // 最大连对题目数
	QuestionCount    int64               `json:"questionCount"`    // 提问次数
	AccuracyRate     float64             `json:"accuracyRate"`     // 正确率(%)
	InteractionCount int64               `json:"interactionCount"` // 互动次数
	ViolationCount   int64               `json:"violationCount"`   // 违规行为次数
	LearningRecords  []LearningRecordDTO `json:"learningRecords"`  // 学习记录列表
}

// LearningRecordDTO 学习记录DTO
type LearningRecordDTO struct {
	RecordID     string  `json:"recordId"`     // 记录ID
	ChapterID    string  `json:"chapterId"`    // 章节ID
	ChapterName  string  `json:"chapterName"`  // 章节名称
	LearningType string  `json:"learningType"` // 学习类型（课程学习/课堂自学）
	Duration     int64   `json:"duration"`     // 学习时长(秒)
	AccuracyRate float64 `json:"accuracyRate"` // 练习正确率(%)
	Progress     float64 `json:"progress"`     // 学习进度(%)
	CreateTime   int64   `json:"createTime"`   // 记录创建时间(UTC秒数)
}

// BehaviorTag 行为标签
type BehaviorTag struct {
	Type  string `json:"type"`  // 标签类型
	Count int64  `json:"count"` // 计数
	Text  string `json:"text"`  // 显示文本
}

// StudentBehaviorCategoryDTO 学生行为分类DTO
type StudentBehaviorCategoryDTO struct {
	StudentID         uint64              `json:"studentId"`         // 学生ID
	StudentName       string              `json:"studentName"`       // 学生姓名
	BehaviorType      consts.BehaviorType `json:"behaviorType"`      // 行为类型
	BehaviorDesc      string              `json:"behaviorDesc"`      // 行为描述
	ReminderCount     int64               `json:"reminderCount"`     // 已提醒次数(关注次数)
	PraiseCount       int64               `json:"praiseCount"`       // 表扬次数
	TotalQuestions    int64               `json:"totalQuestions"`    // 总题目数
	CorrectAnswers    int64               `json:"correctAnswers"`    // 正确题数
	WrongAnswers      int64               `json:"wrongAnswers"`      // 错误题数
	AccuracyRate      float64             `json:"accuracyRate"`      // 正确率(%)
	LearningProgress  float64             `json:"learningProgress"`  // 学习进度(%)
	LastUpdateTime    int64               `json:"lastUpdateTime"`    // 最后更新时间(UTC秒数)
	AvatarUrl         string              `json:"avatarUrl"`         // 头像URL
	IsHandled         bool                `json:"isHandled"`         // 是否已处理
	HandleTime        int64               `json:"handleTime"`        // 处理时间(UTC秒数)
	EarlyLearnCount   int64               `json:"earlyLearnCount"`   // 提前学习次数
	QuestionCount     int64               `json:"questionCount"`     // 提问次数
	CorrectStreak     int64               `json:"correctStreak"`     // 连对次数
	PageSwitchCount   int64               `json:"pageSwitchCount"`   // 频繁切换页面次数
	OtherContentCount int64               `json:"otherContentCount"` // 学习其他内容次数
	PauseCount        int64               `json:"pauseCount"`        // 停顿操作次数
	BehaviorTags      []BehaviorTag       `json:"behaviorTags"`      // 行为标签列表
}

// ClassBehaviorCategoryDTO 课堂行为分类DTO
type ClassBehaviorCategoryDTO struct {
	ClassroomID   uint64                       `json:"classroomId"`   // 课堂ID
	QueryTime     int64                        `json:"queryTime"`     // 查询时间(UTC秒数)
	PraiseList    []StudentBehaviorCategoryDTO `json:"praiseList"`    // 值得表扬学生列表
	AttentionList []StudentBehaviorCategoryDTO `json:"attentionList"` // 建议关注学生列表
	HandledList   []StudentBehaviorCategoryDTO `json:"handledList"`   // 已处理学生列表
}

// BehaviorRepository 行为数据访问接口
type BehaviorRepository interface {
	SaveTeacherBehavior(ctx context.Context, behaviors []*TeacherBehaviorDTO) error
	SaveStudentBehavior(ctx context.Context, behaviors []*StudentBehaviorDTO) error
	SaveCommunication(ctx context.Context, sessions []*CommunicationSessionDTO, messages []*CommunicationMessageDTO) error
	GetClassLatestBehaviors(ctx context.Context, classID uint64) ([]*StudentLatestBehaviorDTO, error)
	GetStudentClassroomDetail(ctx context.Context, studentID, classroomID uint64) (*StudentClassroomDetailDTO, error)
	GetClassBehaviorCategory(ctx context.Context, classID, classroomID uint64) (*ClassBehaviorCategoryDTO, error)
	// 获取指定学生ID列表的行为数据
	GetStudentsBehaviors(ctx context.Context, studentIDs []uint64) ([]*StudentLatestBehaviorDTO, error)
}
