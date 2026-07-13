// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"apiGateway/internal/middleware"
	"context"
	"rpcUser/rpcUser"

	"apiGateway/internal/svc"
	"apiGateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRpcTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRpcTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRpcTestLogic {
	return &UserRpcTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRpcTestLogic) UserRpcTest(req *types.UserRpcTestReq) (resp *types.CommonResp, err error) {
	// todo: add your logic here and delete this line
	if data, err := l.svcCtx.RpcUser.UserRpcTest(l.ctx, &rpcUser.UserRpcTestReq{
		Msg: req.Msg,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
