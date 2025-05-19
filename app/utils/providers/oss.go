package providers

import (
	"gil_teacher/app/conf"
	"gil_teacher/app/utils/oss"

	"github.com/google/wire"
)

// OSSClientProvider 提供OSS客户端
var OSSClientProvider = wire.NewSet(
	NewOSSClient,
)

// NewOSSClient 创建OSS客户端
func NewOSSClient(config *conf.Config) *oss.OSSClient {
	if config.OSS == nil {
		// 如果配置为空，使用默认配置
		return oss.NewOSSClient(
			"",    // AccessKeyID
			"",    // AccessKeySecret
			"",    // Region
			"",    // BucketName
			"",    // Endpoint
			false, // Internal
			true,  // Secure
			"",    // BasePath
		)
	}

	ossConfig := config.OSS
	return oss.NewOSSClient(
		ossConfig.AccessKeyID,
		ossConfig.AccessKeySecret,
		ossConfig.Region,
		ossConfig.BucketName,
		ossConfig.Endpoint,
		ossConfig.Internal,
		ossConfig.Secure,
		ossConfig.BasePath,
	)
}
