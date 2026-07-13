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

type UserAliPayBalanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserAliPayBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserAliPayBalanceLogic {
	return &UserAliPayBalanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserAliPayBalanceLogic) UserAliPayBalance(req *types.UserAliPayBalanceReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if data, err := l.svcCtx.RpcUser.AliPayBalance(l.ctx, &rpcUser.AliPayBalanceReq{
		Price: req.Money,
		Uid:   int64(uid),
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
