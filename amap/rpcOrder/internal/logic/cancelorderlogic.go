package logic

// CancelOrder 【乘客取消】仅 status=1 待接单可取消，已接单不可取消；具体逻辑在 pool.CancelWaitingOrder。
import (
	"common/pool"
	"context"
	"errors"

	"rpcOrder/internal/svc"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelOrderLogic) CancelOrder(in *rpcOrder.CancelOrderReq) (*rpcOrder.CancelOrderResp, error) {
	if in.OrderNo == "" || in.Uid <= 0 {
		return nil, errors.New("参数错误")
	}

	if err := pool.CancelWaitingOrder(l.ctx, in.OrderNo, in.Uid, in.Reason); err != nil {
		return nil, err
	}

	return &rpcOrder.CancelOrderResp{
		Msg: "订单已取消",
	}, nil
}
