# amap-uni 前端

基于 **Vue3 + Vite + Vant + Pinia** 的网约车 H5 双端（乘客 / 司机 / 公司），接口对接 `apiGateway`（默认 `127.0.0.1:8888`）。

> 项目总览与架构见仓库根目录 [README.md](../README.md)。

## 已实现

| 模块 | 功能 |
|------|------|
| 乘客 | 登录/注册、叫车询价下单、等单（WebSocket）、优惠券、充值、实名、订单评价、AI 助手 |
| 司机 | 登录/注册、设置接单位置、抢单大厅、行程、完单、下线、资质认证、AI 助手 |
| 公司 | uid=999 发放优惠券 |

## 未实现 / 简化

- 省市区级联地区库（改为地址文本 + `/map/auth/get/coordinates`）
- 人脸核验 / OCR / 删号等

## 环境变量

| 文件 | 用途 |
|------|------|
| `.env.example` | 模板 |
| `.env.development` | `npm run dev` 默认 |
| `.env.production` | `npm run build` 默认 |

| 变量 | 说明 |
|------|------|
| `VITE_API_BASE_URL` | API 前缀，开发默认 `/api`（Vite 代理到 8888） |
| `VITE_BAIDU_MAP_AK` | 百度地图 JS API Key |

## 开发

```bash
cd amap-uni
npm install
npm run dev
```

访问 http://localhost:5173 ，`/api` 与 `/ws` 代理到 `8888`。

## 构建

```bash
npm run build        # 含类型检查
npm run build:server # 仅构建，供 Docker / CI 使用
```

Docker 全栈启动时，前端由 `web` 容器（Nginx）在 http://localhost 提供服务。

## 关键目录

```
src/
├── api/           # HTTP 封装
├── stores/        # Pinia 登录态
├── views/         # 乘客 / 司机 / 公司页面
├── components/    # AiAssistant、BookTripMap 等
└── utils/orderWs.ts  # 订单 WebSocket（断线重连）
```
