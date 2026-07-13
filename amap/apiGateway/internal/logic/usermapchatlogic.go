// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"context"
	"rpcMap/rpcMap"

	"apiGateway/internal/svc"
	"apiGateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserMapChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserMapChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserMapChatLogic {
	return &UserMapChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserMapChatLogic) UserMapChat(req *types.MapChatReq) (resp *types.CommonResp, err error) {
	// todo: add your logic here and delete this line
	if req.Type != 1 && req.Type != 2 {
		return middleware.FailResponse("type 只能是 1 或 2")
	}
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	data, err := l.svcCtx.RpcMap.MapChat(l.ctx, &rpcMap.MapChatReq{
		Question: req.Question,
		Type:     int32(req.Type),
		Uid:      int64(uid),
		Role:     1,
	})
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	return middleware.SuccessResponse(data)
}
