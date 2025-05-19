package task_service

import (
	"context"
	"errors"

	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/utils"
)

// TaskResourceService 任务资源服务
type TaskResourceService struct {
	log             *logger.ContextLogger
	taskResourceDAO dao_task.TaskResourceDAO
}

// NewTaskResourceService 创建任务资源服务实例
func NewTaskResourceService(
	log *logger.ContextLogger,
	taskResourceDAO dao_task.TaskResourceDAO,
) *TaskResourceService {
	return &TaskResourceService{
		log:             log,
		taskResourceDAO: taskResourceDAO,
	}
}

// GetTaskResource 查询单个任务的资源列表
func (s *TaskResourceService) GetTaskResource(ctx context.Context, taskID int64, resourceID string, resourceType int64) (map[string]*dao_task.TaskResource, error) {
	resources, err := s.taskResourceDAO.GetTaskResources(ctx, taskID, resourceID, resourceType)
	if err != nil {
		return nil, err
	}
	resourceMap := make(map[string]*dao_task.TaskResource)
	for _, resource := range resources {
		resourceKey := utils.JoinList([]any{resource.ResourceID, resource.ResourceType}, consts.CombineKey)
		resourceMap[resourceKey] = resource
	}
	return resourceMap, nil
}

// 获取单个任务的资源
//
//	map[resourceKey]resource, []resourceID
func (s *TaskResourceService) GetTaskResourceByTaskID(ctx context.Context, taskID int64) (map[string]*dao_task.TaskResource, []string, error) {
	resources, resourceIDs, err := s.GetTaskResourcesByTaskIDs(ctx, []int64{taskID})
	if err != nil {
		return nil, nil, err
	}

	if len(resources) == 0 {
		return nil, nil, errors.New("task resource not found")
	}

	return resources[taskID], resourceIDs[taskID], nil
}

// 获取指定任务列表的资源，和按任务分组顺序排列的资源 id 列表
//
//	map[taskID]map[resourceKey]resource, map[int64]string{taskID: resourceIDs}
func (s *TaskResourceService) GetTaskResourcesByTaskIDs(ctx context.Context, taskIDs []int64) (map[int64]map[string]*dao_task.TaskResource, map[int64][]string, error) {
	if len(taskIDs) == 0 {
		return nil, nil, errors.New("taskIDs is empty")
	}

	resources, err := s.taskResourceDAO.GetTaskResourcesByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, nil, err
	}

	resourceIDsMap := make(map[int64][]string)
	for _, resource := range resources {
		if _, ok := resourceIDsMap[resource.TaskID]; !ok {
			resourceIDsMap[resource.TaskID] = make([]string, 0)
		}
		resourceIDsMap[resource.TaskID] = append(resourceIDsMap[resource.TaskID], resource.ResourceID)
	}

	resourcesMap := make(map[int64]map[string]*dao_task.TaskResource)
	for _, resource := range resources {
		if _, ok := resourcesMap[resource.TaskID]; !ok {
			resourcesMap[resource.TaskID] = make(map[string]*dao_task.TaskResource)
		}
		resourceKey := utils.JoinList([]any{resource.ResourceID, resource.ResourceType}, consts.CombineKey)
		resourcesMap[resource.TaskID][resourceKey] = resource
	}
	return resourcesMap, resourceIDsMap, nil
}

// GetTaskResourcesByTaskID 获取指定任务的资源
func (s *TaskResourceService) GetTaskResourcesByTaskID(ctx context.Context, taskID int64) ([]*dao_task.TaskResource, error) {
	return s.taskResourceDAO.GetByTaskID(ctx, taskID)
}
