package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRealNameStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRealNameStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRealNameStatusLogic {
	return &GetRealNameStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func userStatusName(status int) string {
	switch status {
	case 1:
		return "已实名"
	case 2:
		return "未实名"
	case 3:
		return "已禁用"
	default:
		return "状态异常"
	}
}

func maskRealName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(name)
	if r == utf8.RuneError && size == 0 {
		return ""
	}
	rest := strings.Repeat("*", utf8.RuneCountInString(name)-1)
	return string(r) + rest
}

func (l *GetRealNameStatusLogic) GetRealNameStatus(in *rpcUser.GetRealNameStatusReq) (*rpcUser.GetRealNameStatusResp, error) {
	if in.Uid <= 0 {
		return nil, errors.New("用户id无效")
	}
	var userModel model.User
	if err := userModel.UserModelFindId(config.DB, in.Uid); err != nil {
		return nil, errors.New("用户不存在")
	}
	verified := userModel.Status == 1
	resp := &rpcUser.GetRealNameStatusResp{
		Verified:   verified,
		Status:     int32(userModel.Status),
		StatusName: userStatusName(userModel.Status),
		Nickname:   userModel.Nickname,
		Avatar:     userModel.Avatar,
	}
	if verified {
		resp.RealName = maskRealName(userModel.Name)
	}
	return resp, nil
}
