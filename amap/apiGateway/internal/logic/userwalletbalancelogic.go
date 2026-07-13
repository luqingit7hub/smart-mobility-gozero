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

type UserWalletBalanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserWalletBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserWalletBalanceLogic {
	return &UserWalletBalanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserWalletBalanceLogic) UserWalletBalance(req *types.UserWalletBalanceReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	data, err := l.svcCtx.RpcUser.GetWalletBalance(l.ctx, &rpcUser.GetWalletBalanceReq{
		Uid: int64(uid),
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
