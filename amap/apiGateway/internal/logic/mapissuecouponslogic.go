// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"common/model"
	"context"
	"rpcMap/rpcMap"

	"github.com/zeromicro/go-zero/core/logx"
)

type MapIssueCouponsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMapIssueCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MapIssueCouponsLogic {
	return &MapIssueCouponsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MapIssueCoupons 仅 users.id=999 公司账户可调用
func (l *MapIssueCouponsLogic) MapIssueCoupons(req *types.MapIssueCouponsReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if int64(uid) != model.CompanyUserID {
		return middleware.FailResponse("无发券权限，请使用公司账户登录")
	}
	data, err := l.svcCtx.RpcMap.IssueCoupons(l.ctx, &rpcMap.IssueCouponsReq{
		Address:     req.Address,
		Type:        req.Type,
		MoneyQuan:   req.MoneyQuan,
		Discount:    req.Discount,
		OutTime:     req.OutTime,
		OperatorUid: int64(uid),
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
