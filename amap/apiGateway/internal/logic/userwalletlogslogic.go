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

	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserWalletLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserWalletLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserWalletLogsLogic {
	return &UserWalletLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserWalletLogsLogic) UserWalletLogs(req *types.UserWalletLogsReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
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

	data, err := l.svcCtx.RpcUser.ListWalletLogs(l.ctx, &rpcUser.ListWalletLogsReq{
		Uid:      int64(uid),
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

	fmt.Println("用户查流水 token uid:", uid, "total:", data.GetTotal())

	return middleware.SuccessResponse(map[string]interface{}{
		"uid":       uid,
		"list":      list,
		"total":     data.GetTotal(),
		"page":      data.GetPage(),
		"page_size": data.GetPageSize(),
	})
}
