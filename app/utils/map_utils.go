package utils

import (
	"context"
	clogger "gil_teacher/app/core/logger"
)

// GetStringFromMap 从 map 中获取字符串值
func GetStringFromMap(m map[string]interface{}, key string) (string, bool) {
	value, exists := m[key]
	if !exists {
		return "", false
	}

	strValue, ok := value.(string)
	if !ok {
		return "", false
	}

	return strValue, true
}

// GetFloat64FromMap 从 map 中获取 float64 类型的值
func GetFloat64FromMap(m map[string]interface{}, key string) (float64, bool) {
	value, exists := m[key]
	if !exists {
		return 0, false
	}

	floatValue, ok := value.(float64)
	if !ok {
		return 0, false
	}

	return floatValue, true
}

// GetInt64FromMap 从 map 中获取 int64 类型的值
func GetInt64FromMap(m map[string]interface{}, key string) (int64, bool) {
	value, exists := m[key]
	if !exists {
		return 0, false
	}

	floatValue, ok := value.(float64)
	if !ok {
		return 0, false
	}

	return int64(floatValue), true
}

// GetUint64FromMap 从 map 中获取 uint64 类型的值
func GetUint64FromMap(m map[string]interface{}, key string) (uint64, bool) {
	value, exists := m[key]
	if !exists {
		return 0, false
	}

	floatValue, ok := value.(float64)
	if !ok {
		return 0, false
	}

	return uint64(floatValue), true
}

// GetBoolFromMap 从 map 中获取 bool 类型的值
func GetBoolFromMap(m map[string]interface{}, key string) (bool, bool) {
	value, exists := m[key]
	if !exists {
		return false, false
	}

	boolValue, ok := value.(bool)
	if !ok {
		return false, false
	}

	return boolValue, true
}

// GetMapStringKey 从 map 中安全地获取 string 类型的值
// 如果 key 不存在或类型不匹配，返回空字符串
func GetMapStringKey(m map[string]interface{}, key string) string {
	val, _ := GetStringFromMap(m, key)
	return val
}

// GetMapInt64Key 从 map 中安全地获取 int64 类型的值
// 如果 key 不存在或类型不匹配，返回 0
func GetMapInt64Key(m map[string]interface{}, key string) int64 {
	val, _ := GetInt64FromMap(m, key)
	return val
}

// GetMapUint64Key 从 map 中安全地获取 uint64 类型的值
// 如果 key 不存在或类型不匹配，返回 0
func GetMapUint64Key(m map[string]interface{}, key string) uint64 {
	val, _ := GetUint64FromMap(m, key)
	return val
}

// GetMapFloat64Key 从 map 中安全地获取 float64 类型的值
// 如果 key 不存在或类型不匹配，返回 0
func GetMapFloat64Key(m map[string]interface{}, key string) float64 {
	val, _ := GetFloat64FromMap(m, key)
	return val
}

// GetMapBoolKey 从 map 中安全地获取 bool 类型的值
// 如果 key 不存在或类型不匹配，返回 false
func GetMapBoolKey(m map[string]interface{}, key string) bool {
	val, _ := GetBoolFromMap(m, key)
	return val
}

// FieldExtractor 字段提取器
type FieldExtractor struct {
	ctx          context.Context
	logger       *clogger.ContextLogger
	failedFields []string
	rawValues    []interface{}
}

// NewFieldExtractor 创建字段提取器
func NewFieldExtractor(ctx context.Context, logger *clogger.ContextLogger) *FieldExtractor {
	return &FieldExtractor{
		ctx:    ctx,
		logger: logger,
	}
}

// ExtractString 提取字符串字段
func (f *FieldExtractor) ExtractString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
		f.failedFields = append(f.failedFields, key)
		f.rawValues = append(f.rawValues, val)
	}
	return ""
}

// ExtractUint64 提取uint64字段
func (f *FieldExtractor) ExtractUint64(m map[string]interface{}, key string) uint64 {
	if val, ok := m[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return uint64(floatVal)
		}
		f.failedFields = append(f.failedFields, key)
		f.rawValues = append(f.rawValues, val)
	}
	return 0
}

// ExtractInt64 提取int64字段
func (f *FieldExtractor) ExtractInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return int64(floatVal)
		}
		f.failedFields = append(f.failedFields, key)
		f.rawValues = append(f.rawValues, val)
	}
	return 0
}

// LogFailures 记录失败的字段
func (f *FieldExtractor) LogFailures() {
	if len(f.failedFields) > 0 {
		f.logger.Warn(f.ctx, "字段类型转换失败 - 字段: %v, 原始值: %v", f.failedFields, f.rawValues)
	}
}

// GetMapValueString 从 map 中提取字段，没有则返回默认值
func GetMapValueString(m map[string]any, key string, defaultValue string) string {
	if value, ok := m[key]; ok {
		if strVal, ok := value.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// GetMapValueI64 从 map 中提取字段，没有则返回默认值
func GetMapValueI64(m map[string]any, key string, defaultValue int64) int64 {
	if value, ok := m[key]; ok {
		if floatVal, ok := value.(float64); ok {
			return int64(floatVal)
		}
	}
	return defaultValue
}

// GetMapValueF64 从 map 中提取字段，没有则返回默认值
func GetMapValueF64(m map[string]any, key string, defaultValue float64) float64 {
	if value, ok := m[key]; ok {
		if floatVal, ok := value.(float64); ok {
			return floatVal
		}
	}
	return defaultValue
}

// GetMapValueBool 从 map 中提取字段，没有则返回默认值
func GetMapValueBool(m map[string]any, key string, defaultValue bool) bool {
	if value, ok := m[key]; ok {
		if boolVal, ok := value.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// GetMapValueUint64 从 map 中提取字段，没有则返回默认值
func GetMapValueUint64(m map[string]any, key string, defaultValue uint64) uint64 {
	if value, ok := m[key]; ok {
		if floatVal, ok := value.(float64); ok {
			return uint64(floatVal)
		}
	}
	return defaultValue
}

// GetMapIntKey 从 map 中获取 int 类型的值，提供默认值
// 如果 key 不存在或类型不匹配，返回默认值
func GetMapIntKey(m map[string]interface{}, key string, defaultValue int) int {
	if value, ok := m[key]; ok {
		if floatVal, ok := value.(float64); ok {
			return int(floatVal)
		}
	}
	return defaultValue
}
