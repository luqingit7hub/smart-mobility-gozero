package logic

import (
	"context"

	"rpcMap/internal/svc"
	"rpcMap/rpcMap"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PingLogic) Ping(in *rpcMap.Request) (*rpcMap.Response, error) {
	// todo: add your logic here and delete this line

	return &rpcMap.Response{}, nil
}
