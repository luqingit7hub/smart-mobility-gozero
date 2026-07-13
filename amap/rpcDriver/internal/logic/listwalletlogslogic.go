package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"rpcDriver/internal/svc"
	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
)

const timeLayout = "2006-01-02 15:04:05"

type ListWalletLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWalletLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWalletLogsLogic {
	return &ListWalletLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListWalletLogsLogic) ListWalletLogs(in *rpcDriver.ListWalletLogsReq) (*rpcDriver.ListWalletLogsResp, error) {
	if in.DriverId <= 0 {
		return nil, errors.New("司机id无效")
	}

	// 司机查 user_type=2
	userType := model.WalletUserTypeDriver

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

	var logModel model.WalletLog
	list, total, err := logModel.WalletLogListByUser(config.DB, in.DriverId, userType, strings.TrimSpace(in.OrderNo), page, pageSize)
	if err != nil {
		return nil, errors.New("流水查询失败")
	}
	fmt.Println("司机查流水 driverId:", in.DriverId, "user_type:", userType, "total:", total)

	var items []*rpcDriver.WalletLogItem
	for _, row := range list {
		signedAmount, direction := walletLogAmount(row.UserType, row.Type, row.Amount)
		data := &rpcDriver.WalletLogItem{
			Id:            int64(row.ID),
			OrderNo:       row.OrderNo,
			Amount:        row.Amount,
			SignedAmount:  signedAmount,
			Direction:     direction,
			BalanceBefore: row.BalanceBefore,
			BalanceAfter:  row.BalanceAfter,
			Type:          int32(row.Type),
			TypeName:      walletLogTypeName(row.UserType, row.Type),
			Status:        int32(row.Status),
			StatusName:    walletLogStatusName(row.Status),
			Remark:        row.Remark,
			CreatedAt:     row.CreatedAt.Format(timeLayout),
		}
		items = append(items, data)
	}

	return &rpcDriver.ListWalletLogsResp{
		List:     items,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func walletLogTypeName(userType, logType int) string {
	if userType == model.WalletUserTypePassenger {
		switch logType {
		case 1:
			return "充值"
		case 2:
			return "消费"
		case 3:
			return "提现"
		case 4:
			return "退款"
		}
	}
	if userType == model.WalletUserTypeDriver {
		switch logType {
		case 1:
			return "订单收入"
		case 3:
			return "提现"
		case 4:
			return "退款"
		}
	}
	if userType == model.WalletUserTypeCompany {
		switch logType {
		case 1:
			return "平台收入"
		case 2:
			return "优惠券补贴"
		}
	}
	return "其他"
}

func walletLogStatusName(status int) string {
	if status == 1 {
		return "已支付"
	}
	if status == 2 {
		return "待支付"
	}
	return "未知"
}

func walletLogAmount(userType, logType int, amount float64) (signedAmount float64, direction string) {
	abs := math.Abs(amount)
	direction = "in"
	signedAmount = abs

	// 乘客：消费、提现算支出
	if userType == model.WalletUserTypePassenger && (logType == 2 || logType == 3) {
		direction = "out"
		signedAmount = -abs
	}
	// 司机：提现算支出
	if userType == model.WalletUserTypeDriver && logType == 3 {
		direction = "out"
		signedAmount = -abs
	}
	// 公司：消费算支出
	if userType == model.WalletUserTypeCompany && logType == 2 {
		direction = "out"
		signedAmount = -abs
	}
	return
}
