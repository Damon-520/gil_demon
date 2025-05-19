package providers

import (
	"gil_teacher/app/core/logger"
	"gil_teacher/app/dao/file"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// FileRecordDAOProvider 提供文件记录DAO
var FileRecordDAOProvider = wire.NewSet(
	NewFileRecordDAO,
)

// NewFileRecordDAO 创建文件记录DAO
func NewFileRecordDAO(db *gorm.DB) *file.FileRecordDAO {
	return file.NewFileRecordDAO(db)
}

// ResourceDAOProvider 提供资源DAO
var ResourceDAOProvider = wire.NewSet(
	NewResourceDAO,
)

// NewResourceDAO 创建资源DAO
func NewResourceDAO(db *gorm.DB, logger *logger.ContextLogger) *file.ResourceDAO {
	return file.NewResourceDAO(db, logger)
}
