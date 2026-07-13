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

type UserOngoingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserOngoingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserOngoingLogic {
	return &UserOngoingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserOngoingLogic) UserOngoing(in *rpcOrder.UserOngoingReq) (*rpcOrder.UserOngoingResp, error) {
	if in.Uid <= 0 {
		return nil, errors.New("用户id不能为空")
	}

	var order model.Order
	if err := order.OrderFindOngoingByUser(config.DB, in.Uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &rpcOrder.UserOngoingResp{HasOrder: false}, nil
		}
		return nil, errors.New("查询进行中订单失败")
	}

	item := &rpcOrder.UserOngoingItem{
		OrderNo:      order.OrderNo,
		Status:       int64(order.Status),
		StartAddress: order.StartAddress,
		EndAddress:   order.EndAddress,
		Distance:     order.Distance,
		Duration:     int64(order.Duration),
		Price:        order.Price,
	}

	if order.DriverId > 0 {
		item.DriverId = order.DriverId
		var driver model.Driver
		if err := driver.DriverModelFindId(config.DB, order.DriverId); err == nil {
			item.DriverName = driver.Name
			item.CarNumber = driver.CarNumber
			item.CarType = driver.CarType
			item.DriverRating = driver.Rating
		}
	}
	if order.AcceptTime != nil {
		item.AcceptAt = order.AcceptTime.Unix()
	}

	return &rpcOrder.UserOngoingResp{
		HasOrder: true,
		Order:    item,
	}, nil
}
