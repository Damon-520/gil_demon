package task_service

import (
	"context"
	"gil_teacher/app/core/logger"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/model/api"
)

// TempSelectionService 教师临时选择服务实现
type TempSelectionService struct {
	log          *logger.ContextLogger
	selectionDAO dao_task.TeacherTempSelectionDAO
}

// NewTempSelectionService 创建教师临时选择服务实例
func NewTempSelectionService(log *logger.ContextLogger, selectionDAO dao_task.TeacherTempSelectionDAO) *TempSelectionService {
	return &TempSelectionService{
		log:          log,
		selectionDAO: selectionDAO,
	}
}

// CreateSelection 创建教师临时选择
func (s *TempSelectionService) CreateSelection(ctx context.Context, req *api.CreateTempSelectionRequest) error {
	selection := &dao_task.TeacherTempSelection{
		SchoolID:     req.SchoolID,
		TeacherID:    req.TeacherID,
		ResourceID:   req.ResourceID,
		ResourceType: req.ResourceType,
	}
	return s.selectionDAO.CreateSelection(ctx, selection)
}

// DeleteSelections 删除教师临时选择
func (s *TempSelectionService) DeleteSelections(ctx context.Context, resourceType int64, resourceIDs []string, teacherID int64) error {
	return s.selectionDAO.DeleteSelections(ctx, resourceType, resourceIDs, teacherID)
}

// ListSelections 查询教师临时选择列表
func (s *TempSelectionService) ListSelections(ctx context.Context, teacherID int64) ([]*dao_task.TeacherTempSelection, error) {
	return s.selectionDAO.GetSelectionsByTeacherID(ctx, teacherID)
}
