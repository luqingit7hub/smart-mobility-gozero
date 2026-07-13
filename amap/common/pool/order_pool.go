// Package pool 【第2步·抢单池】Redis 侧的高并发抢单能力。
//
// 在本项目中的作用：
//   - 乘客下单后，订单先进入 Redis 池（PublishOrderToPool），司机通过 GEO 看到附近单（GrabListNearby）
//   - 司机点抢单时，Lua 脚本原子判断「是否还可抢」，成功则写入 Stream（第6步消费落库）
//   - 抢中/取消后从池移除（RemoveFromPool），避免司机继续看到无效单
//
// 为什么用 Redis 而不是直接改 MySQL 抢单：抢单并发高，Redis + Lua 更快且能原子防双抢。
package pool

import (
	"common/config"
	"common/constants"
	"common/model"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// 抢单 Lua 返回码（映射到 GrabOrderResp.Code，gRPC 仍返回 nil error）
const (
	GrabCodeOK       = 0
	GrabCodeTaken    = 1 // 订单已被抢
	GrabCodeBusy     = 2 // 司机有未完成单
	GrabCodeExpired  = 3 // 订单过期或不存在
	GrabCodeOffline  = 4 // 司机不在线（预留）
	GrabCodeSysError = 5
)

// PoolOrderItem GEO 圈选后返回给 GrabList 的订单快照
type PoolOrderItem struct {
	OrderNo          string
	UserId           int64
	StartLng         float64
	StartLat         float64
	StartAddress     string
	EndLng           float64
	EndLat           float64
	EndAddress       string
	Distance         float64
	Duration         int64
	Price            float64
	ExpiresAt        int64
	DistanceToDriver float64 // 司机到订单起点距离（公里）
}

// grabScript 原子抢单：校验 pending → 标记 grabbed → 绑定司机 → XADD 落库事件流。
// XADD 放在 Lua 内，保证「抢中」与「入 Stream」同一原子操作，避免 Lua 外发消息丢事件。
var grabScript = redis.NewScript(`
local orderKey  = KEYS[1]
local driverKey = KEYS[2]
local streamKey = KEYS[3]
local driverId  = ARGV[1]
local now       = tonumber(ARGV[2])

if redis.call('EXISTS', orderKey) == 0 then
    return 'EXPIRED'
end
local status = redis.call('HGET', orderKey, 'status')
if status ~= 'pending' then
    return 'TAKEN'
end
local expiresAt = tonumber(redis.call('HGET', orderKey, 'expires_at') or '0')
if expiresAt > 0 and expiresAt < now then
    return 'EXPIRED'
end
if redis.call('EXISTS', driverKey) == 1 then
    return 'BUSY'
end

local orderNo = redis.call('HGET', orderKey, 'order_no')
redis.call('HSET', orderKey, 'status', 'grabbed', 'winner', driverId, 'accept_at', now)
redis.call('SET', driverKey, orderNo)

redis.call('XADD', streamKey, 'MAXLEN', '~', '100000', '*',
    'order_no', orderNo,
    'driver_id', driverId,
    'accept_at', now)

return 'OK'
`)

// PublishOrderToPool 新订单写入抢单池（GEO + ZSET + Hash）
func PublishOrderToPool(ctx context.Context, o *model.Order) error {
	if o == nil || o.OrderNo == "" {
		return errors.New("订单数据无效")
	}
	expiresAt := time.Now().Unix() + constants.OrderPoolTTL
	cacheKey := constants.OrderCacheKey(o.OrderNo)

	pipe := config.Rdb.TxPipeline()
	pipe.GeoAdd(ctx, constants.OrderWaitingGeoKey, &redis.GeoLocation{
		Name:      o.OrderNo,
		Longitude: o.StartLng,
		Latitude:  o.StartLat,
	})
	pipe.ZAdd(ctx, constants.OrderWaitingZSetKey, redis.Z{
		Score:  float64(expiresAt),
		Member: o.OrderNo,
	})
	pipe.HSet(ctx, cacheKey, map[string]interface{}{
		"order_no":      o.OrderNo,
		"user_id":       o.UserId,
		"start_lng":     o.StartLng,
		"start_lat":     o.StartLat,
		"start_address": o.StartAddress,
		"end_lng":       o.EndLng,
		"end_lat":       o.EndLat,
		"end_address":   o.EndAddress,
		"distance":      o.Distance,
		"duration":      o.Duration,
		"price":         o.Price,
		"status":        constants.PoolStatusPending,
		"expires_at":    expiresAt,
	})
	_, err := pipe.Exec(ctx)
	return err
}

// RemoveFromPool 抢中或取消后从 GEO/ZSET 移除，并删除订单 Hash
func RemoveFromPool(ctx context.Context, orderNo string) {
	if orderNo == "" {
		return
	}
	pipe := config.Rdb.Pipeline()
	// GEO 底层是 ZSET，ZRem 可移除 member
	pipe.ZRem(ctx, constants.OrderWaitingGeoKey, orderNo)
	pipe.ZRem(ctx, constants.OrderWaitingZSetKey, orderNo)
	pipe.Del(ctx, constants.OrderCacheKey(orderNo))
	_, _ = pipe.Exec(ctx)
}

// GrabListNearby 按司机坐标 GEO 圈选附近待接单订单
func GrabListNearby(ctx context.Context, lng, lat float64, radiusM int64, limit int) ([]PoolOrderItem, error) {
	if radiusM <= 0 {
		radiusM = 20000 // 默认 20 公里
	}
	const maxRadiusM int64 = 20000
	if radiusM > maxRadiusM {
		radiusM = maxRadiusM
	}
	if limit <= 0 {
		limit = 20
	}
	if lng == 0 && lat == 0 {
		return nil, errors.New("司机位置未上报，请先开启定位")
	}

	locs, err := config.Rdb.GeoSearchLocation(ctx, constants.OrderWaitingGeoKey, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lng,
			Latitude:   lat,
			Radius:     float64(radiusM),
			RadiusUnit: "m",
			Sort:       "ASC",
			Count:      limit * 2, // 多取一些，过滤已抢/过期后截断
		},
		WithCoord: true,
		WithDist:  true,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("geo search: %w", err)
	}
	if len(locs) == 0 {
		return nil, nil
	}

	pipe := config.Rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, 0, len(locs))
	for _, loc := range locs {
		cmds = append(cmds, pipe.HGetAll(ctx, constants.OrderCacheKey(loc.Name)))
	}
	_, _ = pipe.Exec(ctx)

	now := time.Now().Unix()
	items := make([]PoolOrderItem, 0, limit)
	for i, loc := range locs {
		m, err := cmds[i].Result()
		if err != nil || len(m) == 0 {
			continue
		}
		if m["status"] != constants.PoolStatusPending {
			continue
		}
		expiresAt, _ := strconv.ParseInt(m["expires_at"], 10, 64)
		if expiresAt > 0 && expiresAt < now {
			continue
		}

		userId, _ := strconv.ParseInt(m["user_id"], 10, 64)
		startLng, _ := strconv.ParseFloat(m["start_lng"], 64)
		startLat, _ := strconv.ParseFloat(m["start_lat"], 64)
		endLng, _ := strconv.ParseFloat(m["end_lng"], 64)
		endLat, _ := strconv.ParseFloat(m["end_lat"], 64)
		distance, _ := strconv.ParseFloat(m["distance"], 64)
		duration, _ := strconv.ParseInt(m["duration"], 10, 64)
		price, _ := strconv.ParseFloat(m["price"], 64)

		items = append(items, PoolOrderItem{
			OrderNo:          m["order_no"],
			UserId:           userId,
			StartLng:         startLng,
			StartLat:         startLat,
			StartAddress:     m["start_address"],
			EndLng:           endLng,
			EndLat:           endLat,
			EndAddress:       m["end_address"],
			Distance:         distance,
			Duration:         duration,
			Price:            price,
			ExpiresAt:        expiresAt,
			DistanceToDriver: loc.Dist / 1000, // Redis 返回米，对外用公里
		})
		if len(items) >= limit {
			break
		}
	}
	return items, nil
}

// RunGrabOrder 司机抢单，返回业务码（GrabCode*）与提示文案
func RunGrabOrder(ctx context.Context, orderNo string, driverId int64) (code int, msg string, err error) {
	if orderNo == "" || driverId <= 0 {
		return GrabCodeSysError, "参数错误", nil
	}
	now := time.Now().Unix()
	res, err := grabScript.Run(ctx, config.Rdb,
		[]string{
			constants.OrderCacheKey(orderNo),
			constants.DriverOrderKey(driverId),
			constants.OrderGrabbedStream,
		},
		driverId, now,
	).Result()
	if err != nil {
		return GrabCodeSysError, err.Error(), nil
	}

	switch fmt.Sprintf("%v", res) {
	case "OK":
		RemoveFromPool(ctx, orderNo)
		return GrabCodeOK, "抢单成功", nil
	case "TAKEN":
		return GrabCodeTaken, "订单已被抢", nil
	case "BUSY":
		return GrabCodeBusy, "您有未完成订单", nil
	case "EXPIRED":
		return GrabCodeExpired, "订单已过期", nil
	default:
		return GrabCodeSysError, "未知返回", nil
	}
}

// ClearDriverOrderKey 完单或取消后释放司机占位（amap-lq:driver:order:{driverId}）
func ClearDriverOrderKey(driverId int64) error {
	if driverId <= 0 {
		return nil
	}
	// 使用进程级 context，避免 HTTP/gRPC 请求结束后 ctx 取消导致 Del 失败
	if err := config.Rdb.Del(config.Ctx, constants.DriverOrderKey(driverId)).Err(); err != nil {
		return fmt.Errorf("clear driver order key driver=%d: %w", driverId, err)
	}
	return nil
}
