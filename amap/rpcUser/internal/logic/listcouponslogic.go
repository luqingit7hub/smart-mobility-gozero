package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
)

const timeLayout = "2006-01-02 15:04:05"

type ListCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCouponsLogic {
	return &ListCouponsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListCoupons 按用户 id 返回其全部优惠券（不按 type / 城市过滤；GORM 自动排除软删）
func (l *ListCouponsLogic) ListCoupons(in *rpcUser.ListCouponsReq) (*rpcUser.ListCouponsResp, error) {
	if in.Uid <= 0 {
		return nil, errors.New("用户id无效")
	}
	if in.Uid == model.CompanyUserID {
		return nil, errors.New("公司账户无个人优惠券")
	}

	var couponModel model.Coupon
	list, err := couponModel.CouponListByUserID(config.DB, in.Uid)
	if err != nil {
		return nil, errors.New("优惠券查询失败")
	}

	items := make([]*rpcUser.CouponItem, 0, len(list))
	for _, c := range list {
		items = append(items, &rpcUser.CouponItem{
			Id:        int64(c.ID),
			Type:      int64(c.Type),
			MoneyQuan: c.QuanMoney,
			Discount:  c.Discount,
			CityCode:  c.CityCode,
			OutTime:   c.OutTime.Format(timeLayout),
			TypeName:  couponTypeName(c.Type),
		})
	}

	return &rpcUser.ListCouponsResp{
		List:  items,
		Count: int32(len(items)),
	}, nil
}

func couponTypeName(t int) string {
	switch t {
	case 1:
		return "现金券"
	case 2:
		return "折扣券"
	case 3:
		return "免费乘车券"
	default:
		return "未知"
	}
}
