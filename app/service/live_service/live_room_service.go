package live_service

import (
	"context"
	"errors"
	"strings"
	"time"

	live_room2 "gil_teacher/app/dao/live_room"
	"gil_teacher/app/third_party/gorm_builder"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jinzhu/copier"
)

type LiveRoomService struct {
	liveRoomDao live_room2.ILiveRoomDao
	log         *log.Helper
}

func NewLiveRoomService(logger log.Logger, liveRoomDao live_room2.ILiveRoomDao) *LiveRoomService {

	sc := LiveRoomService{
		liveRoomDao: liveRoomDao,
		log:         log.NewHelper(log.With(logger, "x_module", "service/NewLiveRoomService")),
	}

	return &sc
}

func (l *LiveRoomService) Add(ctx context.Context, params LiveRoomAddParams_) (lastId int, err error) {

	var newData live_room2.LiveRoom

	_ = copier.Copy(&newData, &params)

	newData.UniqueID = time.Now().UnixNano()

	return l.liveRoomDao.Save(ctx, newData)
}

func (l *LiveRoomService) Info(ctx context.Context, params LiveRoomInfoParams_) (info live_room2.LiveRoom, err error) {

	condition := make(map[string]interface{}, 8)

	if params.LiveRoomId != 0 {
		condition["eq|id"] = params.LiveRoomId
	}

	options := gorm_builder.Options{
		Conditions: condition,
	}

	return l.liveRoomDao.Info(ctx, options)
}

func (l *LiveRoomService) InfoById(ctx context.Context, liveRoomId int) (info live_room2.LiveRoom, err error) {

	options := gorm_builder.Options{
		Conditions: map[string]interface{}{
			"eq|id": liveRoomId,
		},
	}

	return l.liveRoomDao.Info(ctx, options)

}

func (l *LiveRoomService) List(ctx context.Context, params LiveRoomListParams_) (list []live_room2.LiveRoom, total int64, err error) {

	condition := make(map[string]interface{}, 8)

	if params.LikeName != "" {
		condition["like|name"] = params.LikeName
	}

	// 需求是时间范围只要存在交集，就算符合条件
	// 所以  !(params.StartDate > end_date || params.EndDate < start_date) 计算交集
	// 等价于 (params.StartDate <= end_date && params.EndDate >= start_date)
	// 等价于 (start_date <= params.EndDate && end_date >= params.StartDate)
	if params.StartDate != "" {
		condition["lte|start_date"] = params.EndDate
	}

	if params.EndDate != "" {
		condition["gte|end_date"] = params.StartDate
	}

	if params.IsDisabled != 0 {
		condition["eq|is_disabled"] = params.IsDisabled
	}

	options := gorm_builder.Options{
		Conditions: condition,
		Order:      "sort ASC,updated_at DESC",
		Limit:      params.Limit,
		Offset:     (params.Page - 1) * params.Limit,
		IsCount:    true,
	}

	return l.liveRoomDao.List(ctx, options)
}

func (l *LiveRoomService) Edit(ctx context.Context, params LiveRoomEditParams_) (rows int64, err error) {

	if params.Id <= 0 {
		return 0, errors.New("id不能为空")
	}

	// 条件
	condition := make(map[string]interface{}, 1)
	if params.Id > 0 {
		condition["eq|id"] = params.Id
	}
	options := gorm_builder.Options{
		Conditions: condition,
	}
	// TODO 验证更新条件不能为空

	// // 将结构体转化为 JSON
	// jsonData, err := json.Marshal(params)
	// if err != nil {
	// 	return 0, errors.New("转换更新数据失败")
	// }
	//
	// // 将 JSON 转化为 map
	// var dataMap map[string]interface{}
	// err = json.Unmarshal(jsonData, &dataMap)
	// if err != nil {
	// 	return 0, errors.New("转换更新数据失败")
	// }

	// _uptStringFields := []string{
	// 	"name", "description", "start_date", "end_date", "start_time_slot", "end_time_slot",
	// 	"welcome_note", "icon", "cover", "ending_screen_img", "fixed_bottom_img",
	// }
	//
	// _uptIntFields := []string{
	// 	"time_slot_type", "sort", "is_disabled", "is_default",
	// }

	uptData := make(map[string]interface{}, 20)

	if v := strings.Trim(params.Name, " "); v != "" {
		uptData["name"] = v
	}

	if v := strings.Trim(params.Description, " "); v != "" {
		uptData["description"] = v
	}

	if v := strings.Trim(params.Icon, " "); v != "" {
		uptData["icon"] = v
	}

	if v := strings.Trim(params.Cover, " "); v != "" {
		uptData["cover"] = v
	}

	if params.Sort != 0 {
		uptData["sort"] = params.Sort
	}

	if params.IsDisabled != 0 {
		uptData["is_disabled"] = params.IsDisabled
	}

	if params.IsDefault != 0 {
		uptData["is_default"] = params.IsDefault
	}

	// fmt.Printf("uptData: %+v\n", uptData)

	return l.liveRoomDao.Update(ctx, options, uptData)
}

func (l *LiveRoomService) Update(ctx context.Context, params LiveRoomUpdateParams_) (rows int64, err error) {

	if params.Id <= 0 {
		return 0, errors.New("id不能为空")
	}

	// 条件
	condition := make(map[string]interface{}, 1)
	if params.Id > 0 {
		condition["eq|id"] = params.Id
	}
	options := gorm_builder.Options{
		Conditions: condition,
	}
	// TODO 验证更新条件不能为空

	uptData := make(map[string]interface{}, 20)

	if params.Sort != 0 {
		uptData["sort"] = params.Sort
	}

	if params.IsDisabled != 0 {
		uptData["is_disabled"] = params.IsDisabled
	}

	if params.IsDefault != 0 {
		uptData["is_default"] = params.IsDefault
	}

	if len(uptData) == 0 { // 如果没有更新项，直接返回
		return
	}

	// fmt.Printf("uptData: %+v\n", uptData)

	return l.liveRoomDao.Update(ctx, options, uptData)
}
