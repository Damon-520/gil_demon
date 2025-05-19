package admin_service

import (
	"context"

	"gil_teacher/app/conf"
	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"

	"gitlab.xiaoluxue.cn/be-app/gil_dict_sdk"
)

// AdminClient Admin API 客户端
type AdminClient struct {
	client *gil_dict_sdk.DictClient // 客户端
	log    *logger.ContextLogger
}

// NewAdminClient 创建 Admin API 客户端
func NewAdminClient(config *conf.Conf, l *logger.ContextLogger) (*AdminClient, error) {
	client := gil_dict_sdk.NewDictClient(
		context.Background(), nil, l, gil_dict_sdk.DictClientOption{
			Domain:  config.Config.GilAdminAPI.AdminHost,
			Timeout: consts.TeacherDefaultAPITimeout,
		},
	)
	return &AdminClient{
		client: client,
		log:    l,
	}, nil
}

// GetAllProvince 获取所有省份信息，返回 map[省份编码]省份名称
func (c *AdminClient) GetAllProvince(ctx context.Context) (map[int64]string, error) {
	provinceList, err := c.client.GetAllProvince(ctx)
	if err != nil {
		c.log.Error(ctx, "获取省份列表失败: %v", err)
		return make(map[int64]string), nil
	}
	provinceMap := make(map[int64]string)
	for _, province := range provinceList {
		provinceMap[int64(province.RegionDictCode)] = province.RegionDictName
	}
	return provinceMap, nil
}

// GetProvinceNameByCode 通过省份编码获取省份名称，如果省份编码不存在，则返回空字符串
func (c *AdminClient) GetProvinceNameByCode(ctx context.Context, code int64) (string, error) {
	// SDK 内部已经做了缓存
	provinceMap, err := c.GetAllProvince(ctx)
	if err != nil {
		return "", err
	}
	return provinceMap[code], nil
}
