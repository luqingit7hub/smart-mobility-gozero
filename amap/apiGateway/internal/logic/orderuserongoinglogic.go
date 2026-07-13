// Code scaffolded by goctl. Safe to edit.

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderUserOngoingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderUserOngoingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderUserOngoingLogic {
	return &OrderUserOngoingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderUserOngoingLogic) OrderUserOngoing(_ *types.OrderUserOngoingReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	data, err := l.svcCtx.RpcOrder.UserOngoing(l.ctx, &rpcOrder.UserOngoingReq{
		Uid: int64(uid),
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
