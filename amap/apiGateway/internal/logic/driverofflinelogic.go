// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
)

type DriverOfflineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDriverOfflineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverOfflineLogic {
	return &DriverOfflineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DriverOfflineLogic) DriverOffline(req *types.DriverOfflineReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if data, err := l.svcCtx.RpcDriver.Offline(l.ctx, &rpcDriver.DriverOfflineReq{
		DriverId: int64(uid),
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
