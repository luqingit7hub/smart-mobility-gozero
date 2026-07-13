package logic

import (
	"context"
	"fmt"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRpcTestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserRpcTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRpcTestLogic {
	return &UserRpcTestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserRpcTestLogic) UserRpcTest(in *rpcUser.UserRpcTestReq) (*rpcUser.UserRpcTestResp, error) {
	// todo: add your logic here and delete this line
	data := fmt.Sprintln("apiGateway传输测试的rpcUser消息:", in.Msg)
	return &rpcUser.UserRpcTestResp{
		Status: data,
	}, nil
}
