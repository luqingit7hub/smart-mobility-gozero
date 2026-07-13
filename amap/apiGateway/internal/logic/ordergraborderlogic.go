// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderGrabOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderGrabOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderGrabOrderLogic {
	return &OrderGrabOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// OrderGrabOrder 司机抢单（token 中为司机 id）
func (l *OrderGrabOrderLogic) OrderGrabOrder(req *types.OrderGrabOrderReq) (resp *types.CommonResp, err error) {
	driverId, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if data, err := l.svcCtx.RpcOrder.GrabOrder(l.ctx, &rpcOrder.GrabOrderReq{
		DriverId: int64(driverId),
		OrderNo:  req.OrderNo,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
