package logic

import (
	"context"
	"fmt"

	"rpcOrder/internal/svc"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderRpcTestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrderRpcTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderRpcTestLogic {
	return &OrderRpcTestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OrderRpcTestLogic) OrderRpcTest(in *rpcOrder.OrderRpcTestReq) (*rpcOrder.OrderRpcTestResp, error) {
	// todo: add your logic here and delete this line
	data := fmt.Sprintln("apiGateway传输测试的rpcOrder消息:", in.Msg)
	return &rpcOrder.OrderRpcTestResp{
		Status: data,
	}, nil
}
