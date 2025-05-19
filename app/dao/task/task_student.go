package dao_task

import (
	"context"
	"errors"
	clogger "gil_teacher/app/core/logger"

	"gorm.io/gorm"
)

type taskStudentDao struct {
	db     *gorm.DB
	logger *clogger.ContextLogger
}

func NewTaskStudentDao(db *gorm.DB, logger *clogger.ContextLogger) TaskStudentDAO {
	return &taskStudentDao{
		db:     db,
		logger: logger,
	}
}

// TaskStudent 任务与学生关联表
type TaskStudent struct {
	ID        int64 `gorm:"column:id;type:bigserial;primaryKey"`    // 自增主键ID
	AssignID  int64 `gorm:"column:assign_id;type:bigint;not null"`  // 分配ID，关联任务分配表
	TaskID    int64 `gorm:"column:task_id;type:bigint;not null"`    // 任务ID，关联任务表
	StudentID int64 `gorm:"column:student_id;type:bigint;not null"` // 学生ID，关联学生表
}

// TableName 指定表名
func (TaskStudent) TableName() string {
	return "tbl_task_student"
}

func (t *taskStudentDao) DB(ctx context.Context) *gorm.DB {
	return t.db.Model(&TaskStudent{}).WithContext(ctx)
}

// 获取任务的学生ID列表，如果布置 id 存在，则获取布置对象的学生ID列表，否则获取任务的学生ID列表
func (t *taskStudentDao) GetTaskAssignStudents(ctx context.Context, taskID int64, assignID int64) ([]int64, error) {
	if taskID == 0 {
		return nil, errors.New("taskID is zero")
	}

	query := t.DB(ctx).Where("task_id = ?", taskID)
	if assignID > 0 {
		query = query.Where("assign_id = ?", assignID)
	}

	students := make([]*TaskStudent, 0)
	err := query.Find(&students).Error
	if err != nil {
		t.logger.Error(ctx, "GetTaskAssignStudents error:%v", err)
		return nil, err
	}

	studentIDs := make([]int64, 0, len(students))
	for _, student := range students {
		studentIDs = append(studentIDs, student.StudentID)
	}

	return studentIDs, nil
}

// GetAssignStudents 获取指定布置ID列表的学生ID列表
func (t *taskStudentDao) GetAssignStudents(ctx context.Context, assignIDs []int64) (map[int64][]int64, error) {
	if len(assignIDs) == 0 {
		return nil, errors.New("assignIDs is empty")
	}

	students := make([]*TaskStudent, 0)
	err := t.DB(ctx).Where("assign_id IN (?)", assignIDs).Find(&students).Error
	if err != nil {
		t.logger.Error(ctx, "GetAssignStudents error:%v", err)
		return nil, err
	}

	studentMap := make(map[int64][]int64)
	for _, student := range students {
		studentMap[student.AssignID] = append(studentMap[student.AssignID], student.StudentID)
	}

	return studentMap, nil
}
