package dto

// StudentLearningScoreDTO 学生学习分数DTO
type StudentLearningScoreDTO struct {
	StudentID     uint64 `json:"studentId"`     // 学生ID
	StudentName   string `json:"studentName"`   // 学生姓名
	AvatarURL     string `json:"avatarUrl"`     // 头像URL
	LearningScore int64  `json:"learningScore"` // 学习分数
	LearningTime  uint64 `json:"learningTime"`  // 学习时长(秒)
	CorrectCount  uint64 `json:"correctCount"`  // 正确答题数
	TotalCount    uint64 `json:"totalCount"`    // 总答题数
}
