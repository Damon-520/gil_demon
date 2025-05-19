package cache

import (
	"context"
	"github.com/allegro/bigcache"
	"time"
)

type BigCache struct {
	ctx              context.Context
	bigCacheInstance *bigcache.BigCache
}

func NewBigCache(ctx context.Context) (*BigCache, error) {
	config := bigcache.Config{
		Shards:             1024,                  // 分片数量（推荐 CPU 核数 x 256）
		LifeWindow:         24 * 60 * time.Minute, // 数据存活时间
		CleanWindow:        60 * time.Minute,      // 清理间隔（0 表示禁用自动清理）
		MaxEntriesInWindow: 1000 * 10 * 60,        // 存活窗口内最大条目数（建议 LifeWindow*预估QPS）
		MaxEntrySize:       10 * 1024,             // 单条数据最大大小（字节）
		HardMaxCacheSize:   10,                    // 最大内存限制（MB）- 重要！
		Verbose:            true,                  // 日志输出清理过程
	}
	bigCacheIns, err := bigcache.NewBigCache(config)
	if err != nil {
		return &BigCache{ctx: ctx}, err
	}

	return &BigCache{
		ctx:              ctx,
		bigCacheInstance: bigCacheIns,
	}, nil

}

func (c *BigCache) Set(ctx context.Context, key string, value string) error {
	return c.bigCacheInstance.Set(key, []byte(value))
}

func (c *BigCache) Get(ctx context.Context, key string) (string, error) {
	dataByte, err := c.bigCacheInstance.Get(key)
	return string(dataByte), err
}
