package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWalletBalanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWalletBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWalletBalanceLogic {
	return &GetWalletBalanceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetWalletBalanceLogic) GetWalletBalance(in *rpcUser.GetWalletBalanceReq) (*rpcUser.GetWalletBalanceResp, error) {
	if in.Uid <= 0 {
		return nil, errors.New("用户id无效")
	}
	var userModel model.User
	if err := userModel.UserModelFindId(config.DB, in.Uid); err != nil {
		return nil, errors.New("用户不存在")
	}
	return &rpcUser.GetWalletBalanceResp{Balance: userModel.Balance}, nil
}
