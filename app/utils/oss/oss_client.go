package oss

import (
	"fmt"
	"gil_teacher/app/consts"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSClient 阿里云OSS客户端
type OSSClient struct {
	accessKeyID     string      // AccessKeyID
	accessKeySecret string      // AccessKeySecret
	region          string      // 区域
	bucketName      string      // 存储桶名称
	endpoint        string      // 访问域名
	internal        bool        // 是否使用内网
	secure          bool        // 是否使用HTTPS
	basePath        string      // 基础路径
	client          *oss.Client // 阿里云OSS客户端
	bucket          *oss.Bucket // 阿里云OSS存储桶
}

// NewOSSClient 创建OSS客户端
func NewOSSClient(
	accessKeyID, accessKeySecret, region, bucketName, endpoint string,
	internal, secure bool, basePath string,
) *OSSClient {
	ossClient := &OSSClient{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		region:          region,
		bucketName:      bucketName,
		endpoint:        endpoint,
		internal:        internal,
		secure:          secure,
		basePath:        basePath,
	}

	// 初始化阿里云OSS客户端
	ossClient.initClient()

	return ossClient
}

// initClient 初始化阿里云OSS客户端
func (oc *OSSClient) initClient() error {
	// 构造访问域名
	domain := oc.endpoint
	if oc.internal {
		// 如果域名中不包含"-internal"，则添加
		if !strings.Contains(domain, "-internal") {
			domain = strings.Replace(domain, ".aliyuncs.com", "-internal.aliyuncs.com", 1)
		}
	}

	// 构造访问协议
	protocol := "http"
	if oc.secure {
		protocol = "https"
	}

	// 构造完整的访问域名
	endpoint := fmt.Sprintf("%s://%s", protocol, domain)

	// 创建阿里云OSS客户端
	client, err := oss.New(endpoint, oc.accessKeyID, oc.accessKeySecret)
	if err != nil {
		fmt.Printf("创建阿里云OSS客户端失败: %v\n", err)
		return fmt.Errorf("创建阿里云OSS客户端失败: %w", err)
	}

	// 获取存储桶
	bucket, err := client.Bucket(oc.bucketName)
	if err != nil {
		fmt.Printf("获取阿里云OSS存储桶失败: %v\n", err)
		return fmt.Errorf("获取阿里云OSS存储桶失败: %w", err)
	}

	oc.client = client
	oc.bucket = bucket

	fmt.Printf("OSS客户端初始化成功: endpoint=%s, bucket=%s\n", endpoint, oc.bucketName)

	return nil
}

// GetAccessKeyID 获取AccessKeyID
func (oc *OSSClient) GetAccessKeyID() string {
	return oc.accessKeyID
}

// GetAccessKeySecret 获取AccessKeySecret
func (oc *OSSClient) GetAccessKeySecret() string {
	return oc.accessKeySecret
}

// GetRegion 获取区域
func (oc *OSSClient) GetRegion() string {
	return oc.region
}

// GetBucketName 获取存储桶名称
func (oc *OSSClient) GetBucketName() string {
	return oc.bucketName
}

// GetBasePath 获取基础路径
func (oc *OSSClient) GetBasePath() string {
	return oc.basePath
}

// GetObjectURL 获取对象URL
func (oc *OSSClient) GetObjectURL(objectKey string) string {
	protocol := "http"
	if oc.secure {
		protocol = "https"
	}

	domain := oc.endpoint
	if oc.internal {
		// 如果域名中不包含"-internal"，则添加
		if !strings.Contains(domain, "-internal") {
			domain = strings.Replace(domain, ".aliyuncs.com", "-internal.aliyuncs.com", 1)
		}
	}

	// 如果basePath不为空且不以"/"结尾，则添加"/"
	basePath := oc.basePath
	if basePath != "" && !strings.HasSuffix(basePath, "/") {
		basePath = basePath + "/"
	}

	// 如果objectKey以"/"开头，则去掉
	if strings.HasPrefix(objectKey, "/") {
		objectKey = objectKey[1:]
	}

	return fmt.Sprintf("%s://%s.%s/%s%s", protocol, oc.bucketName, domain, basePath, objectKey)
}

// GeneratePresignedURL 生成预签名URL
func (oc *OSSClient) GeneratePresignedURL(objectKey string, method string, expires interface{}, headers map[string]string) (string, error) {
	// 如果客户端未初始化，则初始化
	if oc.client == nil || oc.bucket == nil {
		if err := oc.initClient(); err != nil {
			return "", fmt.Errorf("初始化OSS客户端失败: %w", err)
		}
	}

	// 处理objectKey，确保正确的路径
	// 如果basePath不为空，则添加basePath前缀
	if oc.basePath != "" {
		basePath := oc.basePath
		// 确保basePath以"/"结尾
		if !strings.HasSuffix(basePath, "/") {
			basePath = basePath + "/"
		}
		// 确保objectKey不以"/"开头
		if strings.HasPrefix(objectKey, "/") {
			objectKey = objectKey[1:]
		}
		objectKey = basePath + objectKey
	}

	// 处理过期时间
	var expiration time.Duration
	switch v := expires.(type) {
	case time.Duration:
		expiration = v
	case int:
		expiration = time.Duration(v) * time.Second
	case int64:
		expiration = time.Duration(v) * time.Second
	default:
		expiration = time.Duration(consts.ExpireSeconds30Min) * time.Second // 默认30分钟
	}

	// 转换HTTP方法
	var ossMethod oss.HTTPMethod
	switch strings.ToUpper(method) {
	case "GET", "HEAD":
		ossMethod = oss.HTTPGet
	case "PUT":
		ossMethod = oss.HTTPPut
	case "POST":
		ossMethod = oss.HTTPPost
	case "DELETE":
		ossMethod = oss.HTTPDelete
	default:
		return "", fmt.Errorf("不支持的HTTP方法: %s", method)
	}

	// 构造签名选项
	options := []oss.Option{}
	// 我们不使用额外的headers选项，这里只是保留代码框架
	// 实际使用中可以添加其他选项
	_ = headers

	fmt.Printf("正在生成预签名URL: objectKey=%s, method=%s, expires=%v\n",
		objectKey, method, expiration)

	// 生成预签名URL
	signedURL, err := oc.bucket.SignURL(objectKey, ossMethod, int64(expiration.Seconds()), options...)
	if err != nil {
		fmt.Printf("生成预签名URL失败: %v\n", err)
		return "", fmt.Errorf("生成预签名URL失败: %w", err)
	}

	fmt.Printf("成功生成预签名URL: %s\n", signedURL)

	return signedURL, nil
}

// GetPresignedPutURL 获取预签名PUT URL
func (oc *OSSClient) GetPresignedPutURL(objectKey string, expireTime time.Time) (string, error) {
	if oc.bucket == nil {
		return "", fmt.Errorf("OSS客户端未初始化")
	}

	// 设置过期时间（从现在开始的秒数）
	expireSeconds := int64(expireTime.Sub(time.Now()).Seconds())
	if expireSeconds <= 0 {
		expireSeconds = consts.ExpireSeconds30Min // 默认30分钟
	}

	// 生成签名URL
	signedURL, err := oc.bucket.SignURL(objectKey, oss.HTTPPut, expireSeconds)
	if err != nil {
		return "", fmt.Errorf("生成预签名URL失败: %w", err)
	}

	return signedURL, nil
}
