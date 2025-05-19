package consts

import "time"

const (
	// DefaultExpireSeconds 默认过期时间（30分钟）
	DefaultExpireSeconds = 1800

	// DefaultAPITimeout API请求超时时间（10秒）
	DefaultAPITimeout = 10 * time.Second

	// TeacherDefaultAPITimeout API请求超时时间（5秒）
	TeacherDefaultAPITimeout = 5 * time.Second

	// TeacherDefaultRetryInterval API 请求重试间隔（3 秒）
	TeacherDefaultRetryInterval = 3 * time.Second

	// DefaultMaxRetries 默认最大重试次数
	DefaultMaxRetries = 3

	// DefaultBufferSize 默认缓冲区大小
	DefaultBufferSize = 4096
)

// API相关默认值
const (
	// DefaultAPIVersion API版本
	DefaultAPIVersion = "v1"

	// DefaultAPIPrefix API前缀
	DefaultAPIPrefix = "/api/" + DefaultAPIVersion
)

// 文件相关默认值
const (
	// DefaultMaxFileSize 默认最大文件大小（10MB）
	DefaultMaxFileSize = 10 << 20

	// DefaultAllowedFileTypes 默认允许的文件类型
	DefaultAllowedFileTypes = "jpg,jpeg,png,gif,doc,docx,xls,xlsx,pdf"
)

// 数据库相关默认值
const (
	// DefaultDBMaxIdleConns 默认最大空闲连接数
	DefaultDBMaxIdleConns = 10

	// DefaultDBMaxOpenConns 默认最大打开连接数
	DefaultDBMaxOpenConns = 100

	// DefaultDBConnMaxLifetime 默认连接最大生命周期（1小时）
	DefaultDBConnMaxLifetime = time.Hour
)
