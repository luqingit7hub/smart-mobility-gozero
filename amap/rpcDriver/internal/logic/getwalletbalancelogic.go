package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"

	"rpcDriver/internal/svc"
	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWalletBalanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWalletBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWalletBalanceLogic {
	return &GetWalletBalanceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetWalletBalanceLogic) GetWalletBalance(in *rpcDriver.GetWalletBalanceReq) (*rpcDriver.GetWalletBalanceResp, error) {
	if in.DriverId <= 0 {
		return nil, errors.New("司机id无效")
	}
	var driverModel model.Driver
	if err := driverModel.DriverModelFindId(config.DB, in.DriverId); err != nil {
		return nil, errors.New("司机不存在")
	}
	return &rpcDriver.GetWalletBalanceResp{Balance: driverModel.Balance}, nil
}
