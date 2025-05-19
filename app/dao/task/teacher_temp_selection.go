package dao_task

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// TeacherTempSelection 教师临时选择表（试题篮、资源篮）
type TeacherTempSelection struct {
	ID           int64  `gorm:"column:id;type:bigserial;primaryKey" json:"id"`                                                           // 自增主键ID
	SchoolID     int64  `gorm:"column:school_id;type:bigint;not null" json:"school_id"`                                                  // 学校ID
	TeacherID    int64  `gorm:"column:teacher_id;type:bigint;not null" json:"teacher_id"`                                                // 教师ID
	ResourceID   string `gorm:"column:resource_id;type:varchar;not null" json:"resource_id"`                                             // 资源ID
	ResourceType int64  `gorm:"column:resource_type;type:bigint;not null" json:"resource_type"`                                          // 资源类型
	CreateTime   int64  `gorm:"column:create_time;type:bigint;default:EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT" json:"create_time"` // 记录创建时间（UTC秒数）
}

// TableName 指定表名
func (TeacherTempSelection) TableName() string {
	return "tbl_teacher_temp_selection"
}

// ----------------------------------------------
// ----------------------------------------------
// teacherTempSelectionDAO 教师临时选择数据访问实现
type teacherTempSelectionDAO struct {
	db *gorm.DB
}

// NewTeacherTempSelectionDAO 创建教师临时选择数据访问实例
func NewTeacherTempSelectionDAO(db *gorm.DB) TeacherTempSelectionDAO {
	return &teacherTempSelectionDAO{db: db}
}

// DB 返回带有上下文的数据库连接
func (d *teacherTempSelectionDAO) DB(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

// CreateSelection 创建教师临时选择
func (d *teacherTempSelectionDAO) CreateSelection(ctx context.Context, selection *TeacherTempSelection) error {
	err := d.DB(ctx).Create(selection).Error
	// 忽略 duplicate key，重复添加视为正常
	if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
		return err
	}
	return nil
}

// DeleteSelection 删除教师临时选择
func (d *teacherTempSelectionDAO) DeleteSelections(ctx context.Context, resourceType int64, resourceIDs []string, teacherID int64) error {
	if err := d.DB(ctx).Model(&TeacherTempSelection{}).Where("resource_type = ? AND resource_id IN ? AND teacher_id = ?", resourceType, resourceIDs, teacherID).Delete(&TeacherTempSelection{}).Error; err != nil {
		return err
	}
	return nil
}

// GetSelectionsByTeacherID 获取教师的临时选择列表
func (d *teacherTempSelectionDAO) GetSelectionsByTeacherID(ctx context.Context, teacherID int64) ([]*TeacherTempSelection, error) {
	var selections []*TeacherTempSelection
	err := d.DB(ctx).Where("teacher_id = ?", teacherID).Find(&selections).Error
	return selections, err
}
