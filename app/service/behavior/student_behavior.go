package behavior

import (
	"context"
	"encoding/json"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/dao/behavior"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils"
	"sort"
)

type StudentBehaviorService struct {
	behaviorDao *behavior.StudentBehaviorDao
	logger      *logger.ContextLogger
}

func NewStudentBehaviorService(behaviorDao *behavior.StudentBehaviorDao, logger *logger.ContextLogger) *StudentBehaviorService {
	return &StudentBehaviorService{
		behaviorDao: behaviorDao,
		logger:      logger,
	}
}

// CalculateMaxCorrectStreak 计算最大连对数
func (s *StudentBehaviorService) CalculateMaxCorrectStreak(ctx context.Context, answers []dto.AnswerContext) int64 {
	var currentStreak int64 = 0
	var maxStreak int64 = 0

	for _, answer := range answers {
		if answer.IsCorrect == 1 {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 0
		}
	}

	return maxStreak
}

// ExtractBehaviorFields 从行为上下文中提取字段
func (s *StudentBehaviorService) ExtractBehaviorFields(ctx context.Context, behavior *dto.StudentLatestBehaviorDTO, contextStr string) {
	if contextStr == "" {
		return
	}

	var contextMap map[string]interface{}
	if err := json.Unmarshal([]byte(contextStr), &contextMap); err != nil {
		s.logger.Error(ctx, "解析行为上下文失败: %v", err)
		return
	}

	// 提取通用字段
	behavior.PageName = utils.GetMapStringKey(contextMap, "pageName")
	behavior.Subject = utils.GetMapStringKey(contextMap, "subject")
	behavior.MaterialID = utils.GetMapUint64Key(contextMap, "materialId")
	behavior.LearningType = utils.GetMapStringKey(contextMap, "learningType")
	behavior.StayDuration = utils.GetMapInt64Key(contextMap, "stayDuration")
	behavior.VideoStatus = utils.GetMapStringKey(contextMap, "videoStatus")

	//TODO  提取答题信息
	if behavior.BehaviorType == "Answer" {
		behavior.TotalQuestions = 1
		if isCorrect := utils.GetMapFloat64Key(contextMap, "isCorrect"); isCorrect == 1 {
			behavior.CorrectAnswers = 1
			behavior.AccuracyRate = 100.0
		} else {
			behavior.WrongAnswers = 1
			behavior.AccuracyRate = 0.0
		}
	}
}

// GetClassroomLearningScores 获取课堂学习分数
func (s *StudentBehaviorService) GetClassroomLearningScores(ctx context.Context, classroomID uint64) ([]*dto.StudentLearningScoreDTO, error) {
	// 获取原始行为数据
	behaviors, err := s.behaviorDao.GetClassLatestBehaviors(ctx, classroomID)
	if err != nil {
		s.logger.Error(ctx, "获取课堂学习分列表失败: %v", err)
		return nil, err
	}

	//TODO 计算学习分数
	results := make([]*dto.StudentLearningScoreDTO, 0, len(behaviors))
	for _, behavior := range behaviors {
		// 计算学习分数：学习时间占30%，答题正确率占70%
		timeScore := utils.F64Min(float64(behavior.StayDuration)/60, 30)
		accuracyScore := utils.F64Percent(float64(behavior.CorrectAnswers), float64(behavior.TotalQuestions), 2) * 70
		totalScore := int64(timeScore + accuracyScore)

		// 获取头像URL和学生姓名
		studentName := ""
		avatarURL := ""
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err == nil {
			studentName = utils.GetMapStringKey(contextMap, "student_name")
			avatarURL = utils.GetMapStringKey(contextMap, "avatar_url")
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

	// 根据学习分数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].LearningScore > results[j].LearningScore
	})

	s.logger.Info(ctx, "获取课堂(%d)学习分列表成功，共%d条记录", classroomID, len(results))
	return results, nil
}
