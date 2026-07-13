// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderOverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderOverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderOverLogic {
	return &OrderOverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// OrderOver 司机完成订单（token 中为司机 id）
func (l *OrderOverLogic) OrderOver(req *types.OrderOverReq) (resp *types.CommonResp, err error) {
	driverId, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if data, err := l.svcCtx.RpcOrder.OrderOver(l.ctx, &rpcOrder.OrderOverReq{
		DriverId: int64(driverId),
		OrderNo:  req.OrderNo,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
