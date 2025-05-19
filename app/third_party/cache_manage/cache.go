package cache_manage

import (
	"errors"
	"fmt"
	"strings"
)

type keyItem struct {
	Key    string
	Params string
	Type   string
}

// 全局key配置
var allKeys = map[string]keyItem{
	// 定义缓存名称
	"userInfo": {
		Key:    "user_info_{task_id}", // 定义缓存key,规则：要包含一级key的名称，比如userInfo->user_info
		Params: "user_id",             // 定义参数列表，逗号分割
		Type:   "hash",
	},
}

// GenerateCacheKey 生成全局key
func GenerateCacheKey(keyName string, keyType string, params map[string]any) (cacheKey string, err error) {

	keyInfo, ok := allKeys[keyName]
	if !ok {
		return "", errors.New("keyName 不存在")
	}

	if keyType == "" || keyType != keyInfo.Type {
		return "", errors.New("keyType 错误")
	}

	cacheKey = keyInfo.Key

	keyParams := strings.Split(keyInfo.Params, ",")
	if nil == keyParams {
		return
	}

	for k, v := range params {
		cacheKey = strings.ReplaceAll(cacheKey, "{"+k+"}", fmt.Sprintf("%v", v))
	}

	return cacheKey, nil
}
