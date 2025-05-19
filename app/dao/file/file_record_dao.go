package file

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// 文件记录实体
type FileRecord struct {
	ID           int64      `gorm:"column:id;primary_key;auto_increment"`  // 主键ID
	PhaseEnum    int64      `gorm:"column:phase_enum"`                     // 学段枚举
	SubjectEnum  int64      `gorm:"column:subject_enum"`                   // 学科枚举
	ObjectKey    string     `gorm:"column:object_key;type:varchar(255)"`   // 对象键名
	FileName     string     `gorm:"column:file_name;type:varchar(255)"`    // 文件名
	FileSize     int64      `gorm:"column:file_size"`                      // 文件大小
	FileURL      string     `gorm:"column:file_url;type:varchar(512)"`     // 文件URL
	UserID       string     `gorm:"column:user_id;type:varchar(64)"`       // 用户ID
	BusinessType string     `gorm:"column:business_type;type:varchar(64)"` // 业务类型
	UploadTime   *time.Time `gorm:"column:upload_time"`                    // 上传时间
	CallbackTime *time.Time `gorm:"column:callback_time"`                  // 回调时间
	Status       string     `gorm:"column:status;type:varchar(32)"`        // 状态
	Remark       string     `gorm:"column:remark;type:varchar(255)"`       // 备注
	CreatedAt    time.Time  `gorm:"column:created_at"`                     // 创建时间
	UpdatedAt    time.Time  `gorm:"column:updated_at"`                     // 更新时间
}

// TableName 指定表名
func (FileRecord) TableName() string {
	return "t_file_upload_record"
}

// FileRecordDAO 文件记录DAO
type FileRecordDAO struct {
	db *gorm.DB
}

// NewFileRecordDAO 创建文件记录DAO
func NewFileRecordDAO(db *gorm.DB) *FileRecordDAO {
	return &FileRecordDAO{
		db: db,
	}
}

// CreateUpload 创建上传记录
func (dao *FileRecordDAO) CreateUpload(ctx context.Context, record interface{}) error {
	// 利用传入的interface{}创建文件记录
	// 这里假设record是一个包含必要字段的匿名结构体
	data, ok := record.(struct {
		PhaseEnum    int64
		SubjectEnum  int64
		ObjectKey    string
		FileName     string
		FileSize     int64
		FileURL      string
		UserID       string
		BusinessType string
		UploadTime   *time.Time
		CallbackTime *time.Time
		Status       string
		Remark       string
	})

	if !ok {
		// 尝试其他可能的结构类型
		// 这只是一个例子，实际上可能需要更复杂的类型断言或反射
		return dao.db.WithContext(ctx).Create(record).Error
	}

	fileRecord := FileRecord{
		PhaseEnum:    data.PhaseEnum,
		SubjectEnum:  data.SubjectEnum,
		ObjectKey:    data.ObjectKey,
		FileName:     data.FileName,
		FileSize:     data.FileSize,
		FileURL:      data.FileURL,
		UserID:       data.UserID,
		BusinessType: data.BusinessType,
		UploadTime:   data.UploadTime,
		CallbackTime: data.CallbackTime,
		Status:       data.Status,
		Remark:       data.Remark,
	}

	return dao.db.WithContext(ctx).Create(&fileRecord).Error
}

// GetUploadByObjectKey 根据对象键名获取上传记录
func (dao *FileRecordDAO) GetUploadByObjectKey(ctx context.Context, objectKey string) (interface{}, error) {
	var record FileRecord
	err := dao.db.WithContext(ctx).Where("object_key = ?", objectKey).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateUploadRecord 更新上传记录
func (dao *FileRecordDAO) UpdateUploadRecord(ctx context.Context, record interface{}) error {
	// 如果record是FileRecord类型，直接更新
	if fileRecord, ok := record.(*FileRecord); ok {
		return dao.db.WithContext(ctx).Save(fileRecord).Error
	}

	// 否则，尝试使用传入的记录进行更新
	return dao.db.WithContext(ctx).Save(record).Error
}

// ListUploadRecords 列出上传记录
func (dao *FileRecordDAO) ListUploadRecords(ctx context.Context, phaseEnum, subjectEnum int64, userId, businessType, status string, limit, offset int) (interface{}, int64, error) {
	var records []FileRecord
	var total int64

	query := dao.db.WithContext(ctx).Model(&FileRecord{})

	// 添加查询条件
	if phaseEnum > 0 {
		query = query.Where("phase_enum = ?", phaseEnum)
	}
	if subjectEnum > 0 {
		query = query.Where("subject_enum = ?", subjectEnum)
	}
	if userId != "" {
		query = query.Where("user_id = ?", userId)
	}
	if businessType != "" {
		query = query.Where("business_type = ?", businessType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总记录数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
