package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"

	"rpcOrder/internal/svc"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DriverOngoingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDriverOngoingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverOngoingLogic {
	return &DriverOngoingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DriverOngoingLogic) DriverOngoing(in *rpcOrder.DriverOngoingReq) (*rpcOrder.DriverOngoingResp, error) {
	if in.DriverId <= 0 {
		return nil, errors.New("司机id不能为空")
	}
	var order model.Order
	err := order.OrderFindOngoingByDriver(config.DB, in.DriverId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &rpcOrder.DriverOngoingResp{HasOrder: false}, nil
		}
		return nil, errors.New("查询进行中订单失败")
	}
	return &rpcOrder.DriverOngoingResp{
		HasOrder: true,
		Order: &rpcOrder.GrabOrderItem{
			OrderNo:      order.OrderNo,
			UserId:       order.UserId,
			StartLng:     order.StartLng,
			StartLat:     order.StartLat,
			StartAddress: order.StartAddress,
			EndLng:       order.EndLng,
			EndLat:       order.EndLat,
			EndAddress:   order.EndAddress,
			Distance:     order.Distance,
			Duration:     int64(order.Duration),
			Price:        order.Price,
			Status:       int64(order.Status),
		},
	}, nil
}
