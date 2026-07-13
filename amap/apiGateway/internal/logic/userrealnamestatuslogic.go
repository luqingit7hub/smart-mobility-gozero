// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRealNameStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRealNameStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRealNameStatusLogic {
	return &UserRealNameStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRealNameStatusLogic) UserRealNameStatus(req *types.UserRealNameStatusReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	data, err := l.svcCtx.RpcUser.GetRealNameStatus(l.ctx, &rpcUser.GetRealNameStatusReq{
		Uid: int64(uid),
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
