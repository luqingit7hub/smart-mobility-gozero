package logic

import (
	"context"
	"fmt"

	"rpcMap/internal/svc"
	"rpcMap/rpcMap"

	"github.com/zeromicro/go-zero/core/logx"
)

type MapRpcTestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMapRpcTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MapRpcTestLogic {
	return &MapRpcTestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MapRpcTestLogic) MapRpcTest(in *rpcMap.MapRpcTestReq) (*rpcMap.MapRpcTestResp, error) {
	// todo: add your logic here and delete this line
	data := fmt.Sprintln("apiGateway传输测试的rpcMap消息:", in.Msg)
	return &rpcMap.MapRpcTestResp{
		Status: data,
	}, nil
}
