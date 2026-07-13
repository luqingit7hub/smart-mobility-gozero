package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"
	"strings"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetOrderRatingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderRatingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderRatingLogic {
	return &GetOrderRatingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetOrderRating 乘客按订单号查看单条自己提交的评价
func (l *GetOrderRatingLogic) GetOrderRating(in *rpcUser.GetOrderRatingReq) (*rpcUser.GetOrderRatingResp, error) {
	if in.Uid <= 0 {
		return nil, errors.New("用户id无效")
	}
	orderNo := strings.TrimSpace(in.OrderNo)
	if orderNo == "" {
		return nil, errors.New("订单号不能为空")
	}

	var orderModel model.Order
	if err := orderModel.OrderModelFindNumber(config.DB, orderNo); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("订单不存在")
		}
		return nil, errors.New("订单查询失败")
	}
	if orderModel.UserId != in.Uid {
		return nil, errors.New("无权查看该订单评价")
	}

	var ratingRow model.OrderRating
	if err := ratingRow.OrderRatingFindByOrderNo(config.DB, orderNo); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("该订单暂无评价")
		}
		return nil, errors.New("评价查询失败")
	}

	return &rpcUser.GetOrderRatingResp{
		Id:        int64(ratingRow.ID),
		OrderNo:   ratingRow.OrderNo,
		UserId:    ratingRow.UserId,
		DriverId:  ratingRow.DriverId,
		Rating:    ratingRow.Rating,
		Comment:   ratingRow.Comment,
		Tags:      ratingRow.Tags,
		CreatedAt: ratingRow.CreatedAt.Format(timeLayout),
	}, nil
}
