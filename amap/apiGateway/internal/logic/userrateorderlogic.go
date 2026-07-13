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

type UserRateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRateOrderLogic {
	return &UserRateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRateOrderLogic) UserRateOrder(req *types.UserRateOrderReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	data, err := l.svcCtx.RpcUser.RateOrder(l.ctx, &rpcUser.RateOrderReq{
		Uid:     int64(uid),
		OrderNo: req.OrderNo,
		Rating:  req.Rating,
		Comment: req.Comment,
		Tags:    req.Tags,
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
