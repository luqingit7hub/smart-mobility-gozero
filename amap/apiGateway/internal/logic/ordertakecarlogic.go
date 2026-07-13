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

type OrderTakeCarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderTakeCarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderTakeCarLogic {
	return &OrderTakeCarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderTakeCarLogic) OrderTakeCar(req *types.OrderTakeCarReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if data, err := l.svcCtx.RpcOrder.TakeCar(l.ctx, &rpcOrder.TakeCarReq{
		StartingPoint: req.StartingPoint,
		Destination:   req.Destination,
		Uid:           int64(uid),
		Tid:           req.Tid,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
