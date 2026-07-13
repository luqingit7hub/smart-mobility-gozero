# amap-uni 前端

基于 **Vue3 + Vite + Vant + Pinia**，UI 参考 `amap-ridehailing`，接口对接 `amap-lq` 网关（`127.0.0.1:8888`）。

## 已实现（与后端一致）

| 模块 | 功能 |
|------|------|
| 乘客 | 登录/注册、叫车询价下单、等单(WebSocket)、优惠券、充值、实名 |
| 司机 | 登录/注册、设置接单位置、抢单大厅、完单、下线、资质认证 |
| 公司 | uid=999 发放优惠券 |

## 未实现（后端无接口）

- AI 助手 / 聊天
- 订单评价
- 省市区级联地区库（改为地址文本 + `/map/auth/get/coordinates`）
- 人脸核验 / OCR / 删号等

## 开发

```bash
cd amap-uni
npm install
npm run dev
```

访问 `http://localhost:5173`，`/api` 与 `/ws` 代理到 `8888`。

## 构建

```bash
npm run build:server
```
