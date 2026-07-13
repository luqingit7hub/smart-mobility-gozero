package logic

import (
	"common/config"
	"common/pkg"
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"rpcDriver/internal/svc"
	"rpcDriver/rpcDriver"

	"github.com/zeromicro/go-zero/core/logx"
)

type SmsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SmsLogic {
	return &SmsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SmsLogic) Sms(in *rpcDriver.SmsReq) (*rpcDriver.SmsResp, error) {
	if regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(in.Phone) == false {
		return nil, errors.New("手机号格式不正确")
	}
	keyCount := fmt.Sprintf("driver_sms_count_%s", in.Phone)
	getCount, _ := config.Rdb.Get(config.Ctx, keyCount).Result()
	if getCount != "" {
		return nil, errors.New("一分钟只能发送一次短信验证码")
	}
	code, err := pkg.Sms(in.Phone)
	if err != nil {
		l.Errorf("短信发送失败 phone=%s err=%v", in.Phone, err)
		return nil, errors.New("短信验证码发送失败")
	}
	key := fmt.Sprintf("driver_sms_%s", in.Phone)
	config.Rdb.Set(config.Ctx, key, code, time.Minute*3)
	config.Rdb.Set(config.Ctx, keyCount, 1, time.Minute*1)
	return &rpcDriver.SmsResp{Status: "短信验证码发送成功"}, nil
}
