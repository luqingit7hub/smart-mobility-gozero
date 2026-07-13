// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"apiGateway/internal/middleware"
	"context"
	"rpcOrder/rpcOrder"

	"apiGateway/internal/svc"
	"apiGateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderRpcTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderRpcTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderRpcTestLogic {
	return &OrderRpcTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderRpcTestLogic) OrderRpcTest(req *types.UserRpcTestReq) (resp *types.CommonResp, err error) {
	// todo: add your logic here and delete this line
	if data, err := l.svcCtx.RpcOrder.OrderRpcTest(l.ctx, &rpcOrder.OrderRpcTestReq{
		Msg: req.Msg,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
