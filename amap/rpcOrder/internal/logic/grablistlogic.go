package logic

// 【第5步·司机拉单】GrabList：从 Redis GEO 抢单池查司机附近的待接单。
//
// 前置条件：司机已登录上线，且 MySQL drivers 表有 current_lng/lat（登录或 map/auth/get/coordinates 上报）。
import (
	"common/config"
	"common/model"
	"common/pool"
	"context"
	"errors"
	"fmt"

	"rpcOrder/internal/svc"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type GrabListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGrabListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrabListLogic {
	return &GrabListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GrabListLogic) GrabList(in *rpcOrder.GrabListReq) (*rpcOrder.GrabListResp, error) {
	if in.DriverId <= 0 {
		return nil, errors.New("司机id无效")
	}

	// 1. 查司机信息与在线状态
	var driverModel model.Driver
	if err := driverModel.DriverModelFindId(config.DB, in.DriverId); err != nil {
		return nil, errors.New("司机信息查询失败")
	}
	if driverModel.OnlineStatus != 1 {
		return nil, errors.New("司机未上线，请先上线")
	}
	if driverModel.CurrentLng == 0 && driverModel.CurrentLat == 0 {
		return nil, errors.New("司机位置未上报，请先开启定位")
	}

	radiusM := in.RadiusM
	limit := int(in.Limit)
	fmt.Printf("[GrabList] driver=%d lng=%v lat=%v radius=%d limit=%d\n",
		in.DriverId, driverModel.CurrentLng, driverModel.CurrentLat, radiusM, limit)

	// 2. GEO 圈选附近待接单
	items, err := pool.GrabListNearby(l.ctx, driverModel.CurrentLng, driverModel.CurrentLat, radiusM, limit)
	if err != nil {
		return nil, err
	}

	// 3. 转换为 proto 列表
	orders := make([]*rpcOrder.GrabOrderItem, 0, len(items))
	for _, item := range items {
		orders = append(orders, &rpcOrder.GrabOrderItem{
			OrderNo:          item.OrderNo,
			UserId:           item.UserId,
			StartLng:         item.StartLng,
			StartLat:         item.StartLat,
			StartAddress:     item.StartAddress,
			EndLng:           item.EndLng,
			EndLat:           item.EndLat,
			EndAddress:       item.EndAddress,
			Distance:         item.Distance,
			Duration:         item.Duration,
			Price:            item.Price,
			ExpiresAt:        item.ExpiresAt,
			DistanceToDriver: item.DistanceToDriver,
		})
	}
	fmt.Printf("[GrabList] driver=%d 返回订单数=%d\n", in.DriverId, len(orders))
	return &rpcOrder.GrabListResp{Orders: orders}, nil
}
