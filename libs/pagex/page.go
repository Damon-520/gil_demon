package pagex

const (
	DefaultPage  = 1
	DefaultLimit = 5
	MaxLimit     = 30
)

type PageInfo struct {
	Page    int32 `json:"page"`
	Limit   int32 `json:"limit"`
	MaxPage int32 `json:"max_page"`
	Total   int32 `json:"total"`
}

type PageParams struct {
	Page            int32
	PageSize        int32
	LimitDefault    int32
	DefaultMaxLimit int32
	Total           int32
}

// GetPageInfo 分页参数
func GetPageInfo(params PageParams) (ret PageInfo) {
	if params.Page <= 0 {
		ret.Page = DefaultPage
	} else {
		ret.Page = params.Page
	}

	if params.LimitDefault <= 0 {
		params.LimitDefault = DefaultLimit
	}

	if params.PageSize <= 0 {
		ret.Limit = params.LimitDefault
	} else {
		ret.Limit = params.PageSize
	}

	if params.DefaultMaxLimit > 0 {
		if ret.Limit > params.DefaultMaxLimit {
			ret.Limit = params.DefaultMaxLimit
		}
	} else {
		if ret.Limit > MaxLimit {
			ret.Limit = MaxLimit
		}
	}

	ret.Total = params.Total
	ret.MaxPage = GetMaxPage(ret.Total, ret.Limit)
	return ret
}

func GetMaxPage(total int32, limit int32) int32 {
	if total == 0 || limit == 0 {
		return 1
	}
	if total <= limit {
		return 1
	}
	f := total % limit
	if f > 0 {
		return (total / limit) + 1
	} else {
		return total / limit
	}
}
