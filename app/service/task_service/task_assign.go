package task_service

import (
	"context"
	"errors"

	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/model/dto"
)

type TaskAssignService struct {
	taskAssignDAO  dao_task.TaskAssignDAO
	taskStudentDAO dao_task.TaskStudentDAO
	log            *logger.ContextLogger
}

func NewTaskAssignService(
	taskAssignDAO dao_task.TaskAssignDAO,
	taskStudentDAO dao_task.TaskStudentDAO,
	log *logger.ContextLogger,
) *TaskAssignService {
	return &TaskAssignService{
		taskAssignDAO:  taskAssignDAO,
		taskStudentDAO: taskStudentDAO,
		log:            log,
	}
}

// 获取指定教师指定对象布置（可能为空）的全部任务布置记录
func (s *TaskAssignService) GetTeacherTasks(ctx context.Context, req *dto.TaskAssignListQuery, pageInfo *consts.DBPageInfo) (int64, []*dao_task.TaskAssign, error) {
	return s.taskAssignDAO.GetTeacherTasks(ctx, req, pageInfo)
}

// 获取指定任务指定布置的统计数据
func (s *TaskAssignService) GetTaskAssignInfo(ctx context.Context, taskID int64, assignID int64) ([]*dao_task.TaskAssign, error) {
	return s.taskAssignDAO.GetTaskAssignInfo(ctx, taskID, assignID)
}

// 获取指定任务指定布置的统计数据
func (s *TaskAssignService) GetTaskAssigns(ctx context.Context, taskID int64, assignIDs []int64) ([]*dao_task.TaskAssign, error) {
	return s.taskAssignDAO.GetTaskAssigns(ctx, taskID, assignIDs)
}

// 获取指定任务的全部布置记录. 返回
// map[taskID]map[assignID]*dao_task.TaskAssign
func (s *TaskAssignService) GetTaskAssignsByTaskIDs(ctx context.Context, taskIDs []int64) (map[int64]map[int64]*dao_task.TaskAssign, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}

	assigns, err := s.taskAssignDAO.GetTaskAssignsByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, err
	}

	assignsMap := make(map[int64]map[int64]*dao_task.TaskAssign)
	for _, assign := range assigns {
		if _, ok := assignsMap[assign.TaskID]; !ok {
			assignsMap[assign.TaskID] = make(map[int64]*dao_task.TaskAssign)
		}
		assignsMap[assign.TaskID][assign.AssignID] = assign
	}
	return assignsMap, nil
}

// 获取指定任务指定布置的学生ID列表
func (s *TaskAssignService) GetTaskAssignStudents(ctx context.Context, taskID int64, assignID int64) ([]int64, error) {
	if taskID == 0 {
		return nil, errors.New("taskID is zero")
	}

	return s.taskStudentDAO.GetTaskAssignStudents(ctx, taskID, assignID)
}

// 获取指定布置ID列表的学生ID列表
func (s *TaskAssignService) GetAssignStudents(ctx context.Context, assignIDs []int64) (map[int64][]int64, error) {
	return s.taskStudentDAO.GetAssignStudents(ctx, assignIDs)
}
