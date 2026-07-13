// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"apiGateway/internal/svc"
	"apiGateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApiGatewayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApiGatewayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApiGatewayLogic {
	return &ApiGatewayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApiGatewayLogic) ApiGateway(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
