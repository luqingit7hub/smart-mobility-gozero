// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"fmt"
	"strings"

	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
)

type DriverWalletLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDriverWalletLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverWalletLogsLogic {
	return &DriverWalletLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DriverWalletLogsLogic) DriverWalletLogs(req *types.DriverWalletLogsReq) (resp *types.CommonResp, err error) {
	driverId, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}

	page := int32(req.Page)
	pageSize := int32(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	data, err := l.svcCtx.RpcDriver.ListWalletLogs(l.ctx, &rpcDriver.ListWalletLogsReq{
		DriverId: int64(driverId),
		Page:     page,
		PageSize: pageSize,
		OrderNo:  strings.TrimSpace(req.OrderNo),
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}

	// 转成普通 map，保证 list 一定有数据（protobuf 空数组会被省略）
	list := make([]map[string]interface{}, 0, len(data.GetList()))
	for _, item := range data.GetList() {
		list = append(list, map[string]interface{}{
			"id":             item.GetId(),
			"order_no":       item.GetOrderNo(),
			"amount":         item.GetAmount(),
			"signed_amount":  item.GetSignedAmount(),
			"direction":      item.GetDirection(),
			"balance_before": item.GetBalanceBefore(),
			"balance_after":  item.GetBalanceAfter(),
			"type":           item.GetType(),
			"type_name":      item.GetTypeName(),
			"status":         item.GetStatus(),
			"status_name":    item.GetStatusName(),
			"remark":         item.GetRemark(),
			"created_at":     item.GetCreatedAt(),
		})
	}

	fmt.Println("司机查流水 token driverId:", driverId, "total:", data.GetTotal())

	return middleware.SuccessResponse(map[string]interface{}{
		"driver_id": driverId,
		"list":      list,
		"total":     data.GetTotal(),
		"page":      data.GetPage(),
		"page_size": data.GetPageSize(),
	})
}
