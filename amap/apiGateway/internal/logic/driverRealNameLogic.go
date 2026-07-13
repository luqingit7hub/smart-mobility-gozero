// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"
	"common/pkg"
	"context"
	"net/http"
	"rpcDriver/rpcdriverclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DriverRealNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDriverRealNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverRealNameLogic {
	return &DriverRealNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DriverRealNameLogic) DriverRealName(req *types.DriverRealNameReq, r *http.Request) (resp *types.CommonResp, err error) {
	uid, err := middleware.GetTokenUserId(l.ctx)
	if err != nil {
		return middleware.FailResponse(err.Error())
	}
	_, avatar, err := r.FormFile("avatar")
	_, licensePhoto, err := r.FormFile("license_photo")
	_, vehiclePhoto, err := r.FormFile("vehicle_photo")
	avatars, _ := pkg.QiNiuYun(avatar)
	licensePhotos, _ := pkg.QiNiuYun(licensePhoto)
	vehiclePhotos, err := pkg.QiNiuYun(vehiclePhoto)
	if err != nil {
		return middleware.FailResponse(err.Error() + "可能是七牛云到期了,请联系管理员(发布者邮箱lqcomjt@qq.com)")
	}
	if data, err := l.svcCtx.RpcDriver.RealName(l.ctx, &rpcdriverclient.RealNameReq{
		Uid:          int64(uid),
		CardNo:       req.CardNo,
		RealName:     req.RealName,
		Email:        req.Email,
		CarNumber:    req.CarNumber,
		CarType:      req.CarType,
		CarColor:     req.CarColor,
		Avatar:       avatars,
		LicensePhoto: licensePhotos,
		VehiclePhoto: vehiclePhotos,
	}); err != nil {
		return middleware.FailResponse(err.Error())
	} else {
		return middleware.SuccessResponse(data)
	}

}
