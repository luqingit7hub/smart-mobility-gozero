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

type UserDeleteOrderRatingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserDeleteOrderRatingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserDeleteOrderRatingLogic {
	return &UserDeleteOrderRatingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserDeleteOrderRatingLogic) UserDeleteOrderRating(req *types.UserOrderRatingReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	data, err := l.svcCtx.RpcUser.DeleteOrderRating(l.ctx, &rpcUser.DeleteOrderRatingReq{
		Uid:     int64(uid),
		OrderNo: req.OrderNo,
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
