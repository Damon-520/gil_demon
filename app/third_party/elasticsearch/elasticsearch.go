package elasticsearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gil_teacher/app/conf"

	"github.com/elastic/go-elasticsearch/v8"
)

var client *elasticsearch.Client

func InitES(esConfig *conf.Elasticsearch) (*elasticsearch.Client, error) {
	if esConfig == nil {
		return nil, fmt.Errorf("Elasticsearch configuration is nil")
	}
	cfg := elasticsearch.Config{
		Addresses: []string{esConfig.EsURL},
		Username:  esConfig.Username,
		Password:  esConfig.Password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 跳过证书验证
			},
		},
	}

	var err error
	client, err = elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// 测试连接
	res, err := client.Ping()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("Error pinging Elasticsearch: %s", res.Status())
	}
	log.Println("Elasticsearch connection successful!")
	return client, nil
}

// IndexExists 检查索引是否存在
func IndexExists(indexName string) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("Elasticsearch client is not initialized")
	}
	res, err := client.Indices.Exists([]string{indexName})
	if err != nil {
		return false, fmt.Errorf("failed to check if index exists: %w", err)
	}
	defer res.Body.Close()
	return res.StatusCode == 200, nil
}

// CreateIndex 创建索引
func CreateIndex(index string) error {
	exists, err := IndexExists(index)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("索引 %s 已经存在", index)
		return nil
	}

	res, err := client.Indices.Create(index)
	if err != nil {
		return fmt.Errorf("cannot create index: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}
	log.Printf("Index %s created successfully\n", index)
	return nil
}

// SearchDocument 执行查询并返回结果
func SearchDocument(index string, query map[string]interface{}) (interface{}, error) {
	// 将查询转换为 JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("error marshaling query: %v", err)
	}

	queryReader := bytes.NewReader(queryJSON)

	// 执行搜索请求
	res, err := client.Search(
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex(index),
		client.Search.WithBody(queryReader),
		client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %v", err)
	}
	defer res.Body.Close()

	// 解析响应结果
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return result, nil
}
