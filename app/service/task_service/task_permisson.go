package task_service

import (
	"gil_teacher/app/core/logger"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/middleware"

	"github.com/gin-gonic/gin"
)

type TaskPermissionService struct {
	taskDAO           dao_task.TaskDAO
	taskAssignDAO     dao_task.TaskAssignDAO
	teacherMiddleware *middleware.TeacherMiddleware
	log               *logger.ContextLogger
}

func NewTaskPermissionService(taskDAO dao_task.TaskDAO, taskAssignDAO dao_task.TaskAssignDAO, log *logger.ContextLogger) *TaskPermissionService {
	return &TaskPermissionService{taskDAO: taskDAO, taskAssignDAO: taskAssignDAO, log: log}
}

// 检查教师对任务的权限
// 1. 班主任有班级所有科目的权限
// 2. 学科教师有指定科目的权限
// 3. 其他教师无权限
func (s *TaskPermissionService) CheckTaskPermission(ctx *gin.Context, taskID int64) (bool, error) {
	tasks, err := s.taskDAO.GetTasksByIDs(ctx, []int64{taskID})
	if err != nil {
		return false, err
	}

	if len(tasks) == 0 {
		return false, nil
	}

	taskAssigns, err := s.taskAssignDAO.GetTaskAssignsByTaskIDs(ctx, []int64{taskID})
	if err != nil {
		return false, err
	}

	if len(taskAssigns) == 0 {
		return false, nil
	}

	taskAssign := taskAssigns[0]
	subjectID := tasks[0].Subject
	auth := s.teacherMiddleware.CheckSubjectPermission(ctx, taskAssign.GroupID, subjectID)
	return auth, nil
}
