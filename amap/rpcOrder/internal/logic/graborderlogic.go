package logic

// 【第5步·司机抢单】GrabOrder：Redis Lua 原子抢单，成功则写入 Stream，由第6步异步落 MySQL。
//
// 注意：gRPC 始终返回 nil error，业务成败看响应里的 Code（0 成功，1 已被抢，2 司机忙…）。
import (
	"common/config"
	"common/model"
	"common/pool"
	"context"
	"fmt"

	"rpcOrder/internal/svc"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type GrabOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGrabOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrabOrderLogic {
	return &GrabOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GrabOrderLogic) GrabOrder(in *rpcOrder.GrabOrderReq) (*rpcOrder.GrabOrderResp, error) {
	if in.OrderNo == "" || in.DriverId <= 0 {
		return &rpcOrder.GrabOrderResp{
			Code: pool.GrabCodeSysError,
			Msg:  "参数错误",
		}, nil
	}

	// 1. 校验司机在线
	var driverModel model.Driver
	if err := driverModel.DriverModelFindId(config.DB, in.DriverId); err != nil {
		return &rpcOrder.GrabOrderResp{
			Code: pool.GrabCodeSysError,
			Msg:  "司机信息查询失败",
		}, nil
	}
	if driverModel.OnlineStatus != 1 {
		return &rpcOrder.GrabOrderResp{
			Code: pool.GrabCodeOffline,
			Msg:  "司机未上线，请先上线",
		}, nil
	}

	// 2. MySQL 侧校验司机是否有进行中订单（双保险，Lua 内也会校验 Redis）
	hasOngoing, err := (&model.Order{}).OrderHasOngoing(config.DB, in.DriverId, model.OngoingOrderStatuses)
	if err != nil {
		return &rpcOrder.GrabOrderResp{
			Code: pool.GrabCodeSysError,
			Msg:  "订单状态查询失败",
		}, nil
	}
	if hasOngoing {
		return &rpcOrder.GrabOrderResp{
			Code: pool.GrabCodeBusy,
			Msg:  "您有未完成订单",
		}, nil
	}

	// 3. Lua 原子抢单 + XADD Stream（MySQL 落库由 Stream 消费者异步完成）
	code, msg, _ := pool.RunGrabOrder(l.ctx, in.OrderNo, in.DriverId)
	fmt.Printf("[GrabOrder] driver=%d order=%s code=%d msg=%s\n", in.DriverId, in.OrderNo, code, msg)

	resp := &rpcOrder.GrabOrderResp{
		Code: int32(code),
		Msg:  msg,
	}
	if code == pool.GrabCodeOK {
		resp.OrderNo = in.OrderNo
	}
	//return &rpcOrder.GrabOrderResp{
	//	Code:    0,
	//	Msg:     "",
	//	OrderNo: "",
	//},nil
	return resp, nil
}
