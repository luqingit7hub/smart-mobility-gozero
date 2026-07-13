package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"rpcDriver/internal/svc"
	"rpcDriver/rpcDriver"

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

func driverStatusName(status int) string {
	switch status {
	case 1:
		return "已认证"
	case 2:
		return "未认证"
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

func (l *GetRealNameStatusLogic) GetRealNameStatus(in *rpcDriver.GetRealNameStatusReq) (*rpcDriver.GetRealNameStatusResp, error) {
	if in.DriverId <= 0 {
		return nil, errors.New("司机id无效")
	}
	var driverModel model.Driver
	if err := driverModel.DriverModelFindId(config.DB, in.DriverId); err != nil {
		return nil, errors.New("司机不存在")
	}
	verified := driverModel.Status == 1
	resp := &rpcDriver.GetRealNameStatusResp{
		Verified:   verified,
		Status:     int32(driverModel.Status),
		StatusName: driverStatusName(driverModel.Status),
		Avatar:     driverModel.Avatar,
	}
	if verified {
		resp.RealName = maskRealName(driverModel.Name)
		resp.CarNumber = driverModel.CarNumber
		resp.CarType = driverModel.CarType
		resp.CarColor = driverModel.CarColor
	}
	return resp, nil
}
