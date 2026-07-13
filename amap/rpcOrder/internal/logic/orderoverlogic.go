package logic

// 【第8步·司机完单】OrderOver：行程结束后的资金结算与状态收尾（须 status=5 用户已上车）。
//
// 在本项目中的动作：事务内扣乘客余额、司机入账 85%、公司账户抽成 15% 与券补贴、
// status→3、乘客/司机 current_lng/lat 同步为订单终点、清 Redis 司机占位、发 order_completed 通知。
import (
	"common/config"
	"common/model"
	"common/pool"
	"context"
	"errors"
	"fmt"
	"time"

	"rpcOrder/internal/svc"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	driverIncomeRate   = 0.85 // 司机入账 85%
	platformIncomeRate = 0.15 // 公司账户平台抽成 15%
)

type OrderOverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrderOverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderOverLogic {
	return &OrderOverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OrderOverLogic) OrderOver(in *rpcOrder.OrderOverReq) (*rpcOrder.OrderOverResp, error) {
	if in.OrderNo == "" || in.DriverId <= 0 {
		return nil, errors.New("参数错误")
	}

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 查询订单并加行锁
	var orderModel model.Order
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("order_no = ?", in.OrderNo).
		First(&orderModel).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("订单不存在")
	}
	if orderModel.DriverId != in.DriverId {
		tx.Rollback()
		return nil, errors.New("你无权完成该订单")
	}
	if orderModel.Status == model.OrderStatusCompleted {
		tx.Rollback()
		// 幂等：若上次完单后 Redis 清理失败，重试时补删占位 key
		if err := pool.ClearDriverOrderKey(in.DriverId); err != nil {
			l.Errorf("[OrderOver] 订单已完成，补清理 Redis 失败 order=%s driver=%d err=%v", in.OrderNo, in.DriverId, err)
		}
		return nil, errors.New("订单已经完成")
	}
	if orderModel.Status != model.OrderStatusOnBoard {
		tx.Rollback()
		return nil, errors.New("订单状态异常，请先确认乘客上车后再完单")
	}

	// 2. 计算优惠后实付金额（并核销优惠券）
	payPrice, err := calcPayPriceInTx(tx, &orderModel)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	couponSubsidy := orderModel.Price - payPrice // 优惠券补贴差额，由公司账户承担

	// 3. 公司账户：优惠券补贴扣减
	if couponSubsidy > 0 {
		if err := deductCompanyBalance(tx, &orderModel, couponSubsidy, payPrice, "公司账户优惠券补贴扣减"); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 4. 扣减乘客余额
	var userModel model.User
	if err := userModel.UserModelFindId(tx, orderModel.UserId); err != nil {
		tx.Rollback()
		return nil, errors.New("用户查找失败")
	}
	if userModel.Balance < payPrice {
		tx.Rollback()
		return nil, errors.New("用户余额不足，无法完成订单")
	}
	beforeUserBalance := userModel.Balance
	userModel.Balance -= payPrice
	applyOrderEndLocation(&userModel, nil, &orderModel)
	if err := userModel.UserModelUpd(tx); err != nil {
		tx.Rollback()
		return nil, errors.New("用户余额扣减失败")
	}

	userLog := model.WalletLog{
		UserId:        int64(userModel.ID),
		UserType:      1,
		OrderNo:       orderModel.OrderNo,
		Amount:        payPrice,
		BalanceBefore: beforeUserBalance,
		BalanceAfter:  userModel.Balance,
		Type:          2, // 消费
		Status:        1,
		Remark:        "用户打车消费扣款",
	}
	if err := userLog.WalletLogModel(tx); err != nil {
		tx.Rollback()
		return nil, errors.New("用户流水记录失败")
	}

	// 5. 司机入账（按订单原价 85%）
	var driverModel model.Driver
	if err := driverModel.DriverModelFindId(tx, orderModel.DriverId); err != nil {
		tx.Rollback()
		return nil, errors.New("司机信息异常")
	}
	driverIncome := orderModel.Price * driverIncomeRate
	beforeDriverBalance := driverModel.Balance
	driverModel.Balance += driverIncome
	driverModel.OrderCount++
	applyOrderEndLocation(nil, &driverModel, &orderModel)
	if err := driverModel.DriverModelUpd(tx); err != nil {
		tx.Rollback()
		return nil, errors.New("司机余额增加失败")
	}

	driverLog := model.WalletLog{
		UserId:        int64(driverModel.ID),
		UserType:      2,
		OrderNo:       orderModel.OrderNo,
		Amount:        driverIncome,
		BalanceBefore: beforeDriverBalance,
		BalanceAfter:  driverModel.Balance,
		Type:          1, // 收入
		Status:        1,
		Remark:        "司机完成订单收入",
	}
	if err := driverLog.WalletLogModel(tx); err != nil {
		tx.Rollback()
		return nil, errors.New("司机流水记录失败")
	}

	// 6. 公司账户：订单平台抽成 15% 入账
	platformIncome := orderModel.Price * platformIncomeRate
	if err := addCompanyBalance(tx, &orderModel, platformIncome, "公司账户订单平台抽成收入"); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 7. 更新订单为已完成（乐观锁）
	result := tx.Model(&orderModel).
		Where("order_no = ? AND status = ?", in.OrderNo, model.OrderStatusOnBoard).
		Updates(map[string]interface{}{
			"status":     model.OrderStatusCompleted,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		tx.Rollback()
		return nil, errors.New("订单状态更新失败")
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, errors.New("订单状态已变更，无法完成")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("事务提交失败")
	}

	// 8. 清理 Redis 司机占位 + 通知乘客
	if err := pool.ClearDriverOrderKey(in.DriverId); err != nil {
		l.Errorf("[OrderOver] 完单成功但清理 Redis 失败 order=%s driver=%d err=%v", in.OrderNo, in.DriverId, err)
	}
	go pool.PublishOrderCompletedNotify(in.OrderNo, orderModel.UserId, in.DriverId, payPrice)
	fmt.Printf("[OrderOver] 完单成功 order=%s driver=%d pay=%.2f\n", in.OrderNo, in.DriverId, payPrice)

	return &rpcOrder.OrderOverResp{
		Status: "订单已完成",
	}, nil
}

// calcPayPriceInTx 根据订单绑定的优惠券计算实付金额，并删除已用券
func calcPayPriceInTx(tx *gorm.DB, order *model.Order) (float64, error) {
	payPrice := order.Price
	if order.FinanceId == 0 {
		return payPrice, nil
	}

	var coupon model.Coupon
	if err := coupon.CouponFindId(tx, int64(order.FinanceId)); err != nil {
		return 0, errors.New("优惠券查询失败")
	}
	switch coupon.Type {
	case 1:
		payPrice = order.Price - coupon.QuanMoney
	case 2:
		payPrice = order.Price * coupon.Discount
	case 3:
		payPrice = 0
	default:
		return 0, errors.New("优惠券类型异常")
	}
	if payPrice < 0 {
		payPrice = 0
	}
	if err := coupon.CouponDel(tx, int64(order.FinanceId)); err != nil {
		return 0, errors.New("优惠券核销失败")
	}
	return payPrice, nil
}

// applyOrderEndLocation 完单后将乘客/司机当前坐标同步为订单终点（终点为 0 时跳过）
func applyOrderEndLocation(user *model.User, driver *model.Driver, order *model.Order) {
	if order == nil || (order.EndLng == 0 && order.EndLat == 0) {
		return
	}
	if user != nil {
		user.CurrentLng = order.EndLng
		user.CurrentLat = order.EndLat
	}
	if driver != nil {
		driver.CurrentLng = order.EndLng
		driver.CurrentLat = order.EndLat
	}
}

// deductCompanyBalance 公司账户扣减余额并记流水（优惠券补贴）
func deductCompanyBalance(tx *gorm.DB, order *model.Order, amount, payPrice float64, remark string) error {
	var company model.User
	if err := company.UserModelFindId(tx, model.CompanyUserID); err != nil {
		return errors.New("公司账户查找失败")
	}
	if company.Balance < amount {
		return errors.New("公司账户余额不足，无法补贴优惠券")
	}
	before := company.Balance
	company.Balance -= amount
	if err := company.UserModelUpd(tx); err != nil {
		return errors.New("公司账户优惠券扣减失败")
	}
	log := model.WalletLog{
		UserId:        model.CompanyUserID,
		UserType:      model.WalletUserTypeCompany,
		OrderNo:       order.OrderNo,
		Amount:        amount,
		BalanceBefore: before,
		BalanceAfter:  company.Balance,
		Type:          2, // 支出
		Status:        1,
		Remark:        fmt.Sprintf("%s(原价%.2f实付%.2f)", remark, order.Price, payPrice),
	}
	if err := log.WalletLogModel(tx); err != nil {
		return errors.New("公司账户优惠券流水记录失败")
	}
	return nil
}

// addCompanyBalance 公司账户增加余额并记流水（平台抽成等收入）
func addCompanyBalance(tx *gorm.DB, order *model.Order, amount float64, remark string) error {
	var company model.User
	if err := company.UserModelFindId(tx, model.CompanyUserID); err != nil {
		return errors.New("公司账户查找失败")
	}
	before := company.Balance
	company.Balance += amount
	if err := company.UserModelUpd(tx); err != nil {
		return errors.New("公司账户入账失败")
	}
	log := model.WalletLog{
		UserId:        model.CompanyUserID,
		UserType:      model.WalletUserTypeCompany,
		OrderNo:       order.OrderNo,
		Amount:        amount,
		BalanceBefore: before,
		BalanceAfter:  company.Balance,
		Type:          1, // 收入
		Status:        1,
		Remark:        remark,
	}
	if err := log.WalletLogModel(tx); err != nil {
		return errors.New("公司账户收入流水记录失败")
	}
	return nil
}
