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

type OrderGrabListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderGrabListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderGrabListLogic {
	return &OrderGrabListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// OrderGrabList 司机拉取附近可抢订单（token 中为司机 id）
func (l *OrderGrabListLogic) OrderGrabList(req *types.OrderGrabListReq) (resp *types.CommonResp, err error) {
	driverId, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if data, err := l.svcCtx.RpcOrder.GrabList(l.ctx, &rpcOrder.GrabListReq{
		DriverId: int64(driverId),
		RadiusM:  req.Radius,
		Limit:    int32(req.Limit),
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
