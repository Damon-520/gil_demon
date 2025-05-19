package file

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gil_teacher/app/core/logger"

	"gorm.io/gorm"
)

// Resource 资源实体，对应tbl_resource表
type Resource struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement:true"` // 自增主键ID
	UserID       int64      `gorm:"column:user_id"`                          // 上传资源的用户ID
	SchoolID     int64      `gorm:"column:school_id"`                        // 资源所属学校ID
	FileName     string     `gorm:"column:file_name"`                        // 资源上传名称
	OSSPath      string     `gorm:"column:oss_path"`                         // 资源存储路径
	OSSBucket    string     `gorm:"column:oss_bucket"`                       // 存储的bucket桶
	FileType     string     `gorm:"column:file_type"`                        // 文件类型
	FileByteSize int64      `gorm:"column:file_byte_size;type:bigint"`       // 文件大小(字节)
	FileHash     string     `gorm:"column:file_hash"`                        // 文件MD5/SHA哈希值
	FileScope    int        `gorm:"column:file_scope"`                       // 访问权限（0:公开, 1:私有, 2:特定班级）
	Metadata     string     `gorm:"column:metadata;type:jsonb"`              // 元数据（JSON格式，如文档页数、视频时长等）
	Status       int64      `gorm:"column:status"`                           // 资源状态（审核中/已通过/未通过等）
	CreateTime   *time.Time `gorm:"column:create_time"`                      // 资源上传时间
	UpdateTime   *time.Time `gorm:"column:update_time"`                      // 资源信息更新时间
	Deleted      bool       `gorm:"column:deleted"`                          // 删除标识（true:已删除，false:未删除）
}

// TableName 指定表名
func (Resource) TableName() string {
	return "tbl_resource"
}

// ResourceDAO 资源DAO
type ResourceDAO struct {
	db  *gorm.DB
	log *logger.ContextLogger
}

// NewResourceDAO 创建资源DAO
func NewResourceDAO(db *gorm.DB, l *logger.ContextLogger) *ResourceDAO {
	return &ResourceDAO{
		db:  db,
		log: l,
	}
}

// getDB 获取带有上下文的数据库连接
func (dao *ResourceDAO) getDB(ctx context.Context) *gorm.DB {
	return dao.db.WithContext(ctx)
}

// CreateResource 创建资源记录
func (dao *ResourceDAO) CreateResource(ctx context.Context, resource *Resource) error {
	// 设置创建和更新时间
	now := time.Now()
	resource.CreateTime = &now
	resource.UpdateTime = &now

	// 添加调试日志
	dao.log.Info(ctx, "执行数据库插入: 表=%s, 数据=%+v", resource.TableName(), resource)

	// 执行数据库操作
	err := dao.getDB(ctx).Create(resource).Error
	if err != nil {
		dao.log.Error(ctx, "数据库插入错误: %v", err)
		return err
	}

	dao.log.Info(ctx, "数据库插入成功, ID=%d", resource.ID)
	return nil
}

// GetResourceByObjectKey 根据对象键获取资源记录
func (dao *ResourceDAO) GetResourceByObjectKey(ctx context.Context, objectKey string) (*Resource, error) {
	var resource Resource

	dao.log.Info(ctx, "查询资源记录: 表=%s, oss_path包含=%s", resource.TableName(), objectKey)

	// 使用LIKE查询匹配包含objectKey的记录
	err := dao.getDB(ctx).Where("oss_path LIKE ? AND deleted = ?", "%"+objectKey+"%", false).First(&resource).Error
	if err != nil {
		dao.log.Error(ctx, "查询失败: %v", err)
		return nil, err
	}

	dao.log.Info(ctx, "查询成功: %+v", resource)
	return &resource, nil
}

// GetResourceByFileID 根据文件ID获取资源记录 (保留兼容旧代码)
func (dao *ResourceDAO) GetResourceByFileID(ctx context.Context, fileID string) (*Resource, error) {
	// 现在通过对象键来查询
	return dao.GetResourceByObjectKey(ctx, fileID)
}

// UpdateResource 更新资源记录
func (dao *ResourceDAO) UpdateResource(ctx context.Context, resource *Resource) error {
	// 更新更新时间
	now := time.Now()
	resource.UpdateTime = &now

	dao.log.Info(ctx, "执行数据库更新: 表=%s, ID=%d, 数据=%+v", resource.TableName(), resource.ID, resource)

	err := dao.getDB(ctx).Save(resource).Error
	if err != nil {
		dao.log.Error(ctx, "数据库更新错误: %v", err)
		return err
	}

	dao.log.Info(ctx, "数据库更新成功")
	return nil
}

// ListResourcesByUserID 根据用户ID列出资源记录
func (dao *ResourceDAO) ListResourcesByUserID(ctx context.Context, userID int64, limit, offset int) ([]*Resource, int64, error) {
	var resources []*Resource
	var total int64

	query := dao.getDB(ctx).Model(&Resource{})

	// 添加条件
	query = query.Where("user_id = ? AND deleted = ?", userID, false)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	if err := query.Order("create_time DESC").Limit(limit).Offset(offset).Find(&resources).Error; err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

// ListResourcesBySchoolID 根据学校ID列出资源记录
func (dao *ResourceDAO) ListResourcesBySchoolID(ctx context.Context, schoolID int64, limit, offset int) ([]*Resource, int64, error) {
	var resources []*Resource
	var total int64

	query := dao.getDB(ctx).Model(&Resource{})

	// 添加条件
	query = query.Where("school_id = ? AND deleted = ?", schoolID, false)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	if err := query.Order("create_time DESC").Limit(limit).Offset(offset).Find(&resources).Error; err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

// DeleteResource 删除资源记录
func (dao *ResourceDAO) DeleteResource(ctx context.Context, id int64) error {
	return dao.getDB(ctx).Delete(&Resource{}, id).Error
}

// GetResource 根据ID获取资源记录
func (dao *ResourceDAO) GetResource(ctx context.Context, id int64) (*Resource, error) {
	var resource Resource

	dao.log.Info(ctx, "通过ID查询资源记录: 表=%s, ID=%d", resource.TableName(), id)

	err := dao.getDB(ctx).Where("id = ? AND deleted = ?", id, false).First(&resource).Error
	if err != nil {
		dao.log.Error(ctx, "查询失败: %v", err)
		return nil, err
	}

	dao.log.Info(ctx, "查询成功: %+v", resource)
	return &resource, nil
}

// QueryResources 通用资源查询方法，支持多种查询条件
func (dao *ResourceDAO) QueryResources(ctx context.Context, params map[string]interface{}, limit, offset int) ([]*Resource, int64, error) {
	var resources []*Resource
	var total int64

	dao.log.Info(ctx, "执行资源查询: 参数=%+v, limit=%d, offset=%d", params, limit, offset)

	// 创建查询构建器
	query := dao.getDB(ctx).Model(&Resource{})

	// 添加 deleted 过滤条件（有索引）
	query = query.Where("deleted = ?", false)

	// 添加其他查询条件
	if params != nil {
		// 用户ID查询（有索引）
		if userID, ok := params["userId"].(int64); ok && userID > 0 {
			query = query.Where("user_id = ?", userID)
		}

		// 学校ID查询（有索引）
		if schoolID, ok := params["schoolId"].(int64); ok && schoolID > 0 {
			query = query.Where("school_id = ?", schoolID)
		}

		// 文件名模糊查询（有索引，但LIKE %xx% 可能不会使用索引）
		if fileName, ok := params["fileName"].(string); ok && fileName != "" {
			query = query.Where("file_name LIKE ?", "%"+fileName+"%")
		}

		// 文件类型查询
		if fileType, ok := params["fileType"].(string); ok && fileType != "" {
			query = query.Where("file_type = ?", strings.ToLower(fileType))
		}

		// 文件范围查询（有索引）
		if fileScope, ok := params["fileScope"].(int); ok {
			query = query.Where("file_scope = ?", fileScope)
		}

		// 状态查询（有索引）
		if status, ok := params["status"].(int64); ok {
			query = query.Where("status = ?", status)
		}

		// 开始时间查询（有索引）
		if startTime, ok := params["startTime"].(time.Time); ok {
			query = query.Where("create_time >= ?", startTime)
		}

		// 结束时间查询（有索引）
		if endTime, ok := params["endTime"].(time.Time); ok {
			query = query.Where("create_time <= ?", endTime)
		}

		// 排序方式（只允许对有索引的字段进行排序）
		if orderBy, ok := params["orderBy"].(string); ok && orderBy != "" {
			// 定义允许排序的字段（必须有索引支持）
			allowedOrderFields := map[string]bool{
				"user_id":     true,
				"school_id":   true,
				"file_name":   true,
				"file_scope":  true,
				"status":      true,
				"create_time": true,
			}

			if allowed := allowedOrderFields[orderBy]; allowed {
				orderDir := "DESC"
				if dir, ok := params["orderDir"].(string); ok && (dir == "ASC" || dir == "DESC") {
					orderDir = dir
				}
				query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))
			} else {
				// 如果不是允许的排序字段，使用默认排序
				query = query.Order("create_time DESC")
			}
		} else {
			// 默认按创建时间倒序（有索引）
			query = query.Order("create_time DESC")
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		dao.log.Error(ctx, "查询总数失败: %v", err)
		return nil, 0, fmt.Errorf("查询总数失败: %v", err)
	}

	// 查询分页数据
	if err := query.Limit(limit).Offset(offset).Find(&resources).Error; err != nil {
		dao.log.Error(ctx, "查询资源列表失败: %v", err)
		return nil, 0, fmt.Errorf("查询资源列表失败: %v", err)
	}

	dao.log.Info(ctx, "查询成功: 总数=%d, 当前页数据=%d", total, len(resources))
	return resources, total, nil
}

// SoftDeleteResource 软删除资源记录
func (dao *ResourceDAO) SoftDeleteResource(ctx context.Context, id int64, userID int64) error {
	// 更新更新时间
	now := time.Now()

	dao.log.Info(ctx, "执行软删除: 表=%s, ID=%d, 用户ID=%d", Resource{}.TableName(), id, userID)

	// 使用事务确保数据一致性
	err := dao.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// 先查询资源是否存在且属于该用户
		var resource Resource
		if err := tx.Where("id = ? AND user_id = ? AND deleted = ?", id, userID, false).First(&resource).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("资源不存在或无权限删除")
			}
			return err
		}

		// 执行软删除
		if err := tx.Model(&resource).Updates(map[string]interface{}{
			"deleted":     true,
			"update_time": now,
		}).Error; err != nil {
			return err
		}

		dao.log.Info(ctx, "软删除成功: ID=%d", id)
		return nil
	})

	if err != nil {
		dao.log.Error(ctx, "软删除失败: %v", err)
		return err
	}

	return nil
}
