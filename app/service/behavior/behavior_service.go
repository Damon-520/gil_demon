package behavior

import (
	"context"

	"gil_teacher/app/core/logger"
	"gil_teacher/app/dao/behavior"
	"gil_teacher/app/model/dto"
)

// BehaviorService 行为服务接口
type BehaviorService interface {
	// 获取学生课堂详情
	GetStudentClassroomDetail(ctx context.Context, studentID, classroomID uint64) (*dto.StudentClassroomDetailDTO, error)
}

// BehaviorServiceImpl 行为服务实现
type BehaviorServiceImpl struct {
	behaviorDAO behavior.BehaviorDAO
	logger      *logger.ContextLogger
}

// NewBehaviorService 创建行为服务
func NewBehaviorService(behaviorDAO behavior.BehaviorDAO, logger *logger.ContextLogger) BehaviorService {
	return &BehaviorServiceImpl{
		behaviorDAO: behaviorDAO,
		logger:      logger,
	}
}

// GetStudentClassroomDetail 获取学生课堂详情
func (s *BehaviorServiceImpl) GetStudentClassroomDetail(ctx context.Context, studentID, classroomID uint64) (*dto.StudentClassroomDetailDTO, error) {
	s.logger.Info(ctx, "获取学生课堂详情, 学生ID: %d, 课堂ID: %d", studentID, classroomID)

	// 通过DAO层获取学生课堂详情
	detail, err := s.behaviorDAO.GetStudentClassroomDetail(ctx, studentID, classroomID)
	if err != nil {
		s.logger.Error(ctx, "获取学生课堂详情失败: %v", err)
		return nil, err
	}

	s.logger.Info(ctx, "成功获取学生课堂详情")
	return detail, nil
}
