package resource_favorite

import (
	"context"

	"gorm.io/gorm"
)

// TeacherResourceFavorite 教师资源收藏表
type TeacherResourceFavorite struct {
	ID           int64  `gorm:"column:id;type:bigserial;primaryKey"`          // 自增主键ID
	TeacherID    int64  `gorm:"column:teacher_id;type:bigint;not null"`       // 教师ID
	SchoolID     int64  `gorm:"column:school_id;type:bigint;not null"`        // 学校ID
	ResourceID   string `gorm:"column:resource_id;type:varchar(16);not null"` // 资源ID
	ResourceType int64  `gorm:"column:resource_type;type:bigint;not null"`    // 资源类型（0:默认类型）
	Status       int64  `gorm:"column:status;type:bigint;not null"`           // 收藏状态（1:已收藏 0:已取消）
	CreateTime   int64  `gorm:"column:create_time;type:bigint"`               // 创建时间
	UpdateTime   int64  `gorm:"column:update_time;type:bigint"`               // 更新时间
}

// TableName 表名
func (t *TeacherResourceFavorite) TableName() string {
	return "tbl_teacher_resource_favorite"
}

// ResourceFavoriteDAO 资源收藏DAO
type ResourceFavoriteDAO struct {
	db *gorm.DB
}

// NewResourceFavoriteDAO 创建资源收藏DAO
func NewResourceFavoriteDAO(db *gorm.DB) *ResourceFavoriteDAO {
	return &ResourceFavoriteDAO{
		db: db,
	}
}

// Create 创建资源收藏
func (d *ResourceFavoriteDAO) Create(ctx context.Context, favorite *TeacherResourceFavorite) error {
	return d.db.WithContext(ctx).Create(favorite).Error
}

// Update 更新资源收藏
func (d *ResourceFavoriteDAO) Update(ctx context.Context, favorite *TeacherResourceFavorite) error {
	return d.db.WithContext(ctx).Save(favorite).Error
}

// GetByTeacherAndResource 根据教师ID和资源ID获取收藏
func (d *ResourceFavoriteDAO) GetByTeacherAndResource(ctx context.Context, teacherID int64, resourceID string) (*TeacherResourceFavorite, error) {
	var favorite TeacherResourceFavorite
	err := d.db.WithContext(ctx).Where("teacher_id = ? AND resource_id = ?", teacherID, resourceID).First(&favorite).Error
	if err != nil {
		return nil, err
	}
	return &favorite, nil
}

// List 获取收藏列表
func (d *ResourceFavoriteDAO) List(ctx context.Context, teacherID int64, schoolID int64, offset, limit int64) ([]*TeacherResourceFavorite, int64, error) {
	var total int64
	err := d.db.WithContext(ctx).Model(&TeacherResourceFavorite{}).
		Where("teacher_id = ? AND school_id = ? AND status = 1", teacherID, schoolID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var favorites []*TeacherResourceFavorite
	err = d.db.WithContext(ctx).
		Where("teacher_id = ? AND school_id = ? AND status = 1", teacherID, schoolID).
		Order("create_time DESC").
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&favorites).Error
	if err != nil {
		return nil, 0, err
	}

	return favorites, total, nil
}
