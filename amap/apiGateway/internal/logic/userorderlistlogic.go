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

type UserOrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserOrderListLogic {
	return &UserOrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserOrderListLogic) UserOrderList(req *types.UserOrderListReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	data, err := l.svcCtx.RpcUser.ListOrders(l.ctx, &rpcUser.ListOrdersReq{
		Uid:      int64(uid),
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		OrderNo:  req.OrderNo,
		Status:   int32(req.Status),
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
