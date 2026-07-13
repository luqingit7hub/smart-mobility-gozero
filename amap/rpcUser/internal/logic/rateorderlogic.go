package logic

import (
	"common/config"
	"common/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RateOrderLogic {
	return &RateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RateOrder 乘客评价自己的已完成订单，并更新司机均分
func (l *RateOrderLogic) RateOrder(in *rpcUser.RateOrderReq) (*rpcUser.RateOrderResp, error) {
	if in.Uid <= 0 {
		return nil, errors.New("用户id无效")
	}
	orderNo := strings.TrimSpace(in.OrderNo)
	if orderNo == "" {
		return nil, errors.New("订单号不能为空")
	}
	if in.Rating < 1 || in.Rating > 5 {
		return nil, errors.New("评分需在 1-5 星之间")
	}

	comment := strings.TrimSpace(in.Comment)
	if utf8.RuneCountInString(comment) > 500 {
		return nil, errors.New("评价内容不能超过 500 字")
	}

	tags := strings.TrimSpace(in.Tags)
	if tags != "" {
		if !json.Valid([]byte(tags)) {
			return nil, errors.New("标签格式需为 JSON 数组")
		}
		var arr []string
		if err := json.Unmarshal([]byte(tags), &arr); err != nil {
			return nil, errors.New("标签格式需为 JSON 数组")
		}
		if len(arr) > 10 {
			return nil, errors.New("标签最多 10 个")
		}
	}

	var orderModel model.Order
	if err := orderModel.OrderModelFindNumber(config.DB, orderNo); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("订单不存在")
		}
		return nil, errors.New("订单查询失败")
	}
	if orderModel.UserId != in.Uid {
		return nil, errors.New("无权评价该订单")
	}
	if orderModel.Status != model.OrderStatusCompleted {
		return nil, errors.New("仅已完成订单可评价")
	}
	if orderModel.DriverId <= 0 {
		return nil, errors.New("订单未绑定司机，无法评价")
	}

	exists, err := model.OrderRatingExistsByOrderNo(config.DB, orderNo)
	if err != nil {
		return nil, errors.New("评价记录查询失败")
	}
	if exists {
		return nil, model.ErrOrderAlreadyRated
	}

	ratingRow := model.OrderRating{
		OrderNo:  orderNo,
		UserId:   in.Uid,
		DriverId: orderModel.DriverId,
		Rating:   in.Rating,
		Comment:  comment,
		Tags:     tags,
	}

	tx := config.DB.Begin()
	if err := ratingRow.OrderRatingAdd(tx); err != nil {
		tx.Rollback()
		return nil, errors.New("评价保存失败")
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
		return nil, errors.New("评价提交失败")
	}

	fmt.Printf("[RateOrder] order=%s user=%d driver=%d rating=%d avg=%.2f\n",
		orderNo, in.Uid, orderModel.DriverId, in.Rating, avg)

	return &rpcUser.RateOrderResp{
		Msg:          "评价成功",
		RatingId:     int64(ratingRow.ID),
		DriverRating: avg,
	}, nil
}
