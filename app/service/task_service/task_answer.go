package task_service

import (
	"context"
	"errors"

	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils"
)

type TaskAnswerService struct {
	taskCompleteDetailsDao dao_task.TaskStudentDetailsDao
	log                    *logger.ContextLogger
}

func NewTaskAnswerService(taskCompleteDetailsDao dao_task.TaskStudentDetailsDao, log *logger.ContextLogger) *TaskAnswerService {
	return &TaskAnswerService{
		taskCompleteDetailsDao: taskCompleteDetailsDao,
		log:                    log,
	}
}

// 获取指定任务指定布置指定学生的作答结果
// 返回 map[resource_id#resource_type#question_id]*dao_task.TaskStudentDetails, totalCount, incorrectCount
func (s *TaskAnswerService) GetTaskStudentAnswers(ctx context.Context, studentId int64, query *dto.StudentTaskReportQuery, pageInfo *consts.DBPageInfo) (map[string]*dao_task.TaskStudentDetails, int64, int64, error) {
	if query.TaskID == 0 || query.AssignID == 0 || studentId == 0 {
		return nil, 0, 0, errors.New("taskID, assignID, studentID is required")
	}

	pageInfo.Check()
	studentAnswers, totalCount, incorrectCount, err := s.taskCompleteDetailsDao.GetTaskStudentAnswers(ctx, studentId, query, pageInfo)
	if err != nil {
		return nil, 0, 0, err
	}

	studentAnswersMap := make(map[string]*dao_task.TaskStudentDetails)
	for _, answer := range studentAnswers {
		questionKey := utils.JoinList([]any{answer.ResourceKey, answer.QuestionID}, consts.CombineKey)
		studentAnswersMap[questionKey] = answer
	}

	return studentAnswersMap, totalCount, incorrectCount, nil
}

// 获取指定任务指定布置指定题目作答统计数据
func (s *TaskAnswerService) GetTaskAnswerCount(ctx context.Context, taskID, assignID int64, resourceQuestionIDs []string) (*dto.TaskAnswerStat, error) {
	if taskID == 0 || assignID == 0 {
		return nil, errors.New("taskID, assignID is required")
	}
	answerStat, err := s.taskCompleteDetailsDao.GetTaskAnswerCountStat(ctx, taskID, assignID, resourceQuestionIDs)
	if err != nil {
		return nil, err
	}

	return answerStat, nil
}
