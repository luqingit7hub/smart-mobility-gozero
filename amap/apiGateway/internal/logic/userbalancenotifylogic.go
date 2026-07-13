// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"rpcUser/rpcUser"

	"apiGateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserBalanceNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserBalanceNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserBalanceNotifyLogic {
	return &UserBalanceNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserBalanceNotifyLogic) UserBalanceNotify(form map[string]string) string {
	if data, err := l.svcCtx.RpcUser.BalanceNotify(l.ctx, &rpcUser.BalanceNotifyReq{
		Form: form,
	}); err != nil {
		fmt.Println("rpc回调失败", err)
		return "fail"
	} else {
		return data.Result
	}
}
