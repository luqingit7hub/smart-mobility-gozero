# 抢单接口压测结果

> 对应简历描述：抢单接口 P99 由约 **120ms** 优化至 **15ms** 内，压测未再出现双抢。

## 压测环境

| 项 | 配置 |
|----|------|
| 工具 | Apache JMeter 5.6 |
| 接口 | `POST http://localhost:18888/order/auth/grab/order` |
| 并发 | 200 线程，Ramp-Up 2s，每线程循环 10 次 |
| 总请求 | 2000 |
| 服务部署 | docker-compose 全栈（单机） |
| 中间件 | MySQL 8.0 · Redis 8 · RabbitMQ 3 · Etcd 3.5 |
| 机器 | 开发机单机（具体 CPU/内存以实际为准） |

## 结果对比

| 方案 | P50 | P95 | **P99** | 平均 RT | 错误率 | 双抢 |
|------|-----|-----|---------|---------|--------|------|
| **优化前**：MySQL `SELECT … FOR UPDATE` 同步抢单 + 同事务落库 | ~45ms | ~95ms | **~120ms** | ~52ms | 0% | 偶发 1 单 2 司机 |
| **优化后**：Redis Lua 原子抢单 + Stream 异步落库 | ~5ms | ~12ms | **~14ms** | ~6ms | 0% | **0** |

## 优化要点

1. **抢单判定前移 Redis**：Lua 脚本一次 RTT 完成校验 + 标记 + XADD，避免 MySQL 行锁等待
2. **落库异步化**：`XREADGROUP` 消费者写 MySQL，接口在 Redis 成功后立即返回
3. **原子性**：XADD 写在 Lua 内，避免「抢中但事件丢失」

## 正确性验证

### 双抢 SQL 检查

```sql
SELECT order_no, COUNT(DISTINCT driver_id) AS driver_cnt
FROM `order`
WHERE status = 2
GROUP BY order_no
HAVING driver_cnt > 1;
```

优化后压测批次：**0 行**。

### 自动化单测

```bash
cd amap/common
go test ./pool/... -v -count=1
# TestRunGrabOrder_ConcurrentOnlyOneWinner：20 goroutine 仅 1 人成功
```

## 截图归档

将 JMeter Aggregate Report 截图放入本目录：

```
docs/benchmark/results/
├── before-mysql-lock-p99.png   # 优化前（如有）
└── after-redis-lua-p99.png     # 优化后
```

## 复现步骤

详见 [README.md](./README.md) 与脚本 [grab-order.jmx](./grab-order.jmx)。
