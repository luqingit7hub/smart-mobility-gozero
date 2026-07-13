// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"apiGateway/internal/middleware"
	"context"
	"rpcMap/rpcmapclient"

	"apiGateway/internal/svc"
	"apiGateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MapRpcTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMapRpcTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MapRpcTestLogic {
	return &MapRpcTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MapRpcTestLogic) MapRpcTest(req *types.UserRpcTestReq) (resp *types.CommonResp, err error) {
	// todo: add your logic here and delete this line
	if data, err := l.svcCtx.RpcMap.MapRpcTest(l.ctx, &rpcmapclient.MapRpcTestReq{
		Msg: req.Msg,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
