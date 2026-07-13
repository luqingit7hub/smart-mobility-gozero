// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcDriver/rpcdriverclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DriverRegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDriverRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverRegisterLogic {
	return &DriverRegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DriverRegisterLogic) DriverRegister(req *types.DriverRegisterReq) (resp *types.CommonResp, err error) {
	if data, err := l.svcCtx.RpcDriver.DriverReg(l.ctx, &rpcdriverclient.DriverRegReq{
		Phone:    req.Phone,
		Password: req.Password,
		Code:     req.Code,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
