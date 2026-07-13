package logic

import (
	"common/config"
	"common/model"
	"common/pkg"
	"context"
	"errors"
	"fmt"
	"time"

	"rpcMap/internal/svc"
	"rpcMap/rpcMap"

	"github.com/zeromicro/go-zero/core/logx"
)

type IssueCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIssueCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IssueCouponsLogic {
	return &IssueCouponsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// IssueCoupons 地址 → 城市编码 → 匹配同城乘客并写 coupons 表
func (l *IssueCouponsLogic) IssueCoupons(in *rpcMap.IssueCouponsReq) (*rpcMap.IssueCouponsResp, error) {
	if in.OperatorUid != model.CompanyUserID {
		return nil, errors.New("仅公司账户可发放优惠券")
	}
	if in.Address == "" {
		return nil, errors.New("地址不能为空")
	}
	if in.Type < 1 || in.Type > 3 {
		return nil, errors.New("优惠券类型无效(1现金券,2折扣券,3免费乘车券)")
	}
	switch in.Type {
	case 1:
		if in.MoneyQuan <= 0 {
			return nil, errors.New("现金券优惠金额必须大于0")
		}
		if in.Discount != 0 {
			return nil, errors.New("现金券不可设置折扣")
		}
	case 2:
		if in.MoneyQuan != 0 {
			return nil, errors.New("折扣券不可设置现金优惠金额")
		}
		if in.Discount <= 0 || in.Discount >= 1 {
			return nil, errors.New("折扣券折扣系数须在 0~1 之间（如 0.8 表示 8 折）")
		}
	case 3:
		if in.MoneyQuan != 0 || in.Discount != 0 {
			return nil, errors.New("免费乘车券不可设置金额或折扣")
		}
	}
	outTime, err := time.ParseInLocation("2006-01-02 15:04:05", in.OutTime, time.Local)
	if err != nil {
		return nil, errors.New("过期时间格式应为 2006-01-02 15:04:05")
	}
	if !outTime.After(time.Now()) {
		return nil, errors.New("过期时间必须晚于当前时间")
	}

	targetCode, cityName, err := pkg.AdcodeFromAddress(in.Address)
	if err != nil {
		return nil, fmt.Errorf("地址解析城市编码失败: %w", err)
	}

	var userModel model.User
	users, err := userModel.UserModelListCouponTargets(config.DB)
	if err != nil {
		return nil, errors.New("查询用户列表失败")
	}

	var issued, skipped int32
	for _, user := range users {
		if user.CurrentLng != 0 || user.CurrentLat != 0 {
			userCode, _, err := pkg.GetCityByLocation(user.CurrentLng, user.CurrentLat)
			if err != nil {
				logx.Infof("[IssueCoupons] uid=%d 逆地理失败: %v", user.ID, err)
				skipped++
				continue
			}
			if !pkg.MatchCouponCityCode(targetCode, userCode) {
				skipped++
				continue
			}
		}
		// 无定位：仍发区域券（city_code 由地址解析），用券时按乘客当时位置校验
		coupon := model.Coupon{
			Uid:       int(user.ID),
			Type:      int(in.Type),
			QuanMoney: in.MoneyQuan,
			Discount:  in.Discount,
			CityCode:  targetCode,
			OutTime:   outTime,
		}
		if err := coupon.CouponCreate(config.DB); err != nil {
			return nil, errors.New("优惠券写入失败")
		}
		issued++
	}

	if issued == 0 {
		return nil, errors.New("未发放任何优惠券：无符合条件的乘客（请确认存在 status=1 的乘客，且已定位乘客需与目标地区同城）")
	}

	return &rpcMap.IssueCouponsResp{
		Status:       fmt.Sprintf("发送优惠券成功，共 %d 人", issued),
		CityCode:     targetCode,
		CityName:     cityName,
		IssuedCount:  issued,
		SkippedCount: skipped,
	}, nil
}
