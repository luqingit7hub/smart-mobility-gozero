package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"
	"fmt"
	"strings"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListOrdersLogic) ListOrders(in *rpcUser.ListOrdersReq) (*rpcUser.ListOrdersResp, error) {
	if in.Uid <= 0 {
		return nil, errors.New("用户id无效")
	}

	page := int(in.Page)
	pageSize := int(in.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	var orderModel model.Order
	list, total, err := orderModel.OrderListByUser(config.DB, in.Uid, strings.TrimSpace(in.OrderNo), int(in.Status), page, pageSize)
	if err != nil {
		return nil, errors.New("订单查询失败")
	}
	fmt.Println("乘客查订单 uid:", in.Uid, "total:", total)

	var items []*rpcUser.OrderListItem
	for _, row := range list {
		items = append(items, toUserOrderListItem(row))
	}

	return &rpcUser.ListOrdersResp{
		List:     items,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func toUserOrderListItem(row model.Order) *rpcUser.OrderListItem {
	item := &rpcUser.OrderListItem{
		Id:           int64(row.ID),
		OrderNo:      row.OrderNo,
		UserId:       row.UserId,
		DriverId:     row.DriverId,
		StartAddress: row.StartAddress,
		EndAddress:   row.EndAddress,
		Distance:     row.Distance,
		Duration:     int64(row.Duration),
		Price:        row.Price,
		PayType:      int32(row.PayType),
		Status:       int32(row.Status),
		StatusName:   model.OrderStatusName(row.Status),
		CancelReason: row.CancelReason,
		CreatedAt:    row.CreatedAt.Format(timeLayout),
	}
	if row.AcceptTime != nil {
		item.AcceptTime = row.AcceptTime.Format(timeLayout)
	}
	return item
}
