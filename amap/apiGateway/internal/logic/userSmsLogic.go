// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcUser/rpcuserclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserSmsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserSmsLogic {
	return &UserSmsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserSmsLogic) UserSms(req *types.UserSmsReq) (resp *types.CommonResp, err error) {
	if data, err := l.svcCtx.RpcUser.Sms(l.ctx, &rpcuserclient.SmsReq{
		Phone: req.Phone,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
