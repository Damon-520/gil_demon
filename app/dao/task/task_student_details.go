package dao_task

import (
	"context"
	"errors"
	"strings"
	"time"

	"gil_teacher/app/consts"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils"

	"gorm.io/gorm"
)

type taskStudentDetailsDao struct {
	db     *gorm.DB
	logger *clogger.ContextLogger
}

func NewTaskStudentDetailsDao(db *gorm.DB, logger *clogger.ContextLogger) TaskStudentDetailsDao {
	return &taskStudentDetailsDao{
		db:     db,
		logger: logger,
	}
}

type TaskStudentDetails struct {
	// uniq_key: task_id#assign_id#student_id#resource_key#question_id
	ID            int64  `gorm:"column:id"`
	TaskID        int64  `gorm:"column:task_id"`
	AssignID      int64  `gorm:"column:assign_id"`
	ResourceKey   string `gorm:"column:resource_key"` // resource_id#resource_type
	QuestionID    string `gorm:"column:question_id"`
	StudentID     int64  `gorm:"column:student_id"`
	AnswerContent string `gorm:"column:answer_content"`
	Correctness   bool   `gorm:"column:correctness"`
	CostTime      int64  `gorm:"column:cost_time"`
	CreateTime    int64  `gorm:"column:create_time"`
	UpdateTime    int64  `gorm:"column:update_time"`
}

type QuestionAnswer struct {
	AnswerCount    int64   `gorm:"column:answer_count"`    // 答题数
	IncorrectCount int64   `gorm:"column:incorrect_count"` // 答错数
	Accuracy       float64 `gorm:"column:accuracy"`        // 正确率
}

func (m *TaskStudentDetails) TableName() string {
	return "tbl_task_student_details"
}

func (m *taskStudentDetailsDao) DB(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx).Model(&TaskStudentDetails{})
}

// 插入任务完成详情
func (d *taskStudentDetailsDao) Insert(ctx context.Context, details *TaskStudentDetails) error {
	if details == nil {
		return errors.New("entity is nil")
	}

	details.CreateTime = time.Now().Unix()
	details.UpdateTime = time.Now().Unix()
	return d.DB(ctx).Create(details).Error
}

// 更新任务完成详情
func (d *taskStudentDetailsDao) Update(ctx context.Context, id int64, details *TaskStudentDetails) error {
	if details == nil {
		return errors.New("entity is nil")
	}
	details.UpdateTime = time.Now().Unix()
	return d.DB(ctx).Where("id = ?", id).Updates(details).Error
}

// 查询指定任务指定学生的完成详情
// queryAll true 查询全部题目，false 查询错误题目
func (d *taskStudentDetailsDao) GetTaskStudentAnswers(ctx context.Context, studentID int64, query *dto.StudentTaskReportQuery, pageInfo *consts.DBPageInfo) ([]*TaskStudentDetails, int64, int64, error) {
	if query.TaskID == 0 || query.AssignID == 0 || studentID == 0 {
		return nil, 0, 0, errors.New("taskID, assignID, studentID is required")
	}

	// 用于接收聚合查询结果
	type CountResult struct {
		TotalCount   int64  `gorm:"column:total_count"`
		IncorrectNum *int64 `gorm:"column:incorrect_num"`
	}

	var countRes CountResult
	countDB := d.DB(ctx).Where("task_id = ? AND assign_id = ? AND student_id = ?", query.TaskID, query.AssignID, studentID)

	// 1. 获取总数和错误数
	if err := countDB.Select("COUNT(*) as total_count, SUM(CASE WHEN correctness = false THEN 1 ELSE 0 END) as incorrect_num").
		Scan(&countRes).Error; err != nil {
		d.logger.Error(ctx, "GetStudentAnswers count query error: %v, query: %+v", err, query)
		return nil, 0, 0, err
	}

	// 2. 获取详情列表
	var answerDetails []*TaskStudentDetails
	detailsDB := d.DB(ctx).Where("task_id = ? AND assign_id = ? AND student_id = ?", query.TaskID, query.AssignID, studentID)

	if !query.AllQuestions {
		detailsDB = detailsDB.Where("correctness = ?", false)
	}
	if query.ResourceID != "" {
		resourceKey := utils.JoinList([]any{query.ResourceID, query.ResourceType}, consts.CombineKey)
		detailsDB = detailsDB.Where("resource_key = ?", resourceKey)
	}
	if len(query.QuestionIDs) > 0 {
		detailsDB = detailsDB.Where("question_id IN ?", query.QuestionIDs)
	}

	// 应用分页
	pageInfo.Check()
	detailsDB = detailsDB.Limit(int(pageInfo.Limit)).Offset(int((pageInfo.Page - 1) * pageInfo.Limit))

	if err := detailsDB.Find(&answerDetails).Error; err != nil {
		d.logger.Error(ctx, "GetStudentAnswers details query error: %v, query: %+v, pageInfo: %+v", err, query, pageInfo)
		return nil, 0, 0, err
	}

	// 处理 IncorrectNum 为 NULL 的情况
	incorrectNum := int64(0)
	if countRes.IncorrectNum != nil {
		incorrectNum = *countRes.IncorrectNum
	}

	return answerDetails, countRes.TotalCount, incorrectNum, nil
}

// 查询指定任务布置指定资源指定题目的全部作答结果
func (d *taskStudentDetailsDao) GetTaskAssignAnswerDetails(ctx context.Context, query *dto.TaskAssignAnswersQuery, pageInfo *consts.DBPageInfo) ([]*TaskStudentDetails, error) {
	if query.TaskID == 0 || query.AssignID == 0 {
		return nil, errors.New("taskID, assignID is required")
	}

	detailsDB := d.DB(ctx).Where("task_id = ? AND assign_id = ?", query.TaskID, query.AssignID)
	if query.ResourceID != "" {
		resourceKey := utils.JoinList([]any{query.ResourceID, query.ResourceType}, consts.CombineKey)
		detailsDB = detailsDB.Where("resource_key = ?", resourceKey)
	}
	if len(query.QuestionIDs) > 0 {
		detailsDB = detailsDB.Where("question_id IN ?", query.QuestionIDs)
	}
	if len(query.StudentIDs) > 0 {
		detailsDB = detailsDB.Where("student_id IN ?", query.StudentIDs)
	}

	// 应用分页
	pageInfo.Check()
	detailsDB = detailsDB.Limit(int(pageInfo.Limit)).Offset(int((pageInfo.Page - 1) * pageInfo.Limit))

	var details []*TaskStudentDetails
	if err := detailsDB.Find(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}

// 查询任务答题结果计数，题目:作答人数，题目:答错人数，题目:总用时
func (d *taskStudentDetailsDao) GetTaskAnswerCountStat(ctx context.Context, taskID int64, assignID int64, resourceQuestionIDs []string) (*dto.TaskAnswerStat, error) {
	type CountResult struct {
		ResourceKey    string `gorm:"column:resource_key"`
		QuestionID     string `gorm:"column:question_id"`
		AnswerCount    int64  `gorm:"column:answer_count"`
		IncorrectCount int64  `gorm:"column:incorrect_count"`
		TotalCostTime  int64  `gorm:"column:total_cost_time"`
	}
	countResult := make([]*CountResult, 0)

	db := d.DB(ctx).
		Select("resource_key, question_id, COUNT(*) as answer_count, SUM(CASE WHEN correctness = false THEN 1 ELSE 0 END) as incorrect_count, SUM(cost_time) as total_cost_time").
		Where("task_id = ? AND assign_id = ?", taskID, assignID)

	if len(resourceQuestionIDs) > 0 {
		db = db.Where("question_id IN ?", resourceQuestionIDs)
	}

	if err := db.Group("resource_key, question_id").
		Scan(&countResult).Error; err != nil {
		d.logger.Error(ctx, "GetTaskAnswerCount error: %v", err)
		return nil, err
	}

	answerMap := make(map[string]int64)
	incorrectMap := make(map[string]int64)
	totalCostTimeMap := make(map[string]int64)
	for _, result := range countResult {
		questionKey := utils.JoinList([]any{result.ResourceKey, result.QuestionID}, consts.CombineKey)
		answerMap[questionKey] = result.AnswerCount
		incorrectMap[questionKey] = result.IncorrectCount
		totalCostTimeMap[questionKey] = result.TotalCostTime
	}

	return &dto.TaskAnswerStat{
		ResourceAnswerCount:    answerMap,
		ResourceIncorrectCount: incorrectMap,
		ResourceTotalCostTime:  totalCostTimeMap,
	}, nil
}

// 获取某个任务布置下每个题目的正确率，同样要满足筛选条件
func (d *taskStudentDetailsDao) GetTaskAnswerAccuracy(ctx context.Context, query *dto.TaskReportCommonQuery) (map[string]float64, error) {
	type AnswerPanelResult struct {
		ResourceKey    string `gorm:"column:resource_key"`
		QuestionID     string `gorm:"column:question_id"`
		AnswerCount    int64  `gorm:"column:answer_count"`
		IncorrectCount int64  `gorm:"column:incorrect_count"`
	}
	answerPanelResult := make([]*AnswerPanelResult, 0)

	db := d.DB(ctx).
		Select("resource_key, question_id, COUNT(*) as answer_count, SUM(CASE WHEN correctness = false THEN 1 ELSE 0 END) as incorrect_count").
		Where("task_id = ? AND assign_id = ?", query.TaskID, query.AssignID)

	if query.ResourceID != "" {
		resourceKey := utils.JoinList([]any{query.ResourceID, query.ResourceType}, consts.CombineKey)
		db = db.Where("resource_key = ?", resourceKey)
	}

	if err := db.Group("resource_key, question_id").
		Order("resource_key, question_id").
		Scan(&answerPanelResult).Error; err != nil {
		d.logger.Error(ctx, "GetTaskAnswerPanel error: %v", err)
		return nil, err
	}

	answerPanelMap := make(map[string]float64)
	for _, result := range answerPanelResult {
		questionKey := utils.JoinList([]any{result.ResourceKey, result.QuestionID}, consts.CombineKey)
		answerPanelMap[questionKey] = utils.F64Div(float64(result.AnswerCount-result.IncorrectCount), float64(result.AnswerCount), 2)
	}

	return answerPanelMap, nil
}

// taskID, assignID 指定任务布置，resourceQuestionIDs 指定资源包含题目列表
//
//	map[questionKey]*QuestionAnswerAccuracy
func (d *taskStudentDetailsDao) GetQuestionAnswers(ctx context.Context, taskID, assignID int64, resourceQuestionIDs map[string][]string) (map[string]*QuestionAnswer, error) {
	type AnswerPanelResult struct {
		ResourceKey    string `gorm:"column:resource_key"`
		QuestionID     string `gorm:"column:question_id"`
		AnswerCount    int64  `gorm:"column:answer_count"`
		IncorrectCount int64  `gorm:"column:incorrect_count"`
	}
	answerPanelResult := make([]*AnswerPanelResult, 0)

	db := d.DB(ctx).
		Select("resource_key, question_id, COUNT(*) as answer_count, SUM(CASE WHEN correctness = false THEN 1 ELSE 0 END) as incorrect_count").
		Where("task_id = ? AND assign_id = ?", taskID, assignID)

	// or 组合查询
	conditions := make([]string, 0)
	args := make([]any, 0)
	for resourceKey, questionIDs := range resourceQuestionIDs {
		if len(questionIDs) == 0 {
			continue
		}
		condition := "(resource_key = ? AND question_id IN ?)"
		conditions = append(conditions, condition)
		args = append(args, resourceKey, questionIDs)
	}
	if len(conditions) > 0 {
		db = db.Where(strings.Join(conditions, " OR "), args...)
	}

	if err := db.Group("resource_key, question_id").
		Order("resource_key, question_id").
		Scan(&answerPanelResult).Error; err != nil {
		d.logger.Error(ctx, "GetTaskAnswerPanel error: %v", err)
		return nil, err
	}

	answerPanelMap := make(map[string]*QuestionAnswer)
	for _, result := range answerPanelResult {
		questionKey := utils.JoinList([]any{result.ResourceKey, result.QuestionID}, consts.CombineKey)
		answerPanelMap[questionKey] = &QuestionAnswer{
			AnswerCount:    result.AnswerCount,
			IncorrectCount: result.IncorrectCount,
			Accuracy:       utils.F64Div(float64(result.AnswerCount-result.IncorrectCount), float64(result.AnswerCount), 4),
		}
	}

	return answerPanelMap, nil
}

// 计算指定布置任务每个题的答题时间和答题人数，用于计算每个题目的平均用时
// 返回值：map[string]int64, map[string]int64
// 第一个map：key为题目key，value为答题人数
// 第二个map：key为题目key，value为答题时间总和
func (d *taskStudentDetailsDao) GetTaskAnswerTime(ctx context.Context, taskID int64, assignID int64, resourceQuestionIDs []string) (map[string]int64, map[string]int64, error) {
	type AnswerTimeResult struct {
		ResourceKey   string `gorm:"column:resource_key"`
		QuestionID    string `gorm:"column:question_id"`
		AnswerCount   int64  `gorm:"column:answer_count"`
		TotalCostTime int64  `gorm:"column:total_cost_time"`
	}
	answerTimeResult := make([]*AnswerTimeResult, 0)

	db := d.DB(ctx).
		Select("resource_key, question_id, COUNT(*) as answer_count, SUM(cost_time) as total_cost_time").
		Where("task_id = ? AND assign_id = ?", taskID, assignID)

	if len(resourceQuestionIDs) > 0 {
		db = db.Where("question_id IN ?", resourceQuestionIDs)
	}

	if err := db.Group("resource_key, question_id").
		Order("resource_key, question_id").
		Scan(&answerTimeResult).Error; err != nil {
		d.logger.Error(ctx, "GetTaskAnswerTime error: %v", err)
		return nil, nil, err
	}

	answerTimeMap := make(map[string]int64)
	answerCountMap := make(map[string]int64)
	for _, result := range answerTimeResult {
		questionKey := utils.JoinList([]any{result.ResourceKey, result.QuestionID}, consts.CombineKey)
		answerTimeMap[questionKey] = result.TotalCostTime
		answerCountMap[questionKey] = result.AnswerCount
	}

	return answerCountMap, answerTimeMap, nil
}
