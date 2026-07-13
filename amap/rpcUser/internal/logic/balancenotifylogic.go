package logic

import (
	"common/config"
	"common/model"
	"common/pkg"
	"context"
	"fmt"
	"net/url"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
)

type BalanceNotifyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBalanceNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BalanceNotifyLogic {
	return &BalanceNotifyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BalanceNotifyLogic) BalanceNotify(in *rpcUser.BalanceNotifyReq) (*rpcUser.BalanceNotifyResp, error) {
	fmt.Println("异步回调开始")
	form := url.Values{}
	for k, v := range in.Form {
		form.Set(k, v)
	}
	if pkg.AlipayPass(form) == false {
		fmt.Println("验签失败")
		return &rpcUser.BalanceNotifyResp{Result: "fail"}, nil
	}
	fmt.Println("验签成功")
	outTradeNo := in.Form["out_trade_no"]
	if outTradeNo == "" {
		fmt.Println("订单号为空")
		return &rpcUser.BalanceNotifyResp{Result: "fail"}, nil
	}
	tradeStatus := in.Form["trade_status"]
	if tradeStatus != "TRADE_SUCCESS" {
		fmt.Println("状态异常")
		return &rpcUser.BalanceNotifyResp{Result: "fail"}, nil
	}
	fmt.Println("获取的订单号:", outTradeNo)
	var walletLogModel model.WalletLog
	if err := walletLogModel.WalletLogModelFindOrderNo(config.DB, outTradeNo); err != nil {
		fmt.Println("订单异常,查找失败")
		return &rpcUser.BalanceNotifyResp{Result: "fail"}, nil
	}
	if walletLogModel.Status == 1 {
		fmt.Println("订单已处理")
		return &rpcUser.BalanceNotifyResp{Result: "success"}, nil
	}
	var userModel model.User
	if err := userModel.UserModelFindId(config.DB, walletLogModel.UserId); err != nil {
		fmt.Println("用户异常,查找失败")
		return &rpcUser.BalanceNotifyResp{Result: "fail"}, nil
	}
	tx := config.DB.Begin()
	beforeBalance := userModel.Balance
	userModel.Balance += walletLogModel.Amount
	if err := userModel.UserModelUpd(tx); err != nil {
		tx.Rollback()
		fmt.Println("用户余额变动异常")
		return &rpcUser.BalanceNotifyResp{Result: "fail"}, nil
	}
	walletLogModel.BalanceBefore = beforeBalance
	walletLogModel.BalanceAfter = userModel.Balance
	walletLogModel.Status = 1
	walletLogModel.Remark = "用户充值成功"
	if err := walletLogModel.WalletLogModelUpd(tx); err != nil {
		tx.Rollback()
		fmt.Println("流水表修改失败")
		return &rpcUser.BalanceNotifyResp{Result: "fail"}, nil
	}
	if err := tx.Commit().Error; err != nil {
		fmt.Println("事务提交失败")
		return &rpcUser.BalanceNotifyResp{Result: "fail"}, nil
	}
	fmt.Println("用户充值成功")
	return &rpcUser.BalanceNotifyResp{Result: "success"}, nil
}
