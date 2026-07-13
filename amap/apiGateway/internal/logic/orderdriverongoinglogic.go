// Code scaffolded by goctl. Safe to edit.

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderDriverOngoingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderDriverOngoingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDriverOngoingLogic {
	return &OrderDriverOngoingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderDriverOngoingLogic) OrderDriverOngoing(_ *types.OrderDriverOngoingReq) (resp *types.CommonResp, err error) {
	driverId, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	data, err := l.svcCtx.RpcOrder.DriverOngoing(l.ctx, &rpcOrder.DriverOngoingReq{
		DriverId: int64(driverId),
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
