package behavior

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gil_teacher/app/consts"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils"
	"gil_teacher/app/utils/idtools"
)

type StudentBehaviorDao struct {
	db     *dao.ClickHouseRWClient
	logger *clogger.ContextLogger
}

func newStudentBehaviorDao(db *dao.ClickHouseRWClient, logger *clogger.ContextLogger) *StudentBehaviorDao {
	return &StudentBehaviorDao{
		db:     db,
		logger: logger,
	}
}

// StudentBehavior 学生行为表结构
type StudentBehavior struct {
	ID                     string    `ch:"id"`
	SchoolID               uint64    `ch:"school_id"`
	ClassID                uint64    `ch:"class_id"`
	ClassroomID            *uint64   `ch:"classroom_id"`
	StudentID              uint64    `ch:"student_id"`
	BehaviorType           string    `ch:"behavior_type"`
	CommunicationSessionID *string   `ch:"communication_session_id"`
	LastMessageID          *string   `ch:"last_message_id"`
	Context                string    `ch:"context"`
	CreateTime             time.Time `ch:"create_time"`
	UpdateTime             time.Time `ch:"update_time"`
}

func (m *StudentBehavior) TableName() string {
	return "tbl_student_behavior_logs"
}

// 给模型数据生成主键 id，方便插入
func (m *StudentBehavior) GenerateID(ctx context.Context) string {
	if m.ID == "" {
		uuid := idtools.GetUUID()
		m.ID = uuid
	}
	return m.ID
}

func (m *StudentBehaviorDao) DB(ctx context.Context) *dao.ClickHouseRWClient {
	return m.db.Model(&StudentBehavior{})
}

// 获取用户在某堂课的全部行为
func (m *StudentBehaviorDao) GetStudentCourseBehaviors(ctx context.Context, userID uint64, courseID, classroomID uint64, pageInfo *consts.DBPageInfo) ([]*dto.StudentBehaviorDTO, error) {
	records := make([]*StudentBehavior, 0)
	err := m.DB(ctx).FindAll(ctx, &records, map[string]any{"student_id": userID, "course_id": courseID, "classroom_id": classroomID}, pageInfo)
	if err != nil {
		return nil, err
	}

	var behaviors []*dto.StudentBehaviorDTO
	for _, behavior := range records {
		behaviors = append(behaviors, &dto.StudentBehaviorDTO{
			SchoolID:               behavior.SchoolID,
			ClassID:                behavior.ClassID,
			ClassroomID:            behavior.ClassroomID,
			StudentID:              behavior.StudentID,
			BehaviorType:           consts.BehaviorType(behavior.BehaviorType),
			CommunicationSessionID: behavior.CommunicationSessionID,
			Context:                behavior.Context,
			CreateTime:             behavior.CreateTime,
		})
	}
	return behaviors, nil
}

// 更新学生最新已读消息 id
func (m *StudentBehaviorDao) UpdateStudentLastMessageID(ctx context.Context, studentID uint64, sessionID, messageID string) error {
	err := m.DB(ctx).Update(ctx, map[string]any{"last_message_id": messageID}, map[string]any{"student_id": studentID, "communication_session_id": sessionID})
	if err != nil {
		return err
	}
	return nil
}

// 批量保存学生行为
func (m *StudentBehaviorDao) SaveStudentBehaviors(ctx context.Context, behaviors []*StudentBehavior) error {
	_, err := m.DB(ctx).BatchInsert(ctx, behaviors)
	if err != nil {
		return err
	}
	return nil
}

// GetClassLatestBehaviors 获取班级学生最新行为
func (m *StudentBehaviorDao) GetClassLatestBehaviors(ctx context.Context, classRoomID uint64) ([]*dto.StudentLatestBehaviorDTO, error) {
	// 定义查询结果接收结构
	var records []*struct {
		StudentID      uint64    `ch:"student_id"`
		BehaviorType   string    `ch:"behavior_type"`
		Context        string    `ch:"context"`
		CreateTime     time.Time `ch:"create_time"`
		StayDuration   uint64    `ch:"stay_duration"`
		TotalQuestions uint64    `ch:"total_questions"`
		CorrectAnswers uint64    `ch:"correct_answers"`
	}

	// 使用 WITH 子句优化查询
	query := `
		WITH 
		latest_behaviors AS (
			SELECT 
				student_id,
				behavior_type,
				context,
				create_time,
				JSONExtractUInt(context, 'stayDuration', 0) as stay_duration,
				if(behavior_type = 'Answer', 
					countIf(JSONExtractInt(context, 'isCorrect') = 1) OVER (PARTITION BY student_id), 
					0) as correct_answers,
				if(behavior_type = 'Answer',
					count(*) OVER (PARTITION BY student_id),
					0) as total_questions
			FROM tbl_student_behavior_logs
			WHERE classroom_id = ? AND behavior_type != 'class_comment'
			ORDER BY student_id, create_time DESC
		)
		SELECT 
			student_id,
			any(behavior_type) as behavior_type,
			any(context) as context,
			max(create_time) as create_time,
			sum(stay_duration) as stay_duration,
			max(total_questions) as total_questions,
			max(correct_answers) as correct_answers
		FROM latest_behaviors
		GROUP BY student_id
		ORDER BY stay_duration DESC
		LIMIT 100
	`

	// 执行查询
	if err := m.db.Read(ctx, &records, query, classRoomID); err != nil {
		m.logger.Error(ctx, "查询班级学生最新行为失败: %v", err)
		return nil, err
	}

	// 转换为 StudentLatestBehaviorDTO 结构
	results := make([]*dto.StudentLatestBehaviorDTO, len(records))
	for i, record := range records {
		// 创建基本行为记录
		behavior := &dto.StudentLatestBehaviorDTO{
			StudentID:      int64(record.StudentID),
			BehaviorType:   consts.BehaviorType(record.BehaviorType),
			Context:        record.Context,
			LastUpdateTime: record.CreateTime.Unix(),
			StayDuration:   int64(record.StayDuration),
			TotalQuestions: int64(record.TotalQuestions),
			CorrectAnswers: int64(record.CorrectAnswers),
			AccuracyRate:   utils.F64Percent(float64(record.CorrectAnswers), float64(record.TotalQuestions), 2),
		}

		// 从Context中提取更多信息
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(record.Context), &contextMap); err == nil {
			// 提取学生基本信息
			behavior.StudentName = utils.GetMapStringKey(contextMap, "student_name")
			behavior.AvatarURL = utils.GetMapStringKey(contextMap, "avatar_url")
			// 提取其他信息
			behavior.PageName = utils.GetMapStringKey(contextMap, "page_name")
			behavior.Subject = utils.GetMapStringKey(contextMap, "subject")
			behavior.MaterialID = utils.GetMapUint64Key(contextMap, "material_id")
			behavior.LearningType = utils.GetMapStringKey(contextMap, "learning_type")
			behavior.VideoStatus = utils.GetMapStringKey(contextMap, "video_status")
			behavior.WrongAnswers = utils.GetMapInt64Key(contextMap, "wrong_answers")
		}

		results[i] = behavior
	}

	m.logger.Debug(ctx, "获取班级(%d)学生最新行为成功，共%d条记录", classRoomID, len(results))
	return results, nil
}

// GetAnswerRecords 获取答题记录
func (m *StudentBehaviorDao) GetAnswerRecords(ctx context.Context, studentID, classroomID uint64) ([]dto.AnswerContext, error) {
	var answers []dto.AnswerContext
	query := `
		SELECT 
			JSONExtractInt(context, 'isCorrect') AS is_correct,
			JSONExtractString(context, 'chapterId') AS chapter_id,
			JSONExtractString(context, 'questionId') AS question_id,
			JSONExtractString(context, 'questionType') AS question_type
		FROM tbl_student_behavior_logs
		WHERE student_id = ? AND classroom_id = ? AND behavior_type = 'Answer'
		ORDER BY create_time ASC
	`

	err := m.db.Read(ctx, &answers, query, studentID, classroomID)
	if err != nil {
		return nil, err
	}

	return answers, nil
}

// GetStudentClassroomDetail 获取学生课堂详情
func (m *StudentBehaviorDao) GetStudentClassroomDetail(ctx context.Context, studentID, classroomID uint64) (*dto.StudentClassroomDetailDTO, error) {
	m.logger.Debug(ctx, "获取学生课堂详情: studentID=%d, classroomID=%d", studentID, classroomID)

	// 参数验证
	if studentID == 0 || classroomID == 0 {
		return nil, errors.New("学生ID和课堂ID不能为0")
	}

	// 初始化结果
	result := &dto.StudentClassroomDetailDTO{
		StudentID:       studentID,
		ClassroomID:     classroomID,
		LearningRecords: make([]dto.LearningRecordDTO, 0),
	}

	// 1. 查询学校ID和班级ID
	if err := m.querySchoolAndClassID(ctx, result, studentID, classroomID); err != nil {
		m.logger.Warn(ctx, "查询学校和班级ID失败: %v", err)
		// 继续执行，不返回错误
	}

	// 2. 查询行为统计数据
	if err := m.queryBehaviorStatistics(ctx, result, studentID, classroomID); err != nil {
		m.logger.Error(ctx, "查询行为统计失败: %v", err)
		return nil, fmt.Errorf("查询行为统计失败: %w", err)
	}

	// 3. 查询学习时长
	if err := m.queryStudyTime(ctx, result, studentID, classroomID); err != nil {
		m.logger.Error(ctx, "查询学习时长失败: %v", err)
	}

	// 4. 查询课堂得分
	if err := m.queryClassroomScore(ctx, result, studentID, classroomID); err != nil {
		m.logger.Error(ctx, "查询课堂得分失败: %v", err)
	}

	// 5. 查询最大连对数
	if err := m.queryAndUpdateMaxCorrectStreak(ctx, result, studentID, classroomID); err != nil {
		m.logger.Error(ctx, "查询最大连对数失败: %v", err)
	}

	// 6. 获取学习记录
	learningRecords, err := m.getStudentLearningRecords(ctx, studentID, classroomID)
	if err != nil {
		m.logger.Error(ctx, "获取学生学习记录失败: %v", err)
	} else {
		result.LearningRecords = learningRecords
	}

	m.logger.Debug(ctx, "成功获取学生课堂详情: %+v", result)

	// 数据完整性检查
	if result.SchoolID == 0 || result.ClassID == 0 {
		m.logger.Warn(ctx, "未能获取有效的SchoolID或ClassID: schoolId=%d, classId=%d", result.SchoolID, result.ClassID)
	}

	return result, nil
}

// querySchoolAndClassID 查询学校ID和班级ID
func (m *StudentBehaviorDao) querySchoolAndClassID(ctx context.Context, result *dto.StudentClassroomDetailDTO, studentID, classroomID uint64) error {
	// 定义完整的结构体，包含所有表字段
	var idInfo struct {
		ID                     string    `ch:"id"`
		SchoolID               uint64    `ch:"school_id"`
		ClassID                uint64    `ch:"class_id"`
		StudentID              uint64    `ch:"student_id"`
		ClassroomID            *uint64   `ch:"classroom_id"`
		BehaviorType           string    `ch:"behavior_type"`
		CommunicationSessionID *string   `ch:"communication_session_id"`
		LastMessageID          *string   `ch:"last_message_id"`
		Context                string    `ch:"context"`
		CreateTime             time.Time `ch:"create_time"`
		UpdateTime             time.Time `ch:"update_time"`
	}

	where := map[string]any{
		"student_id":   studentID,
		"classroom_id": classroomID,
	}

	if err := m.DB(ctx).Find(ctx, &idInfo, where); err != nil {
		return err
	}

	result.SchoolID = idInfo.SchoolID
	result.ClassID = idInfo.ClassID
	return nil
}

// queryBehaviorStatistics 查询行为统计数据
func (m *StudentBehaviorDao) queryBehaviorStatistics(ctx context.Context, result *dto.StudentClassroomDetailDTO, studentID, classroomID uint64) error {
	var behaviorCounts struct {
		QuestionCount    *uint64 `ch:"question_count"`
		InteractionCount *uint64 `ch:"interaction_count"`
		AccuracyRate     float64 `ch:"accuracy_rate"`
	}
	behaviorQuery := `
		SELECT
			countIf(behavior_type = 'Question') AS question_count,
			countIf(behavior_type IN ('Answer', 'Question', 'Interact')) AS interaction_count,
			if(countIf(behavior_type = 'Answer') > 0, 
				(countIf(behavior_type = 'Answer' AND JSONExtractInt(context, 'isCorrect') = 1) * 100.0) / countIf(behavior_type = 'Answer'), 
				0) AS accuracy_rate
		FROM tbl_student_behavior_logs
		WHERE student_id = ? AND classroom_id = ?
	`
	if err := m.db.Read(ctx, &behaviorCounts, behaviorQuery, studentID, classroomID); err != nil {
		return err
	}

	if behaviorCounts.QuestionCount != nil {
		result.QuestionCount = int64(*behaviorCounts.QuestionCount)
	}
	if behaviorCounts.InteractionCount != nil {
		result.InteractionCount = int64(*behaviorCounts.InteractionCount)
	}
	result.AccuracyRate = behaviorCounts.AccuracyRate

	m.logger.Debug(ctx, "获取行为统计成功: 提问次数=%d, 互动次数=%d, 正确率=%.2f%%",
		result.QuestionCount, result.InteractionCount, behaviorCounts.AccuracyRate)
	return nil
}

// queryStudyTime 查询学习时长
func (m *StudentBehaviorDao) queryStudyTime(ctx context.Context, result *dto.StudentClassroomDetailDTO, studentID, classroomID uint64) error {
	var studyTimeData struct {
		TotalMinutes *int64 `ch:"total_minutes"`
	}
	studyTimeQuery := `
		SELECT 
			JSONExtractInt(context, 'totalMinutes') AS total_minutes
		FROM tbl_student_behavior_logs
		WHERE student_id = ? AND classroom_id = ? AND behavior_type = 'StudyTime'
		ORDER BY create_time DESC
		LIMIT 1
	`
	if err := m.db.Read(ctx, &studyTimeData, studyTimeQuery, studentID, classroomID); err != nil {
		return err
	}
	if studyTimeData.TotalMinutes != nil {
		result.TotalStudyTime = *studyTimeData.TotalMinutes
	}
	m.logger.Debug(ctx, "学习时长数据: %+v", studyTimeData)
	return nil
}

// queryClassroomScore 查询课堂得分
func (m *StudentBehaviorDao) queryClassroomScore(ctx context.Context, result *dto.StudentClassroomDetailDTO, studentID, classroomID uint64) error {
	var scoreData struct {
		ClassroomScore *int64 `ch:"classroom_score"`
	}
	scoreQuery := `
		SELECT 
			JSONExtractInt(context, 'score') AS classroom_score
		FROM tbl_student_behavior_logs
		WHERE student_id = ? AND classroom_id = ? AND behavior_type = 'Score'
		ORDER BY create_time DESC
		LIMIT 1
	`
	if err := m.db.Read(ctx, &scoreData, scoreQuery, studentID, classroomID); err != nil {
		return err
	}
	if scoreData.ClassroomScore != nil {
		result.ClassroomScore = *scoreData.ClassroomScore
	}
	m.logger.Debug(ctx, "课堂得分数据: %+v", scoreData)
	return nil
}

// queryAndUpdateMaxCorrectStreak 查询并更新最大连对数
func (m *StudentBehaviorDao) queryAndUpdateMaxCorrectStreak(ctx context.Context, result *dto.StudentClassroomDetailDTO, studentID, classroomID uint64) error {
	var statsData struct {
		MaxCorrectStreak *int64 `ch:"max_correct_streak"`
	}
	statsQuery := `
		SELECT 
			JSONExtractInt(context, 'correct_streak') AS max_correct_streak
		FROM tbl_student_behavior_logs
		WHERE student_id = ? AND classroom_id = ? AND behavior_type = 'Statistics'
		ORDER BY create_time DESC
		LIMIT 1
	`
	if err := m.db.Read(ctx, &statsData, statsQuery, studentID, classroomID); err != nil {
		return err
	}

	if statsData.MaxCorrectStreak != nil {
		result.MaxCorrectStreak = *statsData.MaxCorrectStreak
		m.logger.Debug(ctx, "统计数据: %+v", statsData)
		return nil
	}

	// 如果最大连对数不存在，获取答题记录
	answers, err := m.GetAnswerRecords(ctx, studentID, classroomID)
	if err != nil {
		return err
	}

	// 更新最大连对数到数据库
	if len(answers) > 0 {
		// 构建上下文数据
		contextData := &dto.StatisticsContext{
			CorrectStreak: result.MaxCorrectStreak,
			Description:   "最大连续正确回答数",
		}
		contextJSON, err := json.Marshal(contextData)
		if err != nil {
			m.logger.Error(ctx, "序列化统计数据失败: %v", err)
			return err
		}

		// 先获取 school_id 和 class_id
		var classInfo struct {
			SchoolID uint64 `ch:"school_id"`
			ClassID  uint64 `ch:"class_id"`
		}
		// 使用封装的 Find 方法查询 school_id 和 class_id
		where := map[string]any{
			"student_id":   studentID,
			"classroom_id": classroomID,
		}
		if err := m.db.Find(ctx, &classInfo, where); err != nil {
			return err
		}

		// 构建行为记录
		behavior := &StudentBehavior{
			SchoolID:     classInfo.SchoolID,
			ClassID:      classInfo.ClassID,
			ClassroomID:  &classroomID,
			StudentID:    studentID,
			BehaviorType: "Statistics",
			Context:      string(contextJSON),
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}

		// 使用封装好的Insert方法插入记录
		_, err = m.db.Insert(ctx, behavior)
		if err != nil {
			m.logger.Error(ctx, "插入统计数据失败: %v", err)
			return err
		}
	}

	return nil
}

// 优化获取学习记录的方法
func (m *StudentBehaviorDao) getStudentLearningRecords(ctx context.Context, studentID, classroomID uint64) ([]dto.LearningRecordDTO, error) {
	m.logger.Debug(ctx, "开始获取学习记录: studentID=%d, classroomID=%d", studentID, classroomID)

	query := `
		WITH answer_stats AS (
			SELECT 
				JSONExtractString(context, 'chapterId') AS chapter_id,
				countIf(JSONExtractInt(context, 'isCorrect') = 1) * 100.0 / count(*) AS accuracy_rate
			FROM tbl_student_behavior_logs
			WHERE student_id = ?
				AND classroom_id = ?
				AND behavior_type = 'Answer'
			GROUP BY chapter_id
		)
		SELECT 
			toString(id) AS record_id,
			JSONExtractString(context, 'chapterId') AS chapter_id,
			JSONExtractString(context, 'chapterName') AS chapter_name,
			JSONExtractString(context, 'learningType') AS learning_type,
			JSONExtractInt(context, 'duration') AS duration,
			JSONExtractFloat(context, 'progress') AS progress,
			create_time,
			if(as.accuracy_rate != 0, as.accuracy_rate, 0) AS accuracy_rate
		FROM tbl_student_behavior_logs
		LEFT JOIN answer_stats AS as ON as.chapter_id = JSONExtractString(context, 'chapterId')
		WHERE student_id = ?
			AND classroom_id = ?
			AND behavior_type = 'Learning'
		ORDER BY create_time DESC
	`

	var records []*struct {
		RecordID     string    `ch:"record_id"`
		ChapterID    string    `ch:"chapter_id"`
		ChapterName  string    `ch:"chapter_name"`
		LearningType string    `ch:"learning_type"`
		Duration     int64     `ch:"duration"`
		Progress     float64   `ch:"progress"`
		CreateTime   time.Time `ch:"create_time"`
		AccuracyRate float64   `ch:"accuracy_rate"`
	}

	// 注意：需要传入studentID和classroomID两次，因为在query中使用了两次
	if err := m.db.Read(ctx, &records, query, studentID, classroomID, studentID, classroomID); err != nil {
		m.logger.Error(ctx, "查询学习记录失败: %v", err)
		return nil, fmt.Errorf("查询学习记录失败: %w", err)
	}

	// 如果没有记录，返回空数组
	if len(records) == 0 {
		m.logger.Warn(ctx, "未找到学习记录: studentID=%d, classroomID=%d", studentID, classroomID)
		return []dto.LearningRecordDTO{}, nil
	}

	result := make([]dto.LearningRecordDTO, len(records))
	for i, r := range records {
		result[i] = dto.LearningRecordDTO{
			RecordID:     r.RecordID,
			ChapterID:    r.ChapterID,
			ChapterName:  r.ChapterName,
			LearningType: r.LearningType,
			Duration:     r.Duration,
			Progress:     r.Progress,
			CreateTime:   r.CreateTime.Unix(),
			AccuracyRate: r.AccuracyRate,
		}
	}

	m.logger.Debug(ctx, "成功获取学习记录: studentID=%d, classroomID=%d, count=%d", studentID, classroomID, len(result))
	return result, nil
}

// GetStudentsBehaviors 获取指定学生ID列表的行为数据
func (d *StudentBehaviorDao) GetStudentsBehaviors(ctx context.Context, studentIDs []uint64) ([]*StudentBehavior, error) {
	if len(studentIDs) == 0 {
		return []*StudentBehavior{}, nil
	}

	// 使用封装好的 DB 方法
	var results []*StudentBehavior
	err := d.DB(ctx).FindAll(ctx, &results, map[string]interface{}{
		"student_id": studentIDs,
	}, nil)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetClassAllBehaviors 获取班级学生所有行为（用于汇总统计）
func (m *StudentBehaviorDao) GetClassAllBehaviors(ctx context.Context, classRoomID uint64) ([]*dto.StudentLatestBehaviorDTO, error) {
	// 定义查询结果接收结构
	var records []*struct {
		StudentID      uint64    `ch:"student_id"`
		BehaviorType   string    `ch:"behavior_type"`
		Context        string    `ch:"context"`
		CreateTime     time.Time `ch:"create_time"`
		StayDuration   uint64    `ch:"stay_duration"`
		TotalQuestions *uint8    `ch:"total_questions"`
		CorrectAnswers *uint8    `ch:"correct_answers"`
	}

	// 使用 WITH 子句优化查询，但不按学生ID分组取最新记录
	// 而是获取所有记录用于汇总统计
	query := `
		SELECT 
			student_id,
			behavior_type,
			context,
			create_time,
			JSONExtractUInt(context, 'stayDuration', 0) as stay_duration,
			if(behavior_type = 'Answer', 
				if(JSONExtractInt(context, 'isCorrect') = 1, 1, 0), 
				0) as correct_answers,
			if(behavior_type = 'Answer', 1, 0) as total_questions
		FROM tbl_student_behavior_logs
		WHERE classroom_id = ? AND behavior_type != 'class_comment'
		ORDER BY student_id DESC
	`

	// 执行查询
	if err := m.db.Read(ctx, &records, query, classRoomID); err != nil {
		m.logger.Error(ctx, "查询班级学生所有行为失败: %v", err)
		return nil, err
	}

	// 转换为 StudentLatestBehaviorDTO 结构
	results := make([]*dto.StudentLatestBehaviorDTO, len(records))
	for i, record := range records {
		// 创建基本行为记录
		behavior := &dto.StudentLatestBehaviorDTO{
			StudentID:      int64(record.StudentID),
			BehaviorType:   consts.BehaviorType(record.BehaviorType),
			Context:        record.Context,
			LastUpdateTime: record.CreateTime.Unix(),
			StayDuration:   int64(record.StayDuration),
			TotalQuestions: 0,
			CorrectAnswers: 0,
		}

		// 安全处理TotalQuestions指针
		if record.TotalQuestions != nil {
			behavior.TotalQuestions = int64(*record.TotalQuestions)
		}

		// 安全处理CorrectAnswers指针
		if record.CorrectAnswers != nil {
			behavior.CorrectAnswers = int64(*record.CorrectAnswers)
		}

		// 计算单条记录正确率（单道题目的正确率只会是0%或100%）
		behavior.AccuracyRate = utils.F64Percent(float64(behavior.CorrectAnswers), float64(behavior.TotalQuestions), 2)

		// 从Context中提取更多信息
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(record.Context), &contextMap); err == nil {
			// 提取学生基本信息
			behavior.StudentName = utils.GetMapStringKey(contextMap, "student_name")
			behavior.AvatarURL = utils.GetMapStringKey(contextMap, "avatar_url")
			// 提取其他信息
			behavior.PageName = utils.GetMapStringKey(contextMap, "page_name")
			behavior.Subject = utils.GetMapStringKey(contextMap, "subject")
			behavior.MaterialID = utils.GetMapUint64Key(contextMap, "material_id")
			behavior.LearningType = utils.GetMapStringKey(contextMap, "learning_type")
			behavior.VideoStatus = utils.GetMapStringKey(contextMap, "video_status")
			behavior.WrongAnswers = utils.GetMapInt64Key(contextMap, "wrong_answers")
		}

		results[i] = behavior
	}

	m.logger.Debug(ctx, "获取班级(%d)学生所有行为成功，共%d条记录", classRoomID, len(results))
	return results, nil
}

// GetStudentBehaviorsByType 获取学生特定类型的行为数据
func (d *StudentBehaviorDao) GetStudentBehaviorsByType(ctx context.Context, studentID, classroomID uint64, behaviorType string) ([]*StudentBehavior, error) {
	if studentID == 0 || classroomID == 0 {
		return []*StudentBehavior{}, nil
	}

	records := make([]*StudentBehavior, 0)

	// 构建查询条件
	whereClause := map[string]interface{}{
		"student_id":    studentID,
		"classroom_id":  classroomID,
		"behavior_type": behaviorType,
	}

	// 执行查询
	err := d.DB(ctx).FindAll(ctx, &records, whereClause, nil)
	if err != nil {
		d.logger.Error(ctx, "查询学生行为数据失败: %v", err)
		return nil, err
	}

	return records, nil
}
