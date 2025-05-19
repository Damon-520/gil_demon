package live_http

/**
 * live-room-base
 */

type LiveRoomVo struct {
	ID          int    `json:"id"`
	UniqueID    int64  `json:"unique_id"`
	LiveType    int    `json:"live_type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Cover       string `json:"cover"`
	Sort        int    `json:"sort"`
	IsDisabled  int    `json:"is_disabled"`
	IsDefault   int    `json:"is_default"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

/**
 * live-room-create
 */

type LiveRoomCreateRequest struct {
	Name        string `json:"name" binding:"required"`        // 直播间名称，必填字段
	Description string `json:"description" binding:"required"` // 直播间描述
	Icon        string `json:"icon" binding:"required"`        // 直播间图标URL
	Cover       string `json:"cover" binding:"required"`       // 直播间封面URL
	Sort        int    `json:"sort"`                           // 排序字段
	IsDisabled  int    `json:"is_disabled" binding:"required"` // 是否禁用（1表示禁用，0表示启用）
	IsDefault   int    `json:"is_default" binding:"required"`  // 是否默认直播间（1表示默认，0表示非默认）
}

type LiveRoomCreateResult struct {
	LastId int `json:"last_id"`
}

/**
 * live-room-info
 */

type LiveRoomInfoRequest struct {
	Id int `json:"id" binding:"required"`
}

type LiveRoomInfoResult struct {
	LiveRoomVo
}

/**
 * live-room-list
 */

type LiveRoomListRequest struct {
	Name       string `json:"name"`
	IsDisabled int    `json:"is_disabled"`
	DateRange  struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	} `json:"date_range"`

	// consts.Pagination // 分页参数
}

type LiveRoomListResult struct {
	List []LiveRoomVo `json:"list"`
	// PageInfo consts.PageInfo `json:"page_info"`
}

/**
 * live-room-edit
 */

type LiveRoomEditRequest struct {
	Id          int    `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`        // 直播间名称，必填字段
	Description string `json:"description" binding:"required"` // 直播间描述
	Icon        string `json:"icon" binding:"required"`        // 直播间图标URL
	Cover       string `json:"cover" binding:"required"`       // 直播间封面URL
	Sort        int    `json:"sort"`                           // 排序字段
	IsDisabled  int    `json:"is_disabled" binding:"required"` // 是否禁用（1表示禁用，0表示启用）
	IsDefault   int    `json:"is_default" binding:"required"`  // 是否默认直播间（1表示默认，0表示非默认）
}

type LiveRoomEditResult struct {
	Rows int `json:"rows"`
}

/**
 * live-room-update
 */

type LiveRoomUpdateRequest struct {
	Id         int `json:"id" binding:"required"`
	Sort       int `json:"sort"`        // 排序字段
	IsDisabled int `json:"is_disabled"` // 是否禁用（1表示禁用，0表示启用）
	IsDefault  int `json:"is_default"`  // 是否默认直播间（1表示默认，0表示非默认）
}

type LiveRoomUpdateResult struct {
	Rows int `json:"rows"`
}
