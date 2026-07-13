package logic

import (
	"common/pkg"
	"context"
	"errors"

	"rpcMap/internal/svc"
	"rpcMap/rpcMap"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReverseGeocodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReverseGeocodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReverseGeocodeLogic {
	return &ReverseGeocodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReverseGeocodeLogic) ReverseGeocode(in *rpcMap.ReverseGeocodeReq) (*rpcMap.ReverseGeocodeResp, error) {
	if in.Uid <= 0 {
		return nil, errors.New("用户id不能为空")
	}
	address, err := pkg.ReverseGeocode(in.Lng, in.Lat)
	if err != nil {
		return nil, err
	}
	return &rpcMap.ReverseGeocodeResp{
		Address: address,
		Lng:     in.Lng,
		Lat:     in.Lat,
	}, nil
}
