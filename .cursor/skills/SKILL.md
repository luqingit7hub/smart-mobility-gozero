---
name: amap-lq-dev
description: >-
  amap-lq 网约车项目开发规范：go-zero 微服务目录、goctl 命令（禁止 -m）、
  apiGateway/amap-uni 约定。修改 amap/、amap-uni/、执行 goctl 或新增 RPC/HTTP
  接口前必须阅读本 skill。
---

# amap-lq 团队 AI 协作开发规范

> ## 🛑 硬门禁 — 不读完不许动手
>
> **不读完本文件，不许执行任何写操作**（改 `amap/`、`amap-uni/`、跑 goctl、改 proto/api）。
>
> **不按本文件规范做，不许执行任何写操作。**
>
> 改代码前必须用 Read 工具打开本文件；没 Read = 禁止改文件、禁止 goctl。
>
> **不确定用户意思时，必须先问清楚再行动；禁止乱建新文件**（见 `user-core-principles.mdc` §3）。
>
> ---

> **权威路径（请维护此文件）**：`.cursor/skills/SKILL.md`  
> **Cursor 自动约束**：`user-core-principles.mdc`（`alwaysApply`）§3 要求改代码前 Read 本文  
> **适用对象**：本仓库开发者 + Cursor 任意模型  
> **目的**：统一目录与 goctl 命令，避免乱生成 `client/`、覆盖业务代码。

---

## 1. 使用方式

### Cursor / AI

- `user-core-principles.mdc` §3 要求：改 `amap/`、`amap-uni/` 或跑 goctl 前，必须用 Read 工具阅读本文。
- 默认只分析；仅当用户明确说「帮我改/写/执行」时才改文件或跑命令。

### 队友（非 Cursor）

- 用编辑器直接打开：`.cursor/skills/SKILL.md`（随 Git 提交，与代码同版本）。

---

## 2. 仓库总览

```
amap-lq/
├── .cursor/
│   ├── rules/user-core-principles.mdc  ← alwaysApply（协作 + 安全 + 思考）
│   └── skills/SKILL.md                 ← 本文件（完整规范）
├── amap/                     ← Go 微服务（go-zero）
│   ├── go.work
│   ├── common/
│   ├── apiGateway/           ← HTTP + WS（8888）
│   ├── rpcUser/ rpcDriver/ rpcOrder/ rpcMap/
└── amap-uni/                 ← Vue3 H5 前端
```

联调依赖：MySQL、Redis、RabbitMQ、**Etcd**（`127.0.0.1:2379`）。

---

## 3. 后端目录规范

### 3.1 RPC 正确结构（以 rpcUser 为例）

```
rpcUser/
├── rpcUser.proto
├── rpcuser.go
├── etc/rpcuser.yaml
├── rpcuserclient/rpcuser.go      ← 网关调用，勿改名
├── internal/
│   ├── server/rpcuserserver.go   ← 平铺单文件
│   └── logic/*.go                ← 平铺，一个 RPC 一个 logic
└── rpcUser/*.pb.go               ← protoc 生成，勿手改
```

### 3.2 禁止出现的目录（`goctl -m` 误生成）

| 错误路径 | 处理 |
|----------|------|
| `rpcXxx/client/rpcxxx/` | 删除 |
| `internal/logic/rpcuser/` 等 | 删除空壳 stub |
| `internal/server/rpcuser/` 子目录 | 删除 |

### 3.3 分层

| 模块 | 职责 |
|------|------|
| rpcOrder | 订单状态机：抢单、完单、取消、进行中（**不含**乘客/司机「我的订单列表」） |
| rpcUser | 乘客侧：用户、优惠券、评价、充值、**乘客查流水、乘客查订单列表** |
| rpcDriver | 司机侧：登录、认证、下线、**司机查流水、司机查订单列表** |
| rpcMap | 地理编码、发券 |
| apiGateway | HTTP/WS 入口，**仅**鉴权 + 转发 RPC + `SuccessResponse(data)`（参照 `userlistcouponslogic.go`，**禁止**在网关写业务/map 拼装） |
| common/model | GORM + 乐观锁更新 + 分页查询 |

HTTP：`{ code: 200|400, msg, data }`，Header **`token`**。

---

## 4. 代码生成命令（禁止 `-m`）

### RPC

```bash
cd amap/rpcUser
goctl rpc protoc rpcUser.proto --go_out=./ --go-grpc_out=./ --zrpc_out=.
```

替换服务名即可（rpcDriver / rpcOrder / rpcMap）。**不得加 `-m`。**

生成后人工合并：

1. `internal/logic/xxxlogic.go` 写业务  
2. `internal/server/rpcxxxserver.go` 注册方法  
3. `rpcxxxclient/rpcxxx.go` 增加 client 方法  

### HTTP 网关

```bash
cd amap/apiGateway
goctl api go -api apiGateway.api -dir .
```

已存在的 logic **不会被覆盖**，需手写网关转发（参考 `userlistcouponslogic.go`）。

### 检查清单

- [ ] 无 `client/rpcxxx/`、`internal/logic/rpcxxx/`  
- [ ] `go build .` 通过（rpc + gateway）

---

## 5. 新功能流程

1. `common/model` → 2. `rpcXxx.proto` → 3. goctl（无 `-m`）→ 4. logic  
→ 5. server + client → 6. `apiGateway.api` → 7. goctl api → 8. gateway logic  
→ 9. `amap-uni` → 10. 联调重启服务

---

## 6. 前端 amap-uni

- Vue3 + Vite + Vant + Pinia（**非 uni-app**）
- `VITE_API_BASE_URL=/api`，WS `/ws` 代理 8888
- `postForm` / `uploadForm`，`token` 请求头
- 抢单成功判断：`utils/driverTrip.ts` 的 `isGrabSuccess()`

```bash
cd amap-uni && npm run dev
cd amap-uni && npm run build
```

---

## 7. 订单链路摘要

`1待接单 → 2已接单 → 3已完成 / 4已取消`，更新用 `WHERE status=旧状态`。

完单 WS 事件：`order_completed` → 乘客评价页 `/passenger/rate`。

---

## 8. AI 协作边界

见 `.cursor/rules/user-core-principles.mdc`：未授权不改文件；系统级操作须给回滚与退出步骤。

---

## 9. 常见问题

**goctl 生成了 `client/rpcuser/`？** 用了 `-m`，删掉错误目录，改用无 `-m` 命令。

**api logic 没更新？** `goctl api go` 对已有文件 ignored generation，需手写。

---

## 10. 文档维护

- **只维护本文件** `.cursor/skills/SKILL.md`（规范 + 教训都写在这里，不另建平行 skill 文件）
- 协作原则写 `.cursor/rules/user-core-principles.mdc`
- 勿在文档中写 `config.yaml` 密钥

---

## 11. 开发血泪教训（必读，防重蹈覆辙）

### 11.1 我犯过的错（三句话）

1. **RPC 放错服务**：乘客/司机「查自己的列表」=「查自己的流水」→ `rpcUser` / `rpcDriver`，**不是** `rpcOrder`（rpcOrder 只管询价/下单/抢单/完单/取消/进行中）。
2. **业务写进网关**：在 apiGateway 里 `map` 拼装、分页默认值 → 错；网关只做 token → 调 RPC → `SuccessResponse(data)`，参照 `userlistcouponslogic.go`。
3. **没对照就动手**：说效仿流水，却抄流水网关的 map（特例），没抄主流薄网关；用户质疑时还回避，应直接认错改代码。

### 11.2 效仿现有功能 — 先填表再写码

| 步骤 | 做什么 | 订单列表 | 流水 |
|------|--------|----------|------|
| 1 | HTTP 路由 | `/user/auth/order/list` | `/user/auth/wallet/logs` |
| 2 | 网关 logic 行数/map | ≈ `userlistcouponslogic.go` | 流水网关 map 是**特例**，别当模板 |
| 3 | 调哪个 RPC | `RpcUser.ListOrders` | `RpcUser.ListWalletLogs` |
| 4 | 业务 logic | `rpcUser/internal/logic/listorderslogic.go` | `rpcUser/internal/logic/listwalletlogslogic.go` |
| 5 | 查库 | `common/model/order.go` | `common/model/wallet_log.go` |

**口诀：谁用谁的服务。乘客 rpcUser，司机 rpcDriver，订单状态机 rpcOrder。**

### 11.3 网关 logic 合格线

对照 `amap/apiGateway/internal/logic/userlistcouponslogic.go`：

- `GetTokenUserId` → 组 `XxxReq` → `RpcXxx` → `SuccessResponse(data)`
- **禁止** `map[string]interface{}` 逐字段拷贝（除非用户明确要求）

### 11.4 动手前自检

- [ ] 已 Read 完整本文件（含本节）
- [ ] 用户说效仿某接口 → 已 Grep + Read 全链路
- [ ] RPC 在正确服务（乘客 rpcUser / 司机 rpcDriver）
- [ ] 业务在 `rpcXxx/internal/logic`，不在 apiGateway
- [ ] 网关薄转发，无 map 胖逻辑
- [ ] 无 `goctl -m` 错误目录
- [ ] `go build` 通过
- [ ] **不确定用户意图时已问清楚**；**未擅自新建文件**
