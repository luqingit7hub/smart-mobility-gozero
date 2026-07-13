// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
)

type DriverRpcTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDriverRpcTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverRpcTestLogic {
	return &DriverRpcTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DriverRpcTestLogic) DriverRpcTest(req *types.UserRpcTestReq) (resp *types.CommonResp, err error) {
	// todo: add your logic here and delete this line
	if data, err := l.svcCtx.RpcDriver.DriverRpcTest(l.ctx, &rpcDriver.DriverRpcTestReq{
		Msg: req.Msg,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
