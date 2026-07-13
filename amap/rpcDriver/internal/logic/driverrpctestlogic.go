package logic

import (
	"context"
	"fmt"
	"rpcDriver/internal/svc"
	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
)

type DriverRpcTestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDriverRpcTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverRpcTestLogic {
	return &DriverRpcTestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DriverRpcTestLogic) DriverRpcTest(in *rpcDriver.DriverRpcTestReq) (*rpcDriver.DriverRpcTestResp, error) {
	// todo: add your logic here and delete this line
	data := fmt.Sprintln("apiGateway传输测试的rpcDriver消息:", in.Msg)
	return &rpcDriver.DriverRpcTestResp{
		Status: data,
	}, nil
}
