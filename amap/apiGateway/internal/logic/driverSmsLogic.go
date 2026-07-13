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

type DriverSmsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDriverSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverSmsLogic {
	return &DriverSmsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DriverSmsLogic) DriverSms(req *types.DriverSmsReq) (resp *types.CommonResp, err error) {
	if data, err := l.svcCtx.RpcDriver.Sms(l.ctx, &rpcdriverclient.SmsReq{
		Phone: req.Phone,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
