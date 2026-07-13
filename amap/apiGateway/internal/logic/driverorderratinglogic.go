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

type DriverOrderRatingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDriverOrderRatingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverOrderRatingLogic {
	return &DriverOrderRatingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DriverOrderRatingLogic) DriverOrderRating(req *types.DriverOrderRatingReq) (resp *types.CommonResp, err error) {
	driverId, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	data, err := l.svcCtx.RpcDriver.GetOrderRating(l.ctx, &rpcDriver.GetOrderRatingReq{
		DriverId: int64(driverId),
		OrderNo:  req.OrderNo,
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
