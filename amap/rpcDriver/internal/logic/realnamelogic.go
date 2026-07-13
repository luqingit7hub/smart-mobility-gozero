package logic

import (
	"common/config"
	"common/model"
	"common/pkg"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"rpcDriver/internal/svc"
	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RealNameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRealNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RealNameLogic {
	return &RealNameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RealNameLogic) RealName(in *rpcDriver.RealNameReq) (*rpcDriver.RealNameResp, error) {
	if in.Email == "" {
		return nil, errors.New("邮箱不能为空")
	}
	if in.CarNumber == "" {
		return nil, errors.New("车牌号不能为空")
	}
	if in.CarType == "" {
		return nil, errors.New("车型不能为空")
	}
	if in.CarColor == "" {
		return nil, errors.New("车辆颜色不能为空")
	}
	if in.LicensePhoto == "" {
		return nil, errors.New("驾驶证照片不能为空")
	}
	if in.VehiclePhoto == "" {
		return nil, errors.New("行驶证照片不能为空")
	}
	if in.Avatar == "" {
		return nil, errors.New("头像不能为空")
	}
	if in.LicensePhoto == "" {
		return nil, errors.New("驾驶证不能为空")
	}
	if in.VehiclePhoto == "" {
		return nil, errors.New("行驶证不能为空")
	}
	if in.Uid <= 0 {
		return nil, errors.New("司机id不能为空")
	}
	cardNo := strings.TrimSpace(in.CardNo)
	realName := strings.TrimSpace(in.RealName)
	if cardNo == "" || realName == "" {
		return nil, errors.New("身份证号和真实姓名不能为空")
	}
	if regexp.MustCompile(`^\d{17}[\dXx]$`).MatchString(cardNo) == false {
		return nil, errors.New("身份证号格式不正确")
	}
	var driverModel model.Driver
	if err := driverModel.DriverModelFindId(config.DB, in.Uid); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("司机不存在")
		}
		return nil, errors.New("司机查询失败")
	}
	if driverModel.Status == 1 {
		return nil, errors.New("已完成实名认证")
	}
	if driverModel.Status == 3 {
		return nil, errors.New("账号已禁用")
	}
	if driverModel.Status != 2 {
		return nil, errors.New("账号状态异常，无法实名认证")
	}

	result, err := pkg.RealName(cardNo, realName)
	if err != nil {
		return nil, errors.New("实名认证接口调用失败")
	}
	if result.ErrorCode != 0 || result.Result.Isok == false {
		fmt.Println("实名认证失败:", result.Reason)
		return nil, errors.New("实名认证失败，姓名与身份证号不匹配")
	}
	driverModel.Name = realName
	driverModel.IdCard = cardNo
	driverModel.Email = in.Email
	driverModel.Avatar = in.Avatar
	driverModel.CarNumber = in.CarNumber
	driverModel.CarType = in.CarType
	driverModel.CarColor = in.CarColor
	driverModel.LicensePhoto = in.LicensePhoto
	driverModel.VehiclePhoto = in.VehiclePhoto
	driverModel.Status = 1
	if err := driverModel.DriverModelUpd(config.DB); err != nil {
		return nil, errors.New("实名认证信息保存失败")
	}
	fmt.Println("司机实名认证成功, driver_id:", driverModel.ID)
	return &rpcDriver.RealNameResp{Msg: "实名认证成功"}, nil
}
