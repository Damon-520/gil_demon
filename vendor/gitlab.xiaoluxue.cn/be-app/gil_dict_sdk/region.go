package gil_dict_sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (c *DictClient) GetAllRegion(ctx context.Context) ([]RegionData, error) {

	regions, err := c.cache.Get(ctx, CACHE_KEY_GET_ALL_REGION)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetAllRegion:%v", err))
	}
	if len(regions) > 0 {
		regionData := make([]RegionData, 0)
		err = json.Unmarshal([]byte(regions), &regionData)
		if err != nil {
			c.logger.Error(ctx, fmt.Sprintf("GetAllRegion:%v", err))
		} else {
			return regionData, nil
		}
	}

	var resp RegionResponse

	actionUrl := c.getActionUrl(HTTP_GET_ALL_REGIONS_PATH)

	_, err = c.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resp).
		Get(actionUrl)

	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetAllRegion:%v", err))
		return nil, err
	}

	if resp.Status != http.StatusOK {
		c.logger.Error(ctx, fmt.Sprintf("GetAllRegion:%v,Status:%d", resp.Message, resp.Status))
		return nil, errors.New(resp.Message)
	}

	if resp.Code != 0 {
		c.logger.Error(ctx, fmt.Sprintf("GetAllRegion:%v,Code:%d", resp.Message, resp.Code))
		return nil, errors.New(resp.Message)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetAllRegion:%v", err))
		return nil, err
	}

	dataStr := string(data)

	err = c.cache.Set(ctx, CACHE_KEY_GET_ALL_REGION, dataStr)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetAllRegion:%v", err))
	}

	return resp.Data, nil

}

func (c *DictClient) GetAllProvince(ctx context.Context) ([]RegionData, error) {
	regions, err := c.cache.Get(ctx, CACHE_KEY_GET_ALL_PROVINCE)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetAllProvince:%v", err))
	}
	if len(regions) > 0 {
		regionData := make([]RegionData, 0)
		err = json.Unmarshal([]byte(regions), &regionData)
		if err != nil {
			c.logger.Error(ctx, fmt.Sprintf("GetAllProvince:%v", err))
		} else {
			return regionData, nil
		}
	}

	var resp RegionResponse

	actionUrl := c.getActionUrl(HTTP_GET_ALL_PROVINCE_PATH)

	_, err = c.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resp).
		Get(actionUrl)

	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetAllProvince:%v", err))
		return nil, err
	}

	if resp.Status != http.StatusOK {
		c.logger.Error(ctx, fmt.Sprintf("GetAllProvince:%v,Status:%d", resp.Message, resp.Status))
		return nil, errors.New(resp.Message)
	}

	if resp.Code != 0 {
		c.logger.Error(ctx, fmt.Sprintf("GetAllProvince:%v,Code:%d", resp.Message, resp.Code))
		return nil, errors.New(resp.Message)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetAllProvince:%v", err))
		return nil, err
	}

	dataStr := string(data)

	err = c.cache.Set(ctx, CACHE_KEY_GET_ALL_PROVINCE, dataStr)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("GetAllProvince:%v", err))
	}

	return resp.Data, nil
}
