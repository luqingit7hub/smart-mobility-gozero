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

type MapReverseGeocodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMapReverseGeocodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MapReverseGeocodeLogic {
	return &MapReverseGeocodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MapReverseGeocodeLogic) MapReverseGeocode(req *types.MapReverseGeocodeReq) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	if data, err := l.svcCtx.RpcMap.ReverseGeocode(l.ctx, &rpcMap.ReverseGeocodeReq{
		Lng: req.Lng,
		Lat: req.Lat,
		Uid: int64(uid),
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}
}
