package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"
	"fmt"

	"rpcDriver/internal/svc"
	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type OfflineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOfflineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OfflineLogic {
	return &OfflineLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OfflineLogic) Offline(in *rpcDriver.DriverOfflineReq) (*rpcDriver.DriverOfflineResp, error) {
	if in.DriverId <= 0 {
		return nil, errors.New("司机id不能为空")
	}
	var orderModel model.Order
	has, err := orderModel.OrderHasOngoing(config.DB, in.DriverId, []int{2})
	if err != nil {
		return nil, errors.New("查询进行中订单失败")
	}
	if has {
		return nil, errors.New("有未完成订单,禁止下线")
	}
	var driverModel model.Driver
	if err := driverModel.DriverModelFindId(config.DB, in.DriverId); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("司机不存在")
		}
		return nil, errors.New("司机查询失败")
	}
	if driverModel.OnlineStatus == 2 {
		return &rpcDriver.DriverOfflineResp{
			Success: true,
			Msg:     "司机已处于离线状态",
		}, nil
	}
	driverModel.OnlineStatus = 2
	if err := driverModel.DriverModelUpd(config.DB); err != nil {
		return nil, errors.New("司机下线失败")
	}
	fmt.Println("司机下线成功, driver_id:", driverModel.ID)
	return &rpcDriver.DriverOfflineResp{
		Success: true,
		Msg:     "司机下线成功",
	}, nil
}
