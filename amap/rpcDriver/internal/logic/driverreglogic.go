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

type DriverRegLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDriverRegLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverRegLogic {
	return &DriverRegLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DriverRegLogic) DriverReg(in *rpcDriver.DriverRegReq) (*rpcDriver.DriverRegResp, error) {
	if regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(in.Phone) == false {
		return nil, errors.New("手机号格式不正确")
	}
	if in.Code == "" {
		return nil, errors.New("验证码不能为空")
	}
	if len(in.Password) < 6 {
		return nil, errors.New("密码长度不能少于6位")
	}
	var driverModel model.Driver
	if err := driverModel.DriverModelFindPhone(config.DB, in.Phone); err == nil {
		return nil, errors.New("一个手机号只能注册一次")
	} else if err != gorm.ErrRecordNotFound {
		return nil, errors.New("手机号查询失败")
	}
	key := fmt.Sprintf("driver_sms_%s", in.Phone)
	smsGet, _ := config.Rdb.Get(config.Ctx, key).Result()
	if smsGet != in.Code {
		return nil, errors.New("验证码有误")
	}
	driverData := model.Driver{
		Phone:    in.Phone,
		Password: pkg.Md5Str(in.Password),
		Status:   2, // 未实名
	}
	if err := driverData.DriverDataRegister(config.DB); err != nil {
		return nil, errors.New("司机注册失败")
	}
	config.Rdb.Del(config.Ctx, key)
	fmt.Println("司机注册成功, phone:", in.Phone)
	return &rpcDriver.DriverRegResp{Msg: "司机注册成功"}, nil
}
