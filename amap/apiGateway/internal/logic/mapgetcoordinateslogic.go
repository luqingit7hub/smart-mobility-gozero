// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"context"
	"rpcMap/rpcMap"

	"github.com/zeromicro/go-zero/core/logx"
)

type MapGetCoordinatesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMapGetCoordinatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MapGetCoordinatesLogic {
	return &MapGetCoordinatesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MapGetCoordinatesLogic) MapGetCoordinates(req *types.MapGetCoordinatesReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if data, err := l.svcCtx.RpcMap.GetCoordinates(l.ctx, &rpcMap.GetCoordinatesReq{
		Address: req.Address,
		Uid:     int64(uid),
		Type:    req.Type,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
