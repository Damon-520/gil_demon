package live_http

import (
	"context"
	"fmt"

	"gil_teacher/app/conf"
	liveSc "gil_teacher/app/service/live_service"
	"gil_teacher/app/third_party/errorx"
	pb "gil_teacher/proto/gen/go/proto/gil_teacher/api" // 替换成你的proto包路径

	"github.com/go-kratos/kratos/v2/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc"
)

type LiveRoomHttp struct {
	pb.UnimplementedLiveRoomServer
	liveRoomSc *liveSc.LiveRoomService
	log        *log.Helper
	config     *conf.Config
}

func NewLiveRoomHttp(
	liveRoomSc *liveSc.LiveRoomService,
	logger log.Logger,
	config *conf.Config,
) *LiveRoomHttp {
	return &LiveRoomHttp{
		liveRoomSc: liveRoomSc,
		config:     config,
		log:        log.NewHelper(log.With(logger, "x_module", "controller/NewLiveRoomHttp")),
	}
}

// 注册gRPC服务
func (s *LiveRoomHttp) RegisterGRPC(server grpc.ServiceRegistrar) {
	pb.RegisterLiveRoomServer(server, s)
}

// 注册HTTP服务
func (l *LiveRoomHttp) RegisterHTTP(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pb.RegisterLiveRoomHandler(ctx, mux, conn)
}

// Create 直播间详情
func (l *LiveRoomHttp) Create(ctx context.Context, req *pb.LiveRoomCreateRequest) (*pb.LiveRoomCreateResponse, error) {

	// fmt.Printf("req: %+v\n", req)

	// 权限
	// uInfo, ok := auth.GetAdminInfo(c)
	// if !ok {
	// 	l.log.WithContext(c).Info("LiveVideoPointHttp:Create:认证错误")
	// 	resp.Error(c, errorx.ErrAuthFail)
	// 	return
	// }

	params := liveSc.LiveRoomAddParams_{
		LiveType: 1,
	}

	_ = copier.Copy(&params, &req)

	fmt.Printf("params: %+v\n", params)

	lastId, err := l.liveRoomSc.Add(ctx, params)
	if err != nil {
		l.log.WithContext(ctx).Infof("LiveRoomHttp:Create:添加错误:Err:%v", err)
		return nil, errorx.Cause(err)
	}

	return &pb.LiveRoomCreateResponse{
		Data: &pb.LiveRoomCreateResponse_Data{
			LastId: int64(lastId),
		},
	}, nil
}

// Info 直播间详情
func (l *LiveRoomHttp) Info(ctx context.Context, req *pb.LiveRoomInfoRequest) (*pb.LiveRoomInfoResponse, error) {
	params := liveSc.LiveRoomInfoParams_{LiveRoomId: int(req.Id)}

	liveRoom, err := l.liveRoomSc.Info(ctx, params)
	if err != nil {
		l.log.WithContext(ctx).Infof("LiveRoomHttp:Info:获取结果错误:Err:%v", err)
		return nil, errorx.Cause(err)
	}

	var liveRoomR pb.LiveRoomVo
	if err := copier.Copy(&liveRoomR, liveRoom); err != nil {
		return nil, errorx.Cause(err)
	}
	response := &pb.LiveRoomInfoResponse{
		Data: &pb.LiveRoomInfoResponse_Data{
			LiveRoom: &liveRoomR,
		},
	}

	return response, nil
}

// List 直播间列表
func (l *LiveRoomHttp) List(ctx context.Context, req *pb.LiveRoomListRequest) (*pb.LiveRoomListResponse, error) {
	params := liveSc.LiveRoomListParams_{
		LikeName:   req.Name,
		IsDisabled: int(req.IsDisabled),
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Page:       int(req.Page),
		Limit:      int(req.Limit),
	}

	list, total, err := l.liveRoomSc.List(ctx, params)
	if err != nil {
		l.log.WithContext(ctx).Infof("LiveRoomHttp:List:查询结果失败:Err:%v", err)
		return nil, err
	}

	var listR []*pb.LiveRoomVo
	if err := copier.Copy(&listR, list); err != nil {
		return nil, err
	}

	response := &pb.LiveRoomListResponse{
		Data: &pb.LiveRoomListResponse_Data{
			Total: int32(total),
			Lists: listR,
		},
	}
	return response, nil
}

// Edit 直播间编辑
func (l *LiveRoomHttp) Edit(ctx context.Context, req *pb.LiveRoomEditRequest) (*pb.LiveRoomEditResponse, error) {
	params := liveSc.LiveRoomEditParams_{}

	if err := copier.Copy(&params, req); err != nil {
		return nil, err
	}

	rows, err := l.liveRoomSc.Edit(ctx, params)
	if err != nil {
		l.log.WithContext(ctx).Infof("LiveRoomHttp:Edit:编辑错误:Err:%v", err)
		return nil, err
	}

	return &pb.LiveRoomEditResponse{
		Data: &pb.LiveRoomEditResponse_Data{
			Rows: int32(rows),
		},
	}, nil
}

// Update 直播间更新
func (l *LiveRoomHttp) Update(ctx context.Context, req *pb.LiveRoomUpdateRequest) (*pb.LiveRoomUpdateResponse, error) {
	params := liveSc.LiveRoomUpdateParams_{}

	if err := copier.Copy(&params, req); err != nil {
		return nil, err
	}

	rows, err := l.liveRoomSc.Update(ctx, params)
	if err != nil {
		l.log.WithContext(ctx).Infof("LiveRoomHttp:Update:更新错误:Err:%v", err)
		return nil, err
	}

	return &pb.LiveRoomUpdateResponse{
		Data: &pb.LiveRoomUpdateResponse_Data{
			Rows: int32(rows),
		},
	}, nil
}
