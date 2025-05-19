package dao_task

import (
	"context"
	"errors"
	"gil_teacher/app/core/postgresqlx"

	"gorm.io/gorm"
)

// TaskResource 任务素材资源关联表
type TaskResource struct {
	ID             int64                   `gorm:"column:id;type:bigserial;primaryKey" json:"-"`                   // 自增主键ID
	TaskID         int64                   `gorm:"column:task_id;type:bigint;not null" json:"taskId"`              // 任务ID
	ResourceID     string                  `gorm:"column:resource_id;type:varchar(16);not null" json:"resourceId"` // 资源ID
	ResourceSubIDs postgresqlx.StringArray `gorm:"column:resource_sub_ids;type:text[]" json:"resourceSubIds"`      // 子资源ID列表
	ResourceType   int64                   `gorm:"column:resource_type;type:bigint;not null" json:"resourceType"`  // 资源类型
	ResourceExtra  string                  `gorm:"column:resource_extra;type:text" json:"resourceExtra"`           // 资源额外信息，供前端记录额外信息，后端不解析处理
}

// TableName 指定表名
func (TaskResource) TableName() string {
	return "tbl_task_resource"
}

// ----------------------------------------------
// ----------------------------------------------
// taskResourceDAO 任务资源数据访问实现
type taskResourceDAO struct {
	db *gorm.DB
}

// NewTaskResourceDAO 创建任务资源数据访问实例
func NewTaskResourceDAO(db *gorm.DB) TaskResourceDAO {
	return &taskResourceDAO{db: db}
}

// GetDB 获取数据库连接
func (d *taskResourceDAO) GetDB() *gorm.DB {
	return d.db
}

// DB 返回带有上下文的数据库连接
func (d *taskResourceDAO) DB(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

// Create 创建任务资源关联
func (d *taskResourceDAO) Create(ctx context.Context, resource *TaskResource) error {
	return d.DB(ctx).Create(resource).Error
}

// GetByTaskID 获取任务关联的资源列表
func (d *taskResourceDAO) GetByTaskID(ctx context.Context, taskID int64) ([]*TaskResource, error) {
	var resources []*TaskResource
	err := d.DB(ctx).Where("task_id = ?", taskID).Find(&resources).Error
	return resources, err
}

// GetByResourceID 获取资源关联的任务列表
func (d *taskResourceDAO) GetByResourceID(ctx context.Context, resourceID int64) ([]*TaskResource, error) {
	var resources []*TaskResource
	err := d.DB(ctx).Where("resource_id = ?", resourceID).Find(&resources).Error
	return resources, err
}

// GetTaskResourcesByTaskIDs 获取指定任务列表的资源，按 id 顺序输出
func (d *taskResourceDAO) GetTaskResourcesByTaskIDs(ctx context.Context, taskIDs []int64) ([]*TaskResource, error) {
	resources := make([]*TaskResource, 0)
	err := d.DB(ctx).Where("task_id IN (?)", taskIDs).Order("id ASC").Find(&resources).Error
	if err != nil {
		return nil, err
	}

	return resources, nil
}

// GetTaskResources 获取任务指定资源
func (d *taskResourceDAO) GetTaskResources(ctx context.Context, taskID int64, resourceID string, resourceType int64) ([]*TaskResource, error) {
	if taskID == 0 {
		return nil, errors.New("task_id is required")
	}

	db := d.DB(ctx).Where("task_id = ?", taskID)
	if resourceID != "" && resourceType != 0 {
		db = db.Where("resource_id = ? AND resource_type = ?", resourceID, resourceType)
	}

	var resources []*TaskResource
	err := db.Find(&resources).Error
	return resources, err
}
