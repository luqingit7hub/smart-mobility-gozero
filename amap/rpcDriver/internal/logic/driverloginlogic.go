package logic

import (
	"common/config"
	"common/model"
	"common/pkg"
	"context"
	"errors"
	"fmt"
	"regexp"

	"rpcDriver/internal/svc"
	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DriverLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDriverLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverLoginLogic {
	return &DriverLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DriverLoginLogic) DriverLogin(in *rpcDriver.DriverLoginReq) (*rpcDriver.DriverLoginResp, error) {
	if regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(in.Phone) == false {
		return nil, errors.New("手机号格式不正确")
	}
	if in.Type == 1 {
		if in.Code == "" {
			return nil, errors.New("验证码不能为空")
		}
	} else if in.Type == 2 {
		if in.Password == "" {
			return nil, errors.New("密码不能为空")
		}
	} else {
		return nil, errors.New("type=1(验证码登录),type=2(密码登录),不可为其他")
	}
	var driverModel model.Driver
	if err := driverModel.DriverModelFindPhone(config.DB, in.Phone); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("手机号不存在")
		}
		return nil, errors.New("手机号查询失败")
	}
	if driverModel.Status == 3 {
		return nil, errors.New("账号已禁用，无法登录")
	}
	if in.Type == 1 {
		key := fmt.Sprintf("driver_sms_%s", in.Phone)
		smsGet, _ := config.Rdb.Get(config.Ctx, key).Result()
		if smsGet != in.Code {
			return nil, errors.New("短信验证码有误")
		}
	} else {
		if pkg.Md5Str(in.Password) != driverModel.Password {
			return nil, errors.New("密码有误")
		}
	}

	driverModel.OnlineStatus = 1
	if in.Lng != 0 || in.Lat != 0 {
		driverModel.CurrentLng = in.Lng
		driverModel.CurrentLat = in.Lat
	}
	if err := driverModel.DriverModelUpd(config.DB); err != nil {
		return nil, errors.New("司机上线失败")
	}
	fmt.Println("司机登录成功, driver_id:", driverModel.ID)
	return &rpcDriver.DriverLoginResp{
		Id:  int64(driverModel.ID),
		Msg: "司机登录成功",
	}, nil
}
