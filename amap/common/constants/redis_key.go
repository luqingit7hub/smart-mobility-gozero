// Package constants 【第1步·地基】Redis Key 与抢单池常量。
//
// 在本项目中的作用：所有「待接单订单」在 Redis 里的名字都从这里取，避免 key 冲突。
// 抢单池 = GEO（按起点坐标找附近单）+ ZSET（按过期时间清理）+ Hash（存单详情）。
// Stream = 司机抢单成功后写入的事件队列，供第6步异步落 MySQL。
package constants

import "fmt"

const (
	// RedisPrefix 项目级前缀，格式: amap-lq:{模块}:{业务}:{标识}
	RedisPrefix = "amap-lq"
)

// ========== 抢单池 ==========

const (
	// OrderWaitingGeoKey GEO 池：member=orderNo，坐标=订单起点经纬度
	OrderWaitingGeoKey = RedisPrefix + ":order:geo:waiting"

	// OrderWaitingZSetKey 超时 ZSET：member=orderNo，score=expiresAt(unix秒)
	OrderWaitingZSetKey = RedisPrefix + ":order:zset:waiting"

	// OrderPoolTTL 抢单池默认过期秒数（与 10 分钟无人接单自动取消对齐）
	OrderPoolTTL int64 = 600
)

// ========== 抢单 Stream ==========

const (
	// OrderGrabbedStream 抢单成功事件流，Lua 脚本 XADD 写入
	OrderGrabbedStream = RedisPrefix + ":stream:order:grabbed"

	// OrderGrabbedStreamDLQ 落库失败死信流
	OrderGrabbedStreamDLQ = RedisPrefix + ":stream:order:grabbed:dlq"

	// OrderGrabbedGroup Stream 消费者组名
	OrderGrabbedGroup = "order-db-writer"
)

// ========== Redis Hash 字段名（订单缓存） ==========

const (
	PoolStatusPending = "pending"
	PoolStatusGrabbed = "grabbed"
)

// BuildRedisKey 拼接带项目前缀的 Redis Key，例如 BuildRedisKey("order","cache","123")
func BuildRedisKey(parts ...string) string {
	key := RedisPrefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// OrderCacheKey 单笔订单 Hash：amap-lq:order:cache:{orderNo}
func OrderCacheKey(orderNo string) string {
	return BuildRedisKey("order", "cache", orderNo)
}

// DriverOrderKey 司机当前进行中订单：amap-lq:driver:order:{driverId}
func DriverOrderKey(driverId int64) string {
	return BuildRedisKey("driver", "order", fmt.Sprintf("%d", driverId))
}
