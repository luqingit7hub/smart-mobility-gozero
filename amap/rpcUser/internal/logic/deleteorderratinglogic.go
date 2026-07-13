package logic

import (
	"common/config"
	"common/model"
	"context"
	"errors"
	"fmt"
	"strings"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteOrderRatingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteOrderRatingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrderRatingLogic {
	return &DeleteOrderRatingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteOrderRating 乘客按订单号删除自己提交的评价，并更新司机均分
func (l *DeleteOrderRatingLogic) DeleteOrderRating(in *rpcUser.DeleteOrderRatingReq) (*rpcUser.DeleteOrderRatingResp, error) {
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
		return nil, errors.New("无权删除该订单评价")
	}
	if orderModel.DriverId <= 0 {
		return nil, errors.New("订单未绑定司机，无法删除评价")
	}

	var ratingRow model.OrderRating
	if err := ratingRow.OrderRatingFindByOrderNo(config.DB, orderNo); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("该订单暂无评价")
		}
		return nil, errors.New("评价查询失败")
	}
	if ratingRow.UserId != in.Uid {
		return nil, errors.New("无权删除该评价")
	}

	tx := config.DB.Begin()
	if err := ratingRow.OrderRatingDelete(tx); err != nil {
		tx.Rollback()
		return nil, errors.New("评价删除失败")
	}

	avg, _, err := model.OrderRatingDriverAvg(tx, orderModel.DriverId)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("司机评分统计失败")
	}
	avg = float64(int(avg*100+0.5)) / 100

	var driverModel model.Driver
	if err := driverModel.DriverModelFindId(tx, orderModel.DriverId); err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("司机不存在")
		}
		return nil, errors.New("司机查询失败")
	}
	driverModel.Rating = avg
	if err := driverModel.DriverModelUpd(tx); err != nil {
		tx.Rollback()
		return nil, errors.New("司机评分更新失败")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("评价删除失败")
	}

	fmt.Printf("[DeleteOrderRating] order=%s user=%d driver=%d avg=%.2f\n",
		orderNo, in.Uid, orderModel.DriverId, avg)

	return &rpcUser.DeleteOrderRatingResp{
		Msg:          "评价已删除",
		DriverRating: avg,
	}, nil
}
