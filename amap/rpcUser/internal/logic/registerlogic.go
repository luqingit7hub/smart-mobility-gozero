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

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *rpcUser.RegisterReq) (*rpcUser.RegisterResp, error) {
	if regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(in.Phone) == false {
		return nil, errors.New("手机号格式不正确")
	}
	var userModel model.User
	if err := userModel.UserModelFindPhone(config.DB, in.Phone); err == nil {
		return nil, errors.New("一个手机号只能注册一次")
	} else if err != gorm.ErrRecordNotFound {
		return nil, errors.New("手机号查询失败")
	}
	key := fmt.Sprintf("sms_%s", in.Phone)
	smsGet, _ := config.Rdb.Get(config.Ctx, key).Result()
	if smsGet != in.Code {
		return nil, errors.New("验证码有误")
	}
	userData := model.User{
		Phone:    in.Phone,
		Password: pkg.Md5Str(in.Password),
		Status:   2, //未实名
	}
	if err := userData.UserDataRegister(config.DB); err != nil {
		return nil, errors.New("用户注册失败")
	}
	return &rpcUser.RegisterResp{Id: int64(userData.ID)}, nil
}
