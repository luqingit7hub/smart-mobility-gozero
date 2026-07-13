package logic

import (
	"common/config"
	"common/model"
	"common/pkg"
	"context"
	"errors"
	"fmt"
	"strconv"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
)

type AliPayBalanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAliPayBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AliPayBalanceLogic {
	return &AliPayBalanceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AliPayBalanceLogic) AliPayBalance(in *rpcUser.AliPayBalanceReq) (*rpcUser.AliPayBalanceResp, error) {
	if in.Uid <= 0 {
		return nil, errors.New("用户id不能为空")
	}
	if in.Price <= 0 {
		return nil, errors.New("充值金额必须大于0")
	}
	orderNo := strconv.FormatInt(pkg.GetSnow(), 10)
	alipayUrl := pkg.AliPay(orderNo, in.Price)
	if alipayUrl == "" {
		return nil, errors.New("支付宝链接生成失败")
	}
	walletLogModel := model.WalletLog{
		UserId:        in.Uid,
		UserType:      1,
		OrderNo:       orderNo,
		Amount:        in.Price,
		BalanceBefore: 0,
		BalanceAfter:  0,
		Type:          1,
		Status:        2,
		Remark:        "订单已经添加,暂未支付",
	}
	if err := walletLogModel.WalletLogModel(config.DB); err != nil {
		return nil, errors.New("订单流水添加失败")
	}
	fmt.Println("用户发起充值, uid:", in.Uid, "orderNo:", orderNo)
	return &rpcUser.AliPayBalanceResp{AlipayUrl: alipayUrl}, nil
}
