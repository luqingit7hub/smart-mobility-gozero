// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcUser/rpcuserclient"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLoginLogic {
	return &UserLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLoginLogic) UserLogin(req *types.UserLoginReq) (resp *types.CommonResp, err error) {
	if data, err := l.svcCtx.RpcUser.Login(l.ctx, &rpcuserclient.LoginReq{
		Phone:    req.Phone,
		Password: req.Password,
		Code:     req.Code,
		Type:     req.Type,
		Lng:      req.Lng,
		Lat:      req.Lat,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		token, _ := middleware.TokenHandler(strconv.FormatInt(data.Id, 10))
		return middleware.SuccessResponse(token)
	}
}
