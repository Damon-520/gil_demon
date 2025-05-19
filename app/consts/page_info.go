package consts

import (
	"fmt"
	"gil_teacher/app/controller/http_server/response"
	"slices"
)

const (
	// DB_DEFAULT_PAGE_LIMIT 默认分页大小
	DB_DEFAULT_PAGE_LIMIT = 100

	// API_DEFAULT_PAGE_SIZE 默认分页大小
	API_DEFAULT_PAGE_SIZE = 10
	// API_MAX_PAGE_SIZE 最大分页大小
	API_MAX_PAGE_SIZE = 100
	// API_MAX_TOTAL_SIZE 最大总记录数
	API_MAX_TOTAL_SIZE = 10000
)

// 排序类型
type SortType string

const (
	SortTypeAsc  SortType = "asc"
	SortTypeDesc SortType = "desc"
)

// API 响应分页结构体
type ApiPageResponse struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"pageSize"`
	Total    int64 `json:"total"`
}

// API 请求分页结构体
type APIReqeustPageInfo struct {
	Page          int64    `json:"page,omitempty"`
	PageSize      int64    `json:"pageSize,omitempty"`
	SortBy        string   `json:"sortBy,omitempty"`   // 排序字段
	SortType      SortType `json:"sortType,omitempty"` // 排序类型，asc/desc
	ValidSortKeys []string `json:"-"`                  // 有效排序字段
	All           bool     `json:"-"`                  // 是否获取全部数据
}

// 先检查，不合理赋默认值，最后返回错误，确保使用时既可以使用错误，也可以忽略错误使用默认值
func (p *APIReqeustPageInfo) Check() (resultErr *response.Response) {
	var err error
	p.Page, p.PageSize, err = PageHandler(p.Page, p.PageSize)
	if err != nil {
		p.Page = 1
		p.PageSize = API_DEFAULT_PAGE_SIZE
		resultErr = &response.ERR_INVALID_PAGE
	}

	// 排序字段检查
	if p.SortBy != "" {
		if !slices.Contains(p.ValidSortKeys, p.SortBy) {
			p.SortBy = ""
			resultErr = &response.ERR_INVALID_ORDER_BY
		}
		p.SortBy, p.SortType = SortHandler(p.SortBy, p.SortType)
	}
	return
}

// 将 API 请求分页结构体转换为 DB 查询分页结构体
func (p *APIReqeustPageInfo) ToDBPageInfo() *DBPageInfo {
	return &DBPageInfo{
		Page:     p.Page,
		Limit:    p.PageSize,
		SortBy:   p.SortBy,
		SortType: p.SortType,
		All:      p.All,
	}
}

// 足够大的分页，获取全部数据
func AllDataPageInfo() *APIReqeustPageInfo {
	return &APIReqeustPageInfo{
		Page:     1,
		PageSize: 10000,
		All:      true,
	}
}

// 分页处理
func PageHandler(page, pageSize int64) (int64, int64, error) {
	if page*pageSize > API_MAX_TOTAL_SIZE {
		return page, pageSize, fmt.Errorf("分页数量不能超过 %d", API_MAX_TOTAL_SIZE)
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > API_MAX_PAGE_SIZE {
		pageSize = API_DEFAULT_PAGE_SIZE
	}
	return page, pageSize, nil
}

// 排序处理
func SortHandler(sortBy string, sortType SortType) (string, SortType) {
	if sortBy == "" {
		return "", ""
	}
	if sortType == "" || (sortType != SortTypeAsc && sortType != SortTypeDesc) {
		sortType = SortTypeDesc
	}
	return sortBy, sortType
}

// db 查询使用的分页结构体 page 和 limit, 排序字段及方式
type DBPageInfo struct {
	Page     int64    `json:"page"`
	Limit    int64    `json:"limit"`
	SortBy   string   `json:"sortBy"`
	SortType SortType `json:"sortType"`
	All      bool     `json:"-"` // 是否获取全部数据
}

func (p *DBPageInfo) Check() {
	if p.Page <= 0 {
		p.Page = 1
	}

	if p.Limit <= 0 || p.Limit > DB_DEFAULT_PAGE_LIMIT {
		p.Limit = DB_DEFAULT_PAGE_LIMIT
	}

	p.SortBy, p.SortType = SortHandler(p.SortBy, p.SortType)
}

func DefaultDBPageInfo(p *DBPageInfo) *DBPageInfo {
	if p == nil {
		return &DBPageInfo{
			Page:  1,
			Limit: DB_DEFAULT_PAGE_LIMIT,
		}
	}
	p.Check()
	return p
}
