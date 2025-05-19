package gil_dict_sdk

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gitlab.xiaoluxue.cn/be-app/gil_dict_sdk/cache"
	"gitlab.xiaoluxue.cn/be-app/gil_dict_sdk/logger"
	"time"
)

// DictClient 表示字典客户端
type DictClient struct {
	ctx         context.Context
	restyClient *resty.Client
	cache       cache.DictCache
	logger      logger.Logger
	option      DictClientOption
}

type DictClientOption struct {
	Domain           string
	Timeout          time.Duration
	RetryCount       int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration
}

// NewDictClient 创建一个新的字典客户端
func NewDictClient(ctx context.Context, cacheIns cache.DictCache, logger logger.Logger, dictOption DictClientOption) *DictClient {

	restyClient := resty.New()
	restyClient.SetTimeout(dictOption.Timeout)
	restyClient.SetRetryCount(dictOption.RetryCount)
	restyClient.SetRetryWaitTime(dictOption.RetryWaitTime)
	restyClient.SetRetryMaxWaitTime(dictOption.RetryMaxWaitTime)

	var err error
	//默认使用本地缓存
	if cacheIns == nil {
		cacheIns, err = cache.NewBigCache(ctx)
		if err != nil {
			logger.Fatal(ctx, fmt.Sprintf("NewDictClient:NewBigCache:%v", err))
		}
	}

	return &DictClient{
		ctx:         ctx,
		restyClient: restyClient,
		cache:       cacheIns,
		logger:      logger,
		option:      dictOption,
	}
}

func (c *DictClient) getActionUrl(path string) string {
	return fmt.Sprintf("%s%s", c.option.Domain, path)
}
