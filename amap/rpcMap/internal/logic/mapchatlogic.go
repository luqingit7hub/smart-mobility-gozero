package logic

import (
	"common/ai"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"rpcDriver/rpcDriver"
	"rpcOrder/rpcOrder"
	"rpcUser/rpcUser"

	"rpcMap/internal/svc"
	"rpcMap/rpcMap"

	"github.com/zeromicro/go-zero/core/logx"
)

type MapChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMapChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MapChatLogic {
	return &MapChatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MapChatLogic) MapChat(in *rpcMap.MapChatReq) (*rpcMap.MapChatResp, error) {
	if in.Question == "" {
		return nil, errors.New("问题不能为空")
	}
	if in.Type != 1 && in.Type != 2 {
		return nil, errors.New("type 只能是 1 或 2")
	}

	var answer string
	var err error

	switch in.Type {
	case 1:
		answer, err = ai.Chat(l.ctx, in.Question)
	case 2:
		if in.Uid <= 0 {
			return nil, errors.New("uid 无效")
		}
		if in.Role != 1 && in.Role != 2 {
			return nil, errors.New("role 无效")
		}
		answer, err = ai.BizChatWithFallback(l.ctx, in.Question, in.Uid, int(in.Role), l)
	}

	if err != nil {
		return nil, err
	}
	return &rpcMap.MapChatResp{Answer: answer}, nil
}

// ---------- ai.BizBackend：RPC 查询由 MapChatLogic 实现 ----------

func (l *MapChatLogic) GetMyBalance(ctx context.Context) (string, error) {
	s, ok := ai.SessionFrom(ctx)
	if !ok {
		return "", fmt.Errorf("未登录")
	}
	switch s.Role {
	case 1:
		resp, err := l.svcCtx.RpcUser.GetWalletBalance(ctx, &rpcUser.GetWalletBalanceReq{Uid: s.UID})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`{"role":"乘客","uid":%d,"balance":%.2f}`, s.UID, resp.Balance), nil
	case 2:
		resp, err := l.svcCtx.RpcDriver.GetWalletBalance(ctx, &rpcDriver.GetWalletBalanceReq{DriverId: s.UID})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`{"role":"司机","driver_id":%d,"balance":%.2f}`, s.UID, resp.Balance), nil
	default:
		return "", fmt.Errorf("未知身份 role=%d", s.Role)
	}
}

func (l *MapChatLogic) ListMyOrders(ctx context.Context, page, pageSize int32) (string, error) {
	s, ok := ai.SessionFrom(ctx)
	if !ok {
		return "", fmt.Errorf("未登录")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 20 {
		pageSize = 20
	}

	var resp interface{}
	var err error
	switch s.Role {
	case 1:
		resp, err = l.svcCtx.RpcUser.ListOrders(ctx, &rpcUser.ListOrdersReq{
			Uid:      s.UID,
			Page:     page,
			PageSize: pageSize,
		})
	case 2:
		resp, err = l.svcCtx.RpcDriver.ListOrders(ctx, &rpcDriver.ListOrdersReq{
			DriverId: s.UID,
			Page:     page,
			PageSize: pageSize,
		})
	default:
		return "", fmt.Errorf("未知身份 role=%d", s.Role)
	}
	if err != nil {
		return "", err
	}
	raw, _ := json.Marshal(resp)
	return string(raw), nil
}

func (l *MapChatLogic) ListCoupons(ctx context.Context, couponType int64) (string, error) {
	s, ok := ai.SessionFrom(ctx)
	if !ok || s.Role != 1 {
		return "", fmt.Errorf("只有乘客才能查优惠券")
	}
	resp, err := l.svcCtx.RpcUser.ListCoupons(ctx, &rpcUser.ListCouponsReq{
		Uid:  s.UID,
		Type: couponType,
	})
	if err != nil {
		return "", err
	}
	raw, _ := json.Marshal(resp)
	return string(raw), nil
}

func (l *MapChatLogic) PreviewJourney(ctx context.Context, start, end string) (string, error) {
	s, ok := ai.SessionFrom(ctx)
	if !ok || s.Role != 1 {
		return "", fmt.Errorf("只有乘客才能估价")
	}
	resp, err := l.svcCtx.RpcOrder.Journey(ctx, &rpcOrder.JourneyReq{
		StartingPoint: start,
		Destination:   end,
		Uid:           s.UID,
	})
	if err != nil {
		return "", err
	}
	raw, _ := json.Marshal(resp)
	return string(raw), nil
}

func (l *MapChatLogic) GrabList(ctx context.Context, radiusM int64, limit int32) (string, error) {
	s, ok := ai.SessionFrom(ctx)
	if !ok || s.Role != 2 {
		return "", fmt.Errorf("只有司机才能看抢单列表")
	}
	if radiusM <= 0 {
		radiusM = 5000
	}
	if limit <= 0 {
		limit = 10
	}
	resp, err := l.svcCtx.RpcOrder.GrabList(ctx, &rpcOrder.GrabListReq{
		DriverId: s.UID,
		RadiusM:  radiusM,
		Limit:    limit,
	})
	if err != nil {
		return "", err
	}
	raw, _ := json.Marshal(resp)
	return string(raw), nil
}
