package gil_dict_sdk

// DictItem 表示字典项
type DictItem struct {
	MetaDictID             int    `json:"metaDictId"`
	MetaDictType           string `json:"metaDictType"`
	MetaDictKey            string `json:"metaDictKey"`
	MetaDictName           string `json:"metaDictName"`
	MetaDictParentID       int    `json:"metaDictParentId"`
	MetaDictLevel          int    `json:"metaDictLevel"`
	MetaDictOrderby        int    `json:"metaDictOrderby"`
	MetaDictStatus         int    `json:"metaDictStatus"`
	MetaDictCascadeDictIDs string `json:"metaDictCascadeDictIds"`
	MetaDictChildCount     int    `json:"metaDictChildCount"`
	MetaDictStandardCode   string `json:"metaDictStandardCode"`
	MetaDictScope          string `json:"metaDictScope"`
}

// MetaDict 表示字典类型
type MetaDict struct {
	MetaDictName  string            `json:"metaDictName"`
	MetaDictItems []DictItem        `json:"metaDictItems"`
	MetaDictKvMap map[string]string `json:"metaDictKvMap"`
	MetaDictVkMap map[string]int    `json:"metaDictVkMap"`
}

// DictResponse 表示API响应
type DictResponse struct {
	Status       int                 `json:"status"`
	Code         int                 `json:"code"`
	Message      string              `json:"message"`
	ResponseTime int64               `json:"responseTime"`
	Data         map[string]MetaDict `json:"data"`
}

// DictRequest 表示API请求
type DictRequest struct {
	MetaDictTypes []string `json:"metaDictTypes"`
	MetaDictScope string   `json:"metaDictScope"`
}

type RegionResponse struct {
	Status       int          `json:"status"`
	Code         int          `json:"code"`
	Message      string       `json:"message"`
	ResponseTime int64        `json:"response_time"`
	Data         []RegionData `json:"data"`
}

type RegionData struct {
	RegionDictName       string       `json:"regionDictName"`
	RegionDictCode       int          `json:"regionDictCode"`
	RegionDictParentCode int          `json:"regionDictParentCode"`
	SubList              []RegionData `json:"subList"` // 递归嵌套
}
