package api

// CreateResourceFavoriteReq 创建资源收藏请求
type CreateResourceFavoriteReq struct {
	TeacherID  int64  `json:"-"`
	SchoolID   int64  `json:"-"`
	ResourceID string `json:"resourceId" binding:"required"`
}

// ResourceFavoriteResp 资源收藏响应
type ResourceFavoriteResp struct {
	ID           int64  `json:"id"`
	ResourceID   string `json:"resourceId"`
	ResourceType int64  `json:"resourceType"`
	Status       int64  `json:"status"`
	CreateTime   int64  `json:"createTime"`
	UpdateTime   int64  `json:"updateTime"`
	TeacherID    int64  `json:"-"`
	SchoolID     int64  `json:"-"`
}

// ListResourceFavoriteReq 获取资源收藏列表请求
type ListResourceFavoriteReq struct {
	TeacherID int64 `json:"-"`
	SchoolID  int64 `json:"-"`
	Page      int64 `json:"page" binding:"required,min=1"`
	PageSize  int64 `json:"pageSize" binding:"required,min=1,max=100"`
}

// ListResourceFavoriteResp 获取资源收藏列表响应
type ListResourceFavoriteResp struct {
	Total     int64                  `json:"total"`
	Favorites []ResourceFavoriteResp `json:"favorites"`
}

// CancelResourceFavoriteReq 取消资源收藏请求
type CancelResourceFavoriteReq struct {
	TeacherID  int64  `json:"-"`
	SchoolID   int64  `json:"-"`
	ResourceID string `json:"resourceId" binding:"required"`
}
