package logic

// 【第4步·乘客下单】TakeCar：网约车链路的起点。
//
// 在本项目中的完整动作：
//  1. 百度地图：地址→坐标、路线规划→价格/距离
//  2. 校验余额与优惠券，写 MySQL 订单（status=1 待接单）
//  3. 发两条 RabbitMQ 延迟消息（3min 推司机 / 4min 自动取消，第7步消费）
//  4. 写入 Redis 抢单池（第2步），供司机 GEO 拉单
import (
	"common/config"
	"common/model"
	"common/pkg"
	"common/pool"
	"common/rmq"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"rpcOrder/internal/svc"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type TakeCarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTakeCarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TakeCarLogic {
	return &TakeCarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TakeCarLogic) TakeCar(in *rpcOrder.TakeCarReq) (*rpcOrder.TakeCarResp, error) {
	// --- 以下步骤对应 apiGateway POST /order/auth/take/car ---

	// 1. 地址解析经纬度
	startingPoint, err := pkg.GetCoordinates(in.StartingPoint)
	if err != nil {
		return nil, errors.New("获取起点经纬度失败")
	}
	beginLng := startingPoint.Lng
	beginLat := startingPoint.Lat
	destination, err := pkg.GetCoordinates(in.Destination)
	if err != nil {
		return nil, errors.New("获取终点经纬度失败")
	}
	destLng := destination.Lng
	destLat := destination.Lat

	// 2. 路线规划：价格、距离、时长、路况
	pathPlan := pkg.PathPlanReq{
		OriginLng:      beginLng,
		OriginLat:      beginLat,
		DestinationLng: destLng,
		DestinationLat: destLat,
	}
	simpleRoute, err := pkg.GetPathPlan(pathPlan)
	if err != nil {
		return nil, errors.New("经纬度转换路线失败")
	}
	if len(simpleRoute) == 0 {
		return nil, errors.New("未获取到可用路线")
	}
	route := simpleRoute[0]
	times := route.Duration
	mi := route.Distance
	price := route.Toll
	trafficStatus := route.TrafficCondition

	orderNo := strconv.Itoa(int(pkg.GetSnow()))

	// 3. 查用户 & 优惠券实付金额
	var userModel model.User
	if err := userModel.UserModelFindId(config.DB, in.Uid); err != nil {
		return nil, errors.New("用户异常,查找失败")
	}
	// 优惠券地区校验用行程起点坐标，与发券地区、实际上车点一致（不用登录时缓存的定位）
	couponPrice, financeId, err := l.calcPayPrice(in.Uid, in.Tid, price, beginLng, beginLat)
	if err != nil {
		return nil, err
	}
	if userModel.Balance < couponPrice {
		return nil, errors.New("用户余额不足,请先充值余额")
	}

	// 4. 写入 MySQL（Price 存原价，FinanceId 存优惠券 id）
	orderData := model.Order{
		OrderNo:      orderNo,
		UserId:       in.Uid,
		DriverId:     0,
		StartLng:     beginLng,
		StartLat:     beginLat,
		StartAddress: in.StartingPoint,
		EndLng:       destLng,
		EndLat:       destLat,
		EndAddress:   in.Destination,
		Distance:     float64(mi),
		Duration:     times,
		Price:        price,
		Status:       model.OrderStatusWaiting,
		AcceptTime:   nil,
		FinanceId:    financeId,
	}
	if err := orderData.OrderAdd(config.DB); err != nil {
		return nil, errors.New("订单添加失败")
	}

	// 5. 延迟队列：6 分钟推司机、10 分钟无人接单取消
	if err := rmq.PublishDelay(&rmq.OrderDelayMsg{
		OrderNo: orderNo,
		Action:  rmq.ActionPushDrivers,
	}, rmq.DelayPushDriversMs); err != nil {
		return nil, errors.New("延迟消息传入rabbitMq失败")
	}
	if err := rmq.PublishDelay(&rmq.OrderDelayMsg{
		OrderNo: orderNo,
		Action:  rmq.ActionCancelOrder,
	}, rmq.DelayCancelOrderMs); err != nil {
		return nil, errors.New("延迟消息传入rabbitMq失败")
	}

	// 6. 写入 Redis 抢单池（失败只打日志，订单已在库）
	if err := pool.PublishOrderToPool(l.ctx, &orderData); err != nil {
		fmt.Printf("[TakeCar] 入抢单池失败 order=%s err=%v\n", orderNo, err)
	}

	fmt.Println("用户下单成功, uid:", userModel.ID, "order:", orderNo)
	return &rpcOrder.TakeCarResp{
		Status:   int64(trafficStatus),
		Price:    float32(couponPrice),
		Distance: float32(mi),
		Duration: int64(times),
		Text:     "正在打车中...",
		OrderNo:  orderNo,
	}, nil
}

// calcPayPrice 计算优惠后实付金额；tid=0 不使用优惠券
func (l *TakeCarLogic) calcPayPrice(uid, tid int64, price float64, userLng, userLat float64) (payPrice float64, financeId int, err error) {
	payPrice = price
	if tid == 0 {
		return payPrice, 0, nil
	}

	var coupon model.Coupon
	if err := coupon.CouponFindId(config.DB, tid); err != nil {
		return 0, 0, errors.New("优惠券id查询失败")
	}
	if coupon.Uid != int(uid) {
		return 0, 0, errors.New("请输入您可用的优惠券信息")
	}
	if coupon.OutTime.Before(time.Now()) {
		return 0, 0, errors.New("该优惠券已经过期")
	}
	adcode, _, err := pkg.GetCityByLocation(userLng, userLat)
	if err != nil {
		return 0, 0, errors.New("获取用户城市编码信息失败")
	}
	if !pkg.MatchCouponCityCode(coupon.CityCode, adcode) {
		return 0, 0, errors.New("您选择的优惠券所在地区不可用")
	}

	switch coupon.Type {
	case 1: // 现金券
		payPrice = price - coupon.QuanMoney
	case 2: // 折扣券
		payPrice = price * coupon.Discount
	case 3: // 免费乘车券
		payPrice = 0
	default:
		return 0, 0, errors.New("优惠券类型异常")
	}
	if payPrice < 0 {
		payPrice = 0
	}
	return payPrice, int(tid), nil
}
