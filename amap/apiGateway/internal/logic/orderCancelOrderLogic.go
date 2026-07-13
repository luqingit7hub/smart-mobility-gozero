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

type OrderCancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderCancelOrderLogic {
	return &OrderCancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// OrderCancelOrder 乘客取消订单（仅待接单可取消）
func (l *OrderCancelOrderLogic) OrderCancelOrder(req *types.OrderCancelOrderReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if data, err := l.svcCtx.RpcOrder.CancelOrder(l.ctx, &rpcOrder.CancelOrderReq{
		Uid:     int64(uid),
		OrderNo: req.OrderNo,
		Reason:  req.Reason,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
