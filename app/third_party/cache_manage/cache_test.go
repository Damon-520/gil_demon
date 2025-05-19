package cache_manage

import (
	"fmt"
	"testing"
)

func TestGenerateCacheKey(t *testing.T) {

	var err error
	var params map[string]any
	var cacheKey string

	params = map[string]any{"task_id": 1001}
	cacheKey, err = GenerateCacheKey("taskInfo", "hash", params)
	if nil == err {
		fmt.Println("taskInfo: ", cacheKey)
	} else {
		fmt.Println(err)
	}

	params = map[string]any{"card_id": 1001, "user_id": 1155}
	cacheKey, err = GenerateCacheKey("cardInfo", "list", params)
	if nil == err {
		fmt.Println("cardInfo: ", cacheKey)
	} else {
		fmt.Println(err)
	}

	cacheKey, err = GenerateCacheKey("cardHot", "list", nil)
	if nil == err {
		fmt.Println("cardHot: ", cacheKey)
	} else {
		fmt.Println(err)
	}

}
