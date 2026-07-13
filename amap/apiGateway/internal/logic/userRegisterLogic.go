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

type UserRegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRegisterLogic {
	return &UserRegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRegisterLogic) UserRegister(req *types.UserRegisterReq) (resp *types.CommonResp, err error) {
	data, err := l.svcCtx.RpcUser.Register(l.ctx, &rpcuserclient.RegisterReq{
		Phone:    req.Phone,
		Password: req.Password,
		Code:     req.Code,
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
