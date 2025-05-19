package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateUniqueFileName 生成唯一的文件名
// 使用 SHA256 对用户ID、时间戳和随机数进行哈希，生成唯一的文件名
// 参数:
//   - userID: 用户ID
//   - timestamp: 时间戳
//   - randomStr: 随机字符串
//
// 返回:
//   - string: 生成的文件名
func GenerateUniqueFileName(userID int64, timestamp time.Time, randomStr string) string {
	// 构建输入字符串
	input := fmt.Sprintf("%d_%s_%s", userID, timestamp.Format("20060102150405"), randomStr)

	// 计算 SHA256 哈希
	hash := sha256.New()
	hash.Write([]byte(input))

	// 获取十六进制编码的哈希值
	hashStr := hex.EncodeToString(hash.Sum(nil))

	// 返回前32位作为文件名
	return hashStr[:32]
}

// GenerateObjectKey 生成OSS对象键
// 参数:
//   - basePath: 基础路径
//   - fileName: 文件名
//   - fileExt: 文件扩展名
//
// 返回:
//   - string: 完整的对象键
func GenerateObjectKey(basePath, fileName, fileExt string) string {
	return fmt.Sprintf("%s/%s%s", basePath, fileName, fileExt)
}
