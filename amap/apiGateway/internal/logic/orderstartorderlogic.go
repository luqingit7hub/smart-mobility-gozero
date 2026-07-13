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

type OrderStartOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderStartOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderStartOrderLogic {
	return &OrderStartOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// OrderStartOrder 司机确认乘客上车（token 中为司机 id，需输入乘客手机号后四位）
func (l *OrderStartOrderLogic) OrderStartOrder(req *types.OrderStartOrderReq) (resp *types.CommonResp, err error) {
	driverId, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if data, err := l.svcCtx.RpcOrder.StartOrder(l.ctx, &rpcOrder.StartOrderReq{
		DriverId:  int64(driverId),
		OrderNo:   req.OrderNo,
		PhoneTail: req.PhoneTail,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
