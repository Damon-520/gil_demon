package dto

// StatisticsContext 统计数据上下文
type StatisticsContext struct {
	CorrectStreak int64  `json:"correctStreak"`
	Description   string `json:"description"`
}

// LearningContext 学习行为上下文
type LearningContext struct {
	ChapterID    string  `json:"chapterId"`
	ChapterName  string  `json:"chapterName"`
	LearningType string  `json:"learningType"`
	Duration     int64   `json:"duration"`
	Progress     float64 `json:"progress"`
}

// AnswerContext 答题行为上下文
type AnswerContext struct {
	IsCorrect    int64  `json:"isCorrect"`
	ChapterID    string `json:"chapterId"`
	QuestionID   string `json:"questionId"`
	QuestionType string `json:"questionType"`
}

// StudyTimeContext 学习时长上下文
type StudyTimeContext struct {
	TotalMinutes int64 `json:"totalMinutes"`
}

// ScoreContext 得分上下文
type ScoreContext struct {
	Score int64 `json:"score"`
}

// BehaviorContext 行为上下文
type BehaviorContext struct {
	CorrectStreak   int64   `json:"correctStreak"`   // 连续答对次数
	EarlyLearnCount int64   `json:"earlyLearnCount"` // 提前学习次数
	QuestionCount   int64   `json:"questionCount"`   // 提问次数
	CorrectAnswers  int64   `json:"correctAnswers"`  // 正确答题数
	TotalQuestions  int64   `json:"totalQuestions"`  // 总题数
	LearningType    string  `json:"learningType"`    // 学习类型
	StayDuration    float64 `json:"stayDuration"`    // 停留时长(秒)
	VideoStatus     string  `json:"videoStatus"`     // 视频状态
}
