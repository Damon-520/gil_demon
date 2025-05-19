package task_service

import (
	"context"
	"errors"

	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/service/gil_internal/question_service"
	"gil_teacher/app/utils"
)

// TaskService 任务服务
type TaskService struct {
	log           *logger.ContextLogger
	taskDAO       dao_task.TaskDAO
	taskAssignDAO dao_task.TaskAssignDAO
	questionAPI   *question_service.Client
	redisClient   *dao.ApiRdbClient
}

// NewTaskService 创建任务服务实例
func NewTaskService(
	log *logger.ContextLogger,
	taskDAO dao_task.TaskDAO,
	taskAssignDAO dao_task.TaskAssignDAO,
	questionAPI *question_service.Client,
	redisClient *dao.ApiRdbClient,
) *TaskService {
	return &TaskService{
		log:           log,
		taskDAO:       taskDAO,
		taskAssignDAO: taskAssignDAO,
		questionAPI:   questionAPI,
		redisClient:   redisClient,
	}
}

// GetLastUsedBizTree 获取最近使用过的业务树
func (s *TaskService) GetLastUsedBizTree(ctx context.Context, teacherID int64, subject int64) (int64, error) {
	key := consts.TaskLastUsedBizTreeKey(teacherID, subject)
	var val int64
	_, err := s.redisClient.Get(ctx, key, &val)
	return val, err
}

// RecordLastUsedBizTree 记录最近使用过的业务树
func (s *TaskService) RecordLastUsedBizTree(ctx context.Context, teacherID int64, subject int64, bizTreeID int64) error {
	key := consts.TaskLastUsedBizTreeKey(teacherID, subject)
	return s.redisClient.Set(ctx, key, bizTreeID, consts.CommunicationSessionExpire)
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(ctx context.Context, reqBody *api.CreateTaskRequestBody) error {
	// 如果资源类型为巩固练习，则将下面所有的题目 ID 记录到 resourceSubIDs
	questionSetIDs := []int64{}
	resourceID2resourceSubIDs := make(map[string][]string)
	// 先遍历记录巩固练习的ID
	for _, resource := range reqBody.Resources {
		if resource.ResourceType == consts.RESOURCE_TYPE_PRACTICE {
			questionSetIDs = append(questionSetIDs, utils.Atoi64(resource.ResourceID))
			resourceID2resourceSubIDs[resource.ResourceID] = []string{}
		}
	}
	// 再向题库平台查询巩固练习下面的题目ID
	if len(questionSetIDs) > 0 {
		questionSetList, err := s.questionAPI.GetQuestionSetListByIDs(ctx, questionSetIDs)
		if err != nil {
			s.log.Error(ctx, "获取巩固练习下面的题目ID失败: %v", err)
			return err
		}
		// 三级结构：题集-题组-题目
		for _, questionSet := range questionSetList {
			if questionSet != nil {
				questionIDs := make([]string, 0)
				for _, questionGroup := range questionSet.QuestionGroupStableInfoList {
					for _, question := range questionGroup.QuestionInfoList {
						questionIDs = append(questionIDs, question.QuestionId)
					}
				}
				resourceID2resourceSubIDs[utils.I64ToStr(questionSet.QuestionSetId)] = questionIDs
			}
		}
	}

	// 1. 构建任务实体
	task := &dao_task.Task{
		SchoolID:       reqBody.SchoolID,
		Phase:          reqBody.Phase,
		Subject:        reqBody.Subject,
		TaskType:       reqBody.TaskType,
		TaskName:       reqBody.TaskName,
		TeacherComment: reqBody.TeacherComment,
		TaskExtraInfo:  reqBody.TaskExtraInfo,
		CreatorID:      reqBody.CreatorID,
		UpdaterID:      reqBody.UpdaterID,
	}

	// 开启事务
	tx := s.taskDAO.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		s.log.Error(ctx, "开启事务失败: %v", tx.Error)
		return tx.Error
	}

	// 使用事务创建任务
	if err := tx.Create(task).Error; err != nil {
		tx.Rollback()
		s.log.Error(ctx, "创建任务失败: %v", err)
		return err
	}

	// 创建任务资源关联
	if len(reqBody.Resources) > 0 {
		resources := make([]*dao_task.TaskResource, 0, len(reqBody.Resources))
		for _, res := range reqBody.Resources {
			resourceSubIDs := []string{}
			if subIDs, ok := resourceID2resourceSubIDs[res.ResourceID]; ok {
				resourceSubIDs = subIDs
			}
			resource := &dao_task.TaskResource{
				TaskID:         task.TaskID,
				ResourceID:     res.ResourceID,
				ResourceSubIDs: resourceSubIDs,
				ResourceType:   res.ResourceType,
				ResourceExtra:  res.ResourceExtra,
			}
			resources = append(resources, resource)
		}

		// 使用事务批量创建资源关联
		if err := tx.Create(resources).Error; err != nil {
			tx.Rollback()
			s.log.Error(ctx, "创建任务资源关联失败: %v", err)
			return err
		}
	}

	// 创建任务学生群组关联
	taskAssigns := make([]*dao_task.TaskAssign, 0, len(reqBody.StudentGroups))
	for _, group := range reqBody.StudentGroups {
		taskAssign := &dao_task.TaskAssign{
			TaskID:    task.TaskID,
			SchoolID:  reqBody.SchoolID,
			GroupType: group.GroupType,
			GroupID:   group.GroupID,
			StartTime: group.StartTime,
			Deadline:  group.Deadline,
		}
		taskAssigns = append(taskAssigns, taskAssign)
	}

	// 批量创建任务学生群组关联
	if len(taskAssigns) > 0 {
		if err := tx.Create(taskAssigns).Error; err != nil {
			tx.Rollback()
			s.log.Error(ctx, "批量创建任务学生群组关联失败: %v", err)
			return err
		}
	}

	// 批量创建任务学生关联
	taskStudents := make([]*dao_task.TaskStudent, 0)
	for _, group := range reqBody.StudentGroups {
		// 找到对应的任务分配记录
		var assignID int64
		for _, assign := range taskAssigns {
			if assign.GroupType == group.GroupType && assign.GroupID == group.GroupID {
				assignID = assign.AssignID
				break
			}
		}

		// 添加班级学生
		for _, studentID := range group.StudentIDs {
			taskStudents = append(taskStudents, &dao_task.TaskStudent{
				AssignID:  assignID,
				TaskID:    task.TaskID,
				StudentID: studentID,
			})
		}
	}

	// 批量创建任务学生关联
	if len(taskStudents) > 0 {
		if err := tx.Create(taskStudents).Error; err != nil {
			tx.Rollback()
			s.log.Error(ctx, "批量创建任务学生关联失败: %v", err)
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		s.log.Error(ctx, "提交事务失败: %v", err)
		return err
	}

	return nil
}

// GetTaskByIDAndCreatorID 获取指定任务
func (s *TaskService) GetTaskByIDAndCreatorID(ctx context.Context, taskID int64, creatorID int64) (*dao_task.Task, error) {
	return s.taskDAO.GetTaskByIDAndCreatorID(ctx, taskID, creatorID)
}

// DeleteTask 批量软删除任务
func (s *TaskService) DeleteTask(ctx context.Context, taskIDs []int64, teacherID int64) error {
	if err := s.taskDAO.DeleteTasks(ctx, taskIDs, teacherID); err != nil {
		s.log.Error(ctx, "软删除任务失败: %v", err)
		return err
	}
	return nil
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(ctx context.Context, taskID int64, creatorID int64, updates map[string]interface{}) error {
	if err := s.taskDAO.UpdateTask(ctx, taskID, creatorID, updates); err != nil {
		s.log.Error(ctx, "更新任务失败: %v", err)
		return err
	}
	return nil
}

// UpdateTaskAssign 更新任务分配
func (s *TaskService) UpdateTaskAssign(ctx context.Context, assginID int64, updates map[string]interface{}) error {
	if err := s.taskAssignDAO.UpdateTaskAssign(ctx, assginID, updates); err != nil {
		s.log.Error(ctx, "更新任务失败: %v", err)
		return err
	}
	return nil
}

// DeleteTaskAssign 删除任务分配
func (s *TaskService) DeleteTaskAssign(ctx context.Context, taskID int64, assignIDs []int64, creatorID int64) error {
	if err := s.taskAssignDAO.DeleteTaskAssign(ctx, taskID, assignIDs); err != nil {
		s.log.Error(ctx, "删除任务分配失败: %v", err)
		return err
	}
	// 如果分配的对象全部被删除则任务也一并删除
	count, err := s.taskAssignDAO.CountTaskAssignByTaskID(ctx, taskID)
	if err != nil {
		s.log.Error(ctx, "统计任务分配数量失败: %v", err)
		return err
	}
	if count == 0 {
		if err := s.taskDAO.DeleteTasks(ctx, []int64{taskID}, creatorID); err != nil {
			s.log.Error(ctx, "删除任务失败: %v", err)
			return err
		}
	}
	return nil
}

// GetResourceAssignedClassIDs 获取资源已布置的班级ID列表
func (s *TaskService) GetResourceAssignedClassIDs(ctx context.Context, req *dto.TaskResourceAssignedClassIDsRequest) (map[string][]int64, error) {
	resourceGroupIDs, err := s.taskAssignDAO.GetResourceAssignedClassIDs(ctx, req)
	if err != nil {
		return nil, err
	}
	// map[resource_id][]group_id
	result := make(map[string][]int64)
	for _, item := range resourceGroupIDs {
		result[item.ResourceID] = append(result[item.ResourceID], item.GroupID)
	}
	return result, nil
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks(ctx context.Context, conditions map[string]interface{}, page, pageSize int64) ([]*dao_task.Task, int64, error) {
	return s.taskDAO.ListTasks(ctx, conditions, page, pageSize)
}

// 获取指定用户创建的全部任务列表
func (s *TaskService) GetUserTasks(ctx context.Context, reqs *dto.TaskAssignListQuery, pageInfo *consts.DBPageInfo) ([]*dao_task.Task, error) {
	return s.taskDAO.GetUserTasks(ctx, reqs, pageInfo)
}

// 获取单个任务信息
func (s *TaskService) GetTaskByID(ctx context.Context, taskID int64) (*dao_task.Task, error) {
	tasks, err := s.taskDAO.GetTasksByIDs(ctx, []int64{taskID})
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, errors.New("任务不存在")
	}
	return tasks[0], nil
}

// 获取指定任务列表
func (s *TaskService) GetTasksByIDs(ctx context.Context, taskIDs []int64) (map[int64]*dao_task.Task, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	tasks, err := s.taskDAO.GetTasksByIDs(ctx, taskIDs)
	if err != nil {
		return nil, err
	}
	taskMap := make(map[int64]*dao_task.Task)
	for _, task := range tasks {
		taskMap[task.TaskID] = task
	}
	return taskMap, nil
}

// 获取指定教师最近一次布置的任务
func (s *TaskService) GetLatestTask(ctx context.Context, teacherID int64, schoolID int64, subjectID int64) ([]*dao_task.Task, error) {
	if teacherID == 0 || schoolID == 0 {
		return nil, nil
	}
	return s.taskDAO.GetLatestTask(ctx, teacherID, schoolID, subjectID)
}

// GetTaskAssignsByTaskIDs 获取指定任务的分配
func (s *TaskService) GetTaskAssignsByTaskIDs(ctx context.Context, taskIDs []int64) ([]*dao_task.TaskAssign, error) {
	return s.taskAssignDAO.GetTaskAssignsByTaskIDs(ctx, taskIDs)
}

// GetStudentTaskList 获取指定学生的任务列表
func (s *TaskService) GetStudentTaskList(ctx context.Context, req *dto.StudentTaskListQuery) ([]*dao_task.TaskAndTaskAssign, int64, error) {
	return s.taskDAO.GetStudentTaskList(ctx, req)
}
