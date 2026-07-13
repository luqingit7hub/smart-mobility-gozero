// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcDriver/rpcdriverclient"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type DriverLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDriverLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverLoginLogic {
	return &DriverLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DriverLoginLogic) DriverLogin(req *types.DriverLoginReq) (resp *types.CommonResp, err error) {
	if data, err := l.svcCtx.RpcDriver.DriverLogin(l.ctx, &rpcdriverclient.DriverLoginReq{
		Phone:    req.Phone,
		Password: req.Password,
		Code:     req.Code,
		Type:     req.Type,
		Lng:      req.Lng,
		Lat:      req.Lat,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		token, _ := middleware.TokenHandler(strconv.FormatInt(data.Id, 10))
		return middleware.SuccessResponse(token)
	}
}
