package logic

// 【开始行程】StartOrder：司机输入乘客手机号后四位，确认乘客上车，status 2→5，并通知乘客。
import (
	"common/config"
	"common/model"
	"common/pkg"
	"common/pool"
	"context"
	"errors"
	"fmt"
	"strings"

	"rpcOrder/internal/svc"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartOrderLogic {
	return &StartOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StartOrderLogic) StartOrder(in *rpcOrder.StartOrderReq) (*rpcOrder.StartOrderResp, error) {
	if in.DriverId <= 0 || strings.TrimSpace(in.OrderNo) == "" {
		return nil, errors.New("参数错误")
	}
	phoneTail := strings.TrimSpace(in.PhoneTail)
	if len(phoneTail) != 4 {
		return nil, errors.New("请输入乘客手机号后四位")
	}
	for _, c := range phoneTail {
		if c < '0' || c > '9' {
			return nil, errors.New("手机号后四位只能为数字")
		}
	}

	orderNo := strings.TrimSpace(in.OrderNo)
	var order model.Order
	if err := order.OrderModelFindNumber(config.DB, orderNo); err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.DriverId != in.DriverId {
		return nil, errors.New("你无权操作该订单")
	}

	// 幂等：已上车直接返回成功
	if order.Status == model.OrderStatusOnBoard {
		return &rpcOrder.StartOrderResp{
			Status: "用户已上车",
			Msg:    "行程已开始",
		}, nil
	}
	if order.Status != model.OrderStatusAccepted {
		return nil, errors.New("订单状态异常，仅已接单订单可开始行程")
	}

	var user model.User
	if err := user.UserModelFindId(config.DB, order.UserId); err != nil {
		return nil, errors.New("乘客信息查询失败")
	}
	if !pkg.MatchPhoneTail(user.Phone, phoneTail) {
		return nil, errors.New("手机号后四位与乘客不符，请核对后重试")
	}

	if err := order.OrderUpdateStarted(config.DB, orderNo); err != nil {
		// 并发下可能已被其他请求更新，再查一次做幂等
		var latest model.Order
		if findErr := latest.OrderModelFindNumber(config.DB, orderNo); findErr == nil &&
			latest.Status == model.OrderStatusOnBoard && latest.DriverId == in.DriverId {
			return &rpcOrder.StartOrderResp{
				Status: "用户已上车",
				Msg:    "行程已开始",
			}, nil
		}
		return nil, errors.New("开始行程失败，请稍后重试")
	}

	go pool.PublishTripStartedNotify(orderNo, order.UserId, in.DriverId)
	fmt.Printf("[StartOrder] 行程开始 order=%s driver=%d user=%d\n", orderNo, in.DriverId, order.UserId)

	return &rpcOrder.StartOrderResp{
		Status: "用户已上车",
		Msg:    "行程已开始，请安全驾驶",
	}, nil
}
