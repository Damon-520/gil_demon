package gil_dict_sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// GetDicts 获取字典数据
func (c *DictClient) getDicts(ctx context.Context, types []string, scope string) (*DictResponse, error) {

	req := DictRequest{
		MetaDictTypes: types,
		MetaDictScope: scope,
	}

	var response DictResponse

	actionUrl := c.getActionUrl(HTTP_GET_META_DICTS_PATH)

	_, err := c.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&response).
		Post(actionUrl)

	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("param:%v,getDicts:%v", req, err))
		return nil, err
	}

	if response.Status != http.StatusOK {
		c.logger.Error(ctx, fmt.Sprintf("dict:getDicts:%v,Status:%d", response.Message, response.Status))
		return nil, errors.New(response.Message)
	}

	if response.Code != 0 {
		c.logger.Error(ctx, fmt.Sprintf("dict:getDicts:%v,Code:%d", response.Message, response.Code))
		return nil, errors.New(response.Message)
	}

	return &response, nil
}

// GetDictByType 获取指定类型的字典数据
func (c *DictClient) GetDictByType(ctx context.Context, dictType string, scope string) (*MetaDict, error) {

	cacheKey := fmt.Sprintf(CACHE_KEY_GET_META_DICT_BY_TYPE, scope, dictType)

	dictJsonStr, err := c.cache.Get(ctx, cacheKey)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetDictByType:GetDictByType:cache.Get:%v", err))
	}

	var dict MetaDict
	err = json.Unmarshal([]byte(dictJsonStr), &dict)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetDictByType:json.Unmarshal:%v", err))
	}
	if len(dictJsonStr) > 0 {
		return &dict, nil
	}

	response, err := c.getDicts(ctx, []string{dictType}, scope)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetDictByType:getDicts:%v", err))
		return nil, err
	}

	if dict, ok := response.Data[dictType]; ok {
		dictByte, err := json.Marshal(dict)
		if err == nil && len(dictByte) > 0 {
			err = c.cache.Set(ctx, cacheKey, string(dictByte))
			if err != nil {
				c.logger.Error(ctx, fmt.Sprintf("GetDictByType:Cache.Set:%v", err))
			}
		}
		return &dict, nil
	}
	c.logger.Error(ctx, fmt.Sprintf("dict %s not found", dictType))
	return nil, fmt.Errorf("dict %s not found", dictType)
}

// GetDictByTypes 批量获取指定类型的字典数据
func (c *DictClient) GetDictByTypes(ctx context.Context, dictTypes []string, scope string) (map[string]MetaDict, error) {

	sort.Strings(dictTypes)

	cacheKeyPrefix := fmt.Sprintf(CACHE_KEY_GET_META_DICT_BY_TYPE, scope, "")

	sb := strings.Builder{}
	sb.WriteString(cacheKeyPrefix)
	for _, dictType := range dictTypes {
		sb.WriteString(dictType)
	}
	cacheKey := sb.String()

	cacheStr, err := c.cache.Get(ctx, cacheKey)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetDictByTypes:GetDictByType:cache.Get:%v", err))
	}

	if len(cacheStr) > 0 {
		result := make(map[string]MetaDict, 0)
		err = json.Unmarshal([]byte(cacheStr), &result)
		if err != nil {
			c.logger.Error(ctx, fmt.Sprintf("GetDictByTypes:json.Unmarshal:%v", err))
		}
		return result, err
	}

	response, err := c.getDicts(ctx, dictTypes, scope)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetDictByTypes:getDicts:%v", err))
		return nil, err
	}
	for _, dictType := range dictTypes {
		if _, ok := response.Data[dictType]; !ok {
			response.Data[dictType] = MetaDict{}
		}
	}

	jsonStr, err := json.Marshal(response.Data)
	err = c.cache.Set(ctx, cacheKey, string(jsonStr))
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetDictByTypes:Cache.Set:%v", err))
	}

	return response.Data, nil
}
