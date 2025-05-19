package resource_favorite

import (
	"context"
	"gil_teacher/app/dao/resource_favorite"

	"gil_teacher/app/consts"
	"gil_teacher/app/model/api"
)

// ResourceFavoriteService 资源收藏服务
type ResourceFavoriteService struct {
	favoriteDAO *resource_favorite.ResourceFavoriteDAO
}

// NewResourceFavoriteService 创建资源收藏服务
func NewResourceFavoriteService(favoriteDAO *resource_favorite.ResourceFavoriteDAO) *ResourceFavoriteService {
	return &ResourceFavoriteService{
		favoriteDAO: favoriteDAO,
	}
}

// CreateFavorite 创建资源收藏
func (s *ResourceFavoriteService) CreateFavorite(ctx context.Context, req *api.CreateResourceFavoriteReq) error {
	// 检查是否已经收藏
	favorite, err := s.favoriteDAO.GetByTeacherAndResource(ctx, req.TeacherID, req.ResourceID)
	if err == nil {
		// 已存在，更新状态为收藏
		favorite.Status = boolToInt64(true)
		return s.favoriteDAO.Update(ctx, favorite)
	}

	// 不存在，创建新收藏
	favorite = &resource_favorite.TeacherResourceFavorite{
		TeacherID:    req.TeacherID,
		SchoolID:     req.SchoolID,
		ResourceID:   req.ResourceID,
		ResourceType: consts.RESOURCE_TYPE_OTHER, // 默认为其他资源
		Status:       boolToInt64(true),
	}

	return s.favoriteDAO.Create(ctx, favorite)
}

// ListFavorites 获取资源收藏列表
func (s *ResourceFavoriteService) ListFavorites(ctx context.Context, req *api.ListResourceFavoriteReq) (*api.ListResourceFavoriteResp, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize
	favorites, total, err := s.favoriteDAO.List(ctx, req.TeacherID, req.SchoolID, offset, req.PageSize)
	if err != nil {
		return nil, err
	}

	resp := &api.ListResourceFavoriteResp{
		Total:     total,
		Favorites: make([]api.ResourceFavoriteResp, 0, len(favorites)),
	}

	for _, f := range favorites {
		resp.Favorites = append(resp.Favorites, api.ResourceFavoriteResp{
			ID:           f.ID,
			ResourceID:   f.ResourceID,
			ResourceType: f.ResourceType,
			Status:       f.Status,
			CreateTime:   f.CreateTime,
			UpdateTime:   f.UpdateTime,
			TeacherID:    f.TeacherID,
			SchoolID:     f.SchoolID,
		})
	}

	return resp, nil
}

// CancelFavorite 取消收藏
func (s *ResourceFavoriteService) CancelFavorite(ctx context.Context, req *api.CancelResourceFavoriteReq) error {
	favorite, err := s.favoriteDAO.GetByTeacherAndResource(ctx, req.TeacherID, req.ResourceID)
	if err != nil {
		return err
	}

	favorite.Status = boolToInt64(false)
	return s.favoriteDAO.Update(ctx, favorite)
}

// boolToInt64 将bool转换为int64
func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
