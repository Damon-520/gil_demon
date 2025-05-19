package live_service

type LiveRoomAddParams_ struct {
	LiveType    int    `json:"live_type"`   // 直播类型 1:录播 2-直播
	Name        string `json:"name"`        // 名称
	Description string `json:"description"` // 描述
	Icon        string `json:"icon"`        // 直播间图标
	Cover       string `json:"cover"`       // 直播间封面
	Sort        int    `json:"sort"`        // 排序 越小越靠前
	IsDisabled  int    `json:"is_disabled"` // 是否停用 1-停用 2-启用
	IsDefault   int    `json:"is_default"`  // 是否默认直播间 1:是 2:否
}

type LiveRoomInfoParams_ struct {
	LiveRoomId int
}

type LiveRoomListParams_ struct {
	LikeName   string
	StartDate  string
	EndDate    string
	IsDisabled int
	Page       int
	Limit      int
}

type LiveRoomEditParams_ struct {
	Id          int    `json:"id"`
	LiveType    int    `json:"live_type"`   // 直播类型 1:录播 2-直播
	Name        string `json:"name"`        // 名称
	Description string `json:"description"` // 描述
	Icon        string `json:"icon"`        // 直播间图标
	Cover       string `json:"cover"`       // 直播间封面
	Sort        int    `json:"sort"`        // 排序 越小越靠前
	IsDisabled  int    `json:"is_disabled"` // 是否停用 1-停用 2-启用
	IsDefault   int    `json:"is_default"`  // 是否默认直播间 1:是 2:否
}

type LiveRoomUpdateParams_ struct {
	Id         int `json:"id"`
	Sort       int `json:"sort"`        // 排序 越小越靠前
	IsDisabled int `json:"is_disabled"` // 是否停用 1-停用 2-启用
	IsDefault  int `json:"is_default"`  // 是否默认直播间 1:是 2:否
}
