package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gil_teacher/app/conf"
	"gil_teacher/app/consts"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/core/redisx"
	"gil_teacher/app/utils"

	"github.com/go-redis/redis/v8"
)

type ApiRdbClient redis.Client

var prefix string

func NewApiRedisClient(conf *conf.Conf, logger *clogger.ContextLogger) *ApiRdbClient {
	client := redisx.NewRedis(conf.Data.Redis, logger)
	prefix = fmt.Sprintf("%s:%s", conf.App.Name, conf.Config.Env)
	return (*ApiRdbClient)(client)
}

// 设置公共前缀, app:env:key
func (c *ApiRdbClient) realKey(key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}

// expire 设置过期时间
func (c *ApiRdbClient) expire(ttlSec int64) time.Duration {
	if ttlSec <= 0 {
		ttlSec = consts.DefaultExpire
	}
	return time.Duration(ttlSec) * time.Second
}

// checkParams 检查参数是否合法
// key: redis中的key
// value: 值
// dest: 目标指针
// 返回值: 错误信息
func (c *ApiRdbClient) checkParams(key string, value any, dest any) error {
	if key == "" {
		return fmt.Errorf("key不能为空")
	}

	if value != nil {
		switch v := value.(type) {
		case string:
			if v == "" {
				return fmt.Errorf("value不能为空")
			}
		case []*redis.Z:
			if len(v) == 0 {
				return fmt.Errorf("members不能为空")
			}
		case *redis.ZRangeBy:
			if v == nil {
				return fmt.Errorf("opt不能为空")
			}
		}
	}

	if dest != nil {
		destValue := reflect.ValueOf(dest)
		if destValue.Kind() != reflect.Ptr {
			return fmt.Errorf("dest必须是指针类型")
		}
		if destValue.IsNil() {
			return fmt.Errorf("dest不能为空")
		}
	}

	return nil
}

// handleRedisError 统一处理Redis错误
// err: Redis错误
// operation: 操作名称
// 返回值: 处理后的错误
func (c *ApiRdbClient) handleRedisError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// 特别注意：redis.Nil表示键不存在，这不是一个真正的错误
	// 但调用方需要额外处理这种情况，正确判断键是否存在
	if err == redis.Nil {
		// 返回nil表示没有错误，但这不代表键存在
		// 调用此方法的函数必须在调用前特殊处理redis.Nil情况
		return nil
	}
	return fmt.Errorf("%s失败: %w", operation, err)
}

// Set 支持将任意数据写入 cache
// key: redis中的key
// data: 任意数据，包括基本类型和结构体
// ttlSec: 过期时间，秒
// 返回值: 错误信息
func (c *ApiRdbClient) Set(ctx context.Context, key string, data any, ttlSec int64) error {
	if err := c.checkParams(key, data, nil); err != nil {
		return err
	}

	// 使用 json.Marshal 处理所有类型
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("数据序列化失败: %w", err)
	}

	// 写入redis并设置过期时间
	result := (*redis.Client)(c).Set(ctx, c.realKey(key), bytes, c.expire(ttlSec))
	return c.handleRedisError(result.Err(), "写入redis")
}

// Get 从redis获取数据并解析到目标结构体
// key: redis中的key
// dest: 目标结构体指针
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) Get(ctx context.Context, key string, dest any) (bool, error) {
	if err := c.checkParams(key, nil, dest); err != nil {
		return false, err
	}

	// 从redis获取数据
	result := (*redis.Client)(c).Get(ctx, c.realKey(key))

	// 特殊处理redis.Nil错误，表示键不存在
	if result.Err() == redis.Nil {
		return false, nil
	}

	if err := c.handleRedisError(result.Err(), "获取redis数据"); err != nil {
		return false, err
	}

	// 解析数据到目标结构体
	bytes, err := result.Bytes()
	if err != nil {
		// 如果获取字节失败但不是redis.Nil，说明有其他错误
		// 此时键可能存在但值有问题，返回true表示键存在，同时返回错误
		return true, fmt.Errorf("获取数据字节失败: %w", err)
	}

	if err := json.Unmarshal(bytes, dest); err != nil {
		return true, fmt.Errorf("数据反序列化失败: %w", err)
	}

	return true, nil
}

// Incr 将key的值加1
// key: redis中的key
// ttlSec: 过期时间，秒
// 返回值: (增加后的值, 错误信息)
func (c *ApiRdbClient) Incr(ctx context.Context, key string, ttlSec int64) (int64, error) {
	if err := c.checkParams(key, nil, nil); err != nil {
		return 0, err
	}

	realKey := c.realKey(key)
	result := (*redis.Client)(c).Incr(ctx, realKey)
	if err := c.handleRedisError(result.Err(), "增加redis值"); err != nil {
		return 0, err
	}

	// 设置过期时间
	(*redis.Client)(c).Expire(ctx, realKey, c.expire(ttlSec))

	return result.Val(), nil
}

// HSet 将任意类型的数据写入redis hash
// key: redis中的key
// data: 任意类型的数据，包括基本类型、map、结构体等
// ttlSec: 过期时间，秒
// 返回值: 错误信息
func (c *ApiRdbClient) HSet(ctx context.Context, key string, data any, ttlSec int64) error {
	if err := c.checkParams(key, data, nil); err != nil {
		return err
	}

	redisData := make(map[string]string)
	if err := handleStruct(data, redisData, true); err != nil {
		return err
	}

	realKey := c.realKey(key)
	result := (*redis.Client)(c).HSet(ctx, realKey, redisData)
	if err := c.handleRedisError(result.Err(), "写入redis hash"); err != nil {
		return err
	}

	(*redis.Client)(c).Expire(ctx, realKey, c.expire(ttlSec))

	return nil
}

// HSetField 将任意类型的数据写入redis hash的指定字段
// key: redis中的key
// field: 字段名
// value: 值
// ttlSec: 过期时间，秒
// 返回值: 错误信息
func (c *ApiRdbClient) HSetField(ctx context.Context, key string, field string, value any, ttlSec int64) error {
	if err := c.checkParams(key, field, value); err != nil {
		return err
	}

	realKey := c.realKey(key)
	result := (*redis.Client)(c).HSet(ctx, realKey, field, value)
	if err := c.handleRedisError(result.Err(), "写入redis hash字段"); err != nil {
		return err
	}

	(*redis.Client)(c).Expire(ctx, realKey, c.expire(ttlSec))

	return nil
}

// HGetAll 从redis获取hash数据并解析到目标结构体
// key: redis中的key
// dest: 目标结构体指针或map，支持 map[K]V 其中 K 可以是任意基本类型
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) HGetAll(ctx context.Context, key string, dest any) (bool, error) {
	if err := c.checkParams(key, nil, dest); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).HGetAll(ctx, c.realKey(key))
	if err := c.handleRedisError(result.Err(), "获取redis hash数据"); err != nil {
		return false, err
	}

	if len(result.Val()) == 0 {
		return false, nil
	}

	if err := handleStruct(dest, result.Val(), false); err != nil {
		return false, err
	}

	return true, nil
}

// HGetField 从redis获取hashmap中的指定字段
// key: redis中的key
// field: 字段名
// dest: 目标结构体指针或基本类型指针
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) HGetField(ctx context.Context, key string, field string, dest any) (bool, error) {
	if err := c.checkParams(key, field, dest); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).HGet(ctx, c.realKey(key), field)

	// 特殊处理redis.Nil错误，表示键或字段不存在
	if result.Err() == redis.Nil {
		return false, nil
	}

	if err := c.handleRedisError(result.Err(), "获取redis hash字段"); err != nil {
		return false, err
	}

	destType := reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		return false, fmt.Errorf("目标必须是指针类型")
	}
	destElemType := destType.Elem()

	value, err := stringToValue(result.Val(), destElemType)
	if err != nil {
		return true, fmt.Errorf("转换value失败: %w", err)
	}
	reflect.ValueOf(dest).Elem().Set(value)

	return true, nil
}

// 判断 key 是否存在
func (c *ApiRdbClient) KeyExists(ctx context.Context, key string) (bool, error) {
	if err := c.checkParams(key, nil, nil); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).Exists(ctx, c.realKey(key))
	if err := c.handleRedisError(result.Err(), "检查key是否存在"); err != nil {
		return false, err
	}

	return result.Val() > 0, nil
}

// ZAdd 向有序集合添加一个成员
// key: redis中的key
// score: 成员的分数
// member: 成员的值
// ttlSec: 过期时间，秒
// 返回值: 错误信息
func (c *ApiRdbClient) ZAdd(ctx context.Context, key string, score float64, member string, ttlSec int64) error {
	if err := c.checkParams(key, member, nil); err != nil {
		return err
	}

	realKey := c.realKey(key)
	// 创建有序集合成员
	z := &redis.Z{
		Score:  score,
		Member: member,
	}

	// 添加成员到有序集合
	result := (*redis.Client)(c).ZAdd(ctx, realKey, z)
	if result.Err() != nil {
		return fmt.Errorf("添加有序集合成员失败: %w", result.Err())
	}

	// 设置过期时间
	(*redis.Client)(c).Expire(ctx, realKey, c.expire(ttlSec))

	return nil
}

// ZAddBatch 批量向有序集合添加成员
// key: redis中的key
// members: 成员列表，每个成员包含分数和值
// ttlSec: 过期时间，秒
// 返回值: 错误信息
func (c *ApiRdbClient) ZAddBatch(ctx context.Context, key string, members []*redis.Z, ttlSec int64) error {
	if err := c.checkParams(key, members, nil); err != nil {
		return err
	}

	realKey := c.realKey(key)
	// 批量添加成员到有序集合
	result := (*redis.Client)(c).ZAdd(ctx, realKey, members...)
	if result.Err() != nil {
		return fmt.Errorf("批量添加有序集合成员失败: %w", result.Err())
	}

	// 设置过期时间
	(*redis.Client)(c).Expire(ctx, realKey, c.expire(ttlSec))

	return nil
}

// ZRange 返回有序集合中指定区间内的成员
// key: redis中的key
// start: 开始位置
// stop: 结束位置
// dest: 目标切片指针，用于存储结果
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) ZRange(ctx context.Context, key string, start, stop int64, dest *[]string) (bool, error) {
	if err := c.checkParams(key, nil, dest); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).ZRange(ctx, c.realKey(key), start, stop)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("获取有序集合成员失败: %w", result.Err())
	}

	*dest = result.Val()
	return true, nil
}

// ZCount 返回有序集合中分数在指定区间内的成员数量
// key: redis中的key
// min: 最小分数
// max: 最大分数
// dest: 目标指针，用于存储结果
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) ZCount(ctx context.Context, key string, min, max float64, dest *int64) (bool, error) {
	if err := c.checkParams(key, nil, dest); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).ZCount(ctx, c.realKey(key), utils.F64ToStr(min), utils.F64ToStr(max))
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("统计有序集合成员数量失败: %w", result.Err())
	}

	*dest = result.Val()
	return true, nil
}

// ZRangeByScore 返回有序集合中分数在指定区间内的成员
// key: redis中的key
// min: 最小分数, 0 表示最小值
// max: 最大分数, +inf 表示最大值
// dest: 目标切片指针，用于存储结果
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) ZRangeByScore(ctx context.Context, key string, min, max float64, dest *[]string) (bool, error) {
	if err := c.checkParams(key, nil, dest); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).ZRangeByScore(ctx, c.realKey(key), &redis.ZRangeBy{
		Min: utils.F64ToString(min, -1, "0"),
		Max: utils.F64ToString(max, -1, "+inf"),
	})
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("获取有序集合成员失败: %w", result.Err())
	}

	*dest = result.Val()
	return true, nil
}

// Keys 查找所有符合给定模式的键
// 警告：此方法在生产环境中应谨慎使用，因为它会扫描所有键，可能导致性能问题
// pattern: 匹配模式
// dest: 目标切片指针，用于存储结果
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) Keys(ctx context.Context, pattern string, dest *[]string) (bool, error) {
	if err := c.checkParams(pattern, nil, dest); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).Keys(ctx, c.realKey(pattern))
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("查找键失败: %w", result.Err())
	}

	*dest = result.Val()
	return true, nil
}

// SMembers 返回集合中的所有成员
// key: redis中的key
// dest: 目标切片指针，用于存储结果
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) SMembers(ctx context.Context, key string, dest *[]string) (bool, error) {
	if err := c.checkParams(key, nil, dest); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).SMembers(ctx, c.realKey(key))
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("获取集合成员失败: %w", result.Err())
	}

	*dest = result.Val()
	return true, nil
}

// SAdd 向集合添加一个成员
// key: redis中的key
// member: 要添加的成员
// ttlSec: 过期时间，秒
// 返回值: 错误信息
func (c *ApiRdbClient) SAdd(ctx context.Context, key string, member string, ttlSec int64) error {
	if err := c.checkParams(key, member, nil); err != nil {
		return err
	}

	realKey := c.realKey(key)
	// 添加成员到集合
	result := (*redis.Client)(c).SAdd(ctx, realKey, member)
	if result.Err() != nil {
		return fmt.Errorf("添加集合成员失败: %w", result.Err())
	}

	// 设置过期时间
	(*redis.Client)(c).Expire(ctx, realKey, c.expire(ttlSec))

	return nil
}

// SIsMember 判断成员是否在集合中
// key: redis中的key
// member: 要判断的成员
// 返回值: (成员是否存在, 错误信息)
func (c *ApiRdbClient) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	if err := c.checkParams(key, member, nil); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).SIsMember(ctx, c.realKey(key), member)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("判断集合成员失败: %w", result.Err())
	}

	return result.Val(), nil
}

// ZRem 从有序集合中移除一个成员
// key: redis中的key
// member: 要移除的成员
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) ZRem(ctx context.Context, key string, member string) (bool, error) {
	if err := c.checkParams(key, member, nil); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).ZRem(ctx, c.realKey(key), member)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("移除有序集合成员失败: %w", result.Err())
	}

	return true, nil
}

// ZRemRangeByScore 移除有序集合中分数在指定区间内的成员
// key: redis中的key
// min: 最小分数
// max: 最大分数
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) ZRemRangeByScore(ctx context.Context, key string, min, max string) (bool, error) {
	if err := c.checkParams(key, min, max); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).ZRemRangeByScore(ctx, c.realKey(key), min, max)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("移除有序集合成员失败: %w", result.Err())
	}

	return true, nil
}

// ZCard 返回有序集合的成员数量
// key: redis中的key
// dest: 目标指针，用于存储结果
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) ZCard(ctx context.Context, key string, dest *int64) (bool, error) {
	if err := c.checkParams(key, nil, dest); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).ZCard(ctx, c.realKey(key))
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("获取有序集合成员数量失败: %w", result.Err())
	}

	*dest = result.Val()
	return true, nil
}

// ZScore 返回有序集合中成员的分数
// key: redis中的key
// member: 成员的值
// dest: 目标指针，用于存储结果
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) ZScore(ctx context.Context, key string, member string, dest *float64) (bool, error) {
	if err := c.checkParams(key, member, dest); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).ZScore(ctx, c.realKey(key), member)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("获取有序集合成员分数失败: %w", result.Err())
	}

	*dest = result.Val()
	return true, nil
}

// SRem 从集合中移除一个成员
// key: redis中的key
// member: 要移除的成员
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) SRem(ctx context.Context, key string, member string) (bool, error) {
	if err := c.checkParams(key, member, nil); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).SRem(ctx, c.realKey(key), member)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("移除集合成员失败: %w", result.Err())
	}

	return true, nil
}

// SCard 返回集合的成员数量
// key: redis中的key
// dest: 目标指针，用于存储结果
// 返回值: (key是否存在, 错误信息)
func (c *ApiRdbClient) SCard(ctx context.Context, key string, dest *int64) (bool, error) {
	if err := c.checkParams(key, nil, dest); err != nil {
		return false, err
	}

	result := (*redis.Client)(c).SCard(ctx, c.realKey(key))
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("获取集合成员数量失败: %w", result.Err())
	}

	*dest = result.Val()
	return true, nil
}

/****************************************
				类型转换
****************************************/

// 类型转换接口
type typeConverter interface {
	toString() string
	fromString(string) error
}

// 基本类型转换实现
type (
	stringConverter struct{ value string }
	intConverter    struct{ value int64 }
	uintConverter   struct{ value uint64 }
	floatConverter  struct{ value float64 }
	boolConverter   struct{ value bool }
)

func (s stringConverter) toString() string { return s.value }
func (s *stringConverter) fromString(str string) error {
	s.value = str
	return nil
}

func (i intConverter) toString() string { return strconv.FormatInt(i.value, 10) }
func (i *intConverter) fromString(str string) error {
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}
	i.value = val
	return nil
}

func (u uintConverter) toString() string { return strconv.FormatUint(u.value, 10) }
func (u *uintConverter) fromString(str string) error {
	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return err
	}
	u.value = val
	return nil
}

func (f floatConverter) toString() string { return strconv.FormatFloat(f.value, 'f', -1, 64) }
func (f *floatConverter) fromString(str string) error {
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}
	f.value = val
	return nil
}

func (b boolConverter) toString() string { return strconv.FormatBool(b.value) }
func (b *boolConverter) fromString(str string) error {
	val, err := strconv.ParseBool(str)
	if err != nil {
		return err
	}
	b.value = val
	return nil
}

// 创建类型转换器
func newConverter(v reflect.Value) (typeConverter, error) {
	switch v.Kind() {
	case reflect.String:
		return &stringConverter{value: v.String()}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &intConverter{value: v.Int()}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &uintConverter{value: v.Uint()}, nil
	case reflect.Float32, reflect.Float64:
		return &floatConverter{value: v.Float()}, nil
	case reflect.Bool:
		return &boolConverter{value: v.Bool()}, nil
	default:
		return nil, fmt.Errorf("不支持的value类型: %v", v.Kind())
	}
}

// valueToString 将任意值转换为字符串
// v: 要转换的值
// 返回值: (转换后的字符串, 错误信息)
func valueToString(v reflect.Value) (string, error) {
	// 处理 interface{} 类型
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	// 处理指针类型
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}

	// 处理结构体类型
	if v.Kind() == reflect.Struct {
		bytes, err := json.Marshal(v.Interface())
		if err != nil {
			return "", fmt.Errorf("序列化结构体失败: %w", err)
		}
		return string(bytes), nil
	}

	// 处理基本类型
	converter, err := newConverter(v)
	if err != nil {
		return "", fmt.Errorf("创建类型转换器失败: %w", err)
	}
	return converter.toString(), nil
}

// stringToValue 将字符串转换为指定类型的值
// str: 要转换的字符串
// destType: 目标类型
// 返回值: (转换后的值, 错误信息)
func stringToValue(str string, destType reflect.Type) (reflect.Value, error) {
	// 处理结构体类型
	if destType.Kind() == reflect.Struct || (destType.Kind() == reflect.Ptr && destType.Elem().Kind() == reflect.Struct) {
		value := reflect.New(destType)
		if err := json.Unmarshal([]byte(str), value.Interface()); err != nil {
			return reflect.Value{}, fmt.Errorf("解析结构体失败: %w", err)
		}
		return value.Elem(), nil
	}

	// 处理基本类型
	value := reflect.New(destType).Elem()
	converter, err := newConverter(value)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("创建类型转换器失败: %w", err)
	}

	if err := converter.fromString(str); err != nil {
		return reflect.Value{}, fmt.Errorf("转换值失败: %w", err)
	}

	// 将转换后的值设置到目标值中
	switch v := converter.(type) {
	case *stringConverter:
		value.SetString(v.value)
	case *intConverter:
		value.SetInt(v.value)
	case *uintConverter:
		value.SetUint(v.value)
	case *floatConverter:
		value.SetFloat(v.value)
	case *boolConverter:
		value.SetBool(v.value)
	default:
		return reflect.Value{}, fmt.Errorf("不支持的转换器类型: %T", converter)
	}

	return value, nil
}

// handleStruct 统一处理结构体的序列化和反序列化
// value: 要处理的值
// redisData: 当反序列化时传入的redis数据
// isSerialize: true表示序列化(写入 value -> redisData),false表示反序列化(读取 redisData -> value)
// 返回值: 错误信息
func handleStruct(value any, redisData map[string]string, isSerialize bool) error {
	valueValue := reflect.ValueOf(value)

	// 反序列化时需要确保value是指针类型
	if isSerialize {
		// 序列化时处理指针类型
		if valueValue.Kind() == reflect.Ptr {
			if valueValue.IsNil() {
				return fmt.Errorf("value不能为空指针")
			}
			valueValue = valueValue.Elem()
		}
	} else {
		if valueValue.Kind() != reflect.Ptr {
			return fmt.Errorf("反序列化时value必须是指针类型")
		}
		if valueValue.IsNil() {
			return fmt.Errorf("value不能为空指针")
		}
		valueValue = valueValue.Elem()
	}

	// 序列化处理
	if isSerialize {
		// 处理map类型
		if valueValue.Kind() == reflect.Map {
			iter := valueValue.MapRange()
			for iter.Next() {
				keyStr, err := valueToString(iter.Key())
				if err != nil {
					return fmt.Errorf("转换key失败: %w", err)
				}

				valueStr, err := valueToString(iter.Value())
				if err != nil {
					return fmt.Errorf("转换value失败: %w", err)
				}
				redisData[keyStr] = valueStr
			}
			return nil
		}

		// 处理结构体类型
		if valueValue.Kind() == reflect.Struct {
			valueType := valueValue.Type()
			for i := 0; i < valueValue.NumField(); i++ {
				field := valueType.Field(i)
				// 跳过未导出的字段
				if !field.IsExported() {
					continue
				}
				fieldName := field.Name
				fieldValue := valueValue.Field(i)

				valueStr, err := valueToString(fieldValue)
				if err != nil {
					return fmt.Errorf("转换字段 %s 失败: %w", fieldName, err)
				}
				redisData[fieldName] = valueStr
			}
			return nil
		}

		// 处理基本类型
		valueStr, err := valueToString(valueValue)
		if err != nil {
			return fmt.Errorf("转换value失败: %w", err)
		}
		redisData["value"] = valueStr
		return nil
	}

	// 反序列化处理
	if valueValue.Kind() == reflect.Map {
		keyType := valueValue.Type().Key()
		valueType := valueValue.Type().Elem()
		destMap := reflect.MakeMap(valueValue.Type())

		for k, v := range redisData {
			keyValue, err := stringToValue(k, keyType)
			if err != nil {
				return fmt.Errorf("转换key失败: %w", err)
			}

			valueValue, err := stringToValue(v, valueType)
			if err != nil {
				return fmt.Errorf("转换value失败: %w", err)
			}
			destMap.SetMapIndex(keyValue, valueValue)
		}
		reflect.ValueOf(value).Elem().Set(destMap)
		return nil
	}

	if valueValue.Kind() == reflect.Struct {
		valueType := valueValue.Type()
		for i := 0; i < valueType.NumField(); i++ {
			field := valueType.Field(i)
			// 跳过未导出的字段
			if !field.IsExported() {
				continue
			}
			fieldName := field.Name
			fieldValue := valueValue.Field(i)

			if v, ok := redisData[fieldName]; ok {
				value, err := stringToValue(v, field.Type)
				if err != nil {
					return fmt.Errorf("转换字段 %s 失败: %w", fieldName, err)
				}
				fieldValue.Set(value)
			}
		}
		return nil
	}

	// 处理基本类型
	if v, ok := redisData["value"]; ok {
		value, err := stringToValue(v, valueValue.Type())
		if err != nil {
			return fmt.Errorf("转换value失败: %w", err)
		}
		valueValue.Set(value)
		return nil
	}

	return fmt.Errorf("不支持的类型: %v", valueValue.Kind())
}
