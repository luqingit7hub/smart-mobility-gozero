package logic

import (
	"common/config"
	"common/model"
	"common/pkg"
	"context"
	"errors"
	"fmt"
	"regexp"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *rpcUser.LoginReq) (*rpcUser.LoginResp, error) {
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
	var userModel model.User
	if err := userModel.UserModelFindPhone(config.DB, in.Phone); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("手机号不存在")
		}
		return nil, errors.New("手机号查询失败")
	}
	if userModel.Status == 3 {
		return nil, errors.New("账号已禁用，无法登录")
	}
	if in.Type == 1 {
		key := fmt.Sprintf("sms_%s", in.Phone)
		smsGet, _ := config.Rdb.Get(config.Ctx, key).Result()
		if smsGet != in.Code {
			return nil, errors.New("验证码有误")
		}
	} else if in.Type == 2 {
		if pkg.Md5Str(in.Password) != userModel.Password {
			return nil, errors.New("密码错误")
		}
	}
	if in.Lng != 0 || in.Lat != 0 {
		userModel.CurrentLng = in.Lng
		userModel.CurrentLat = in.Lat
		if err := userModel.UserModelUpd(config.DB); err != nil {
			return nil, errors.New("用户经纬度修改失败")
		}
	}
	fmt.Println("用户登录成功, uid:", userModel.ID)
	return &rpcUser.LoginResp{Id: int64(userModel.ID)}, nil
}
