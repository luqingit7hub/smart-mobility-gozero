package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"
	"fmt"
	"strings"

	"rpcDriver/internal/svc"
	"rpcDriver/rpcDriver"

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

func (l *ListOrdersLogic) ListOrders(in *rpcDriver.ListOrdersReq) (*rpcDriver.ListOrdersResp, error) {
	if in.DriverId <= 0 {
		return nil, errors.New("司机id无效")
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
	list, total, err := orderModel.OrderListByDriver(config.DB, in.DriverId, strings.TrimSpace(in.OrderNo), int(in.Status), page, pageSize)
	if err != nil {
		return nil, errors.New("订单查询失败")
	}
	fmt.Println("司机查订单 driverId:", in.DriverId, "total:", total)

	var items []*rpcDriver.OrderListItem
	for _, row := range list {
		items = append(items, toDriverOrderListItem(row))
	}

	return &rpcDriver.ListOrdersResp{
		List:     items,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func toDriverOrderListItem(row model.Order) *rpcDriver.OrderListItem {
	item := &rpcDriver.OrderListItem{
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
