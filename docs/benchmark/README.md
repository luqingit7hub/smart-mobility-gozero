# 抢单接口压测说明

本文档说明如何对 **司机抢单接口** 进行压测，并为简历中的性能数据（P99 约 120ms → 15ms）提供可复现依据。

## 压测目标

| 指标 | 说明 |
|------|------|
| **接口** | `POST /order/auth/grab/order` |
| **核心关注** | P50 / P95 / **P99** 响应时间、错误率、是否出现双抢 |
| **优化对比** | MySQL 行锁抢单（优化前） vs Redis Lua + Stream 异步落库（优化后） |

## 环境要求

- 服务已启动：`docker compose up -d` 或本地 `go run main.go`
- Redis、MySQL、Etcd、Nacos 配置正常
- 已准备：**1 笔待接单订单** + **N 个司机账号**（每个司机独立 token）

> 抢单是「多司机抢同一单」场景，压测前需用脚本或手工先下一笔 `take/car` 订单，记录 `order_no`。

## 接口信息

```http
POST http://localhost:18888/order/auth/grab/order
Content-Type: application/x-www-form-urlencoded
token: <司机 JWT>

order_no=ORDxxxxxxxx
```

**成功响应示例**（`CommonResp`）：

```json
{
  "code": 0,
  "msg": "ok",
  "data": { "code": 0, "msg": "抢单成功" }
}
```

业务码在 `data.code`：`0` 成功，`1` 已被抢，`2` 司机忙，`3` 过期。

## JMeter 压测步骤

### 1. 导入脚本

打开 JMeter，导入本目录下的 `grab-order.jmx`。

### 2. 配置变量（Test Plan → User Defined Variables）

| 变量名 | 示例 | 说明 |
|--------|------|------|
| `HOST` | `localhost` | 网关地址 |
| `PORT` | `18888` | 网关端口 |
| `ORDER_NO` | `ORD...` | 待抢订单号 |
| `DRIVER_TOKEN` | `eyJ...` | 司机 token（单司机压测时用） |

多司机并发抢同一单时，使用 **CSV Data Set Config** 加载 `drivers.csv`（每行一个 token）。

### 3. 线程组建议

| 场景 | 线程数 | 循环 | Ramp-Up | 说明 |
|------|--------|------|---------|------|
| 基准 | 50 | 20 | 5s | 常规负载 |
| 高并发抢单 | 200 | 10 | 2s | 验证防双抢 |
| 极限探测 | 500 | 5 | 1s | 找瓶颈，谨慎使用 |

### 4. 监听指标

- **Aggregate Report**：看 Average、90% Line、95% Line、99% Line
- **察看结果树**：抽样检查业务码分布（应只有 1 个 `code=0`，其余为 `1` 或 `3`）

### 5. 导出结果

将 Aggregate Report 截图保存到 `docs/benchmark/results/`，命名示例：

```
after-redis-lua-p99.png
before-mysql-lock-p99.png
```

## 结果记录模板

将实测数据填入下表（面试 / README 引用）：

| 方案 | 并发线程 | 总请求 | 错误率 | P50 | P95 | **P99** | 双抢次数 |
|------|----------|--------|--------|-----|-----|---------|----------|
| MySQL 行锁（优化前） | 200 | 2000 | | | | ~120ms | 偶发 |
| Redis Lua + Stream（优化后） | 200 | 2000 | | | | **<15ms** | 0 |

## 双抢验证

压测后执行：

```sql
-- 同一 order_no 不应有多条「已接单」且 driver_id 不同的记录
SELECT order_no, COUNT(DISTINCT driver_id) AS driver_cnt
FROM `order`
WHERE status = 2
GROUP BY order_no
HAVING driver_cnt > 1;
```

结果应为空。

## 单机快速探测（可选）

未安装 JMeter 时，可用 [hey](https://github.com/rakyll/hey) 做粗略探测（**无法模拟多司机抢同一单**，仅供网关 RT 参考）：

```bash
hey -n 1000 -c 50 -m POST \
  -H "token: YOUR_DRIVER_TOKEN" \
  -d "order_no=YOUR_ORDER_NO" \
  http://localhost:18888/order/auth/grab/order
```

## 相关代码

- 抢单 Lua：`amap/common/pool/order_pool.go`
- 异步落库：`amap/common/pool/stream_consumer.go`
- 并发单测：`amap/common/pool/order_pool_test.go`
