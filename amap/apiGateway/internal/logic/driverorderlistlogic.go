// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"

	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
)

type DriverOrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDriverOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverOrderListLogic {
	return &DriverOrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DriverOrderListLogic) DriverOrderList(req *types.DriverOrderListReq) (resp *types.CommonResp, err error) {
	driverId, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	data, err := l.svcCtx.RpcDriver.ListOrders(l.ctx, &rpcDriver.ListOrdersReq{
		DriverId: int64(driverId),
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		OrderNo:  req.OrderNo,
		Status:   int32(req.Status),
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
