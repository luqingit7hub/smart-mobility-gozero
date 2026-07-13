# amap-lq — AI Agent Instructions

跨工具通用入口（Cursor、Copilot、Codex、Claude Code 等）。技术细节以 `.cursor/skills/SKILL.md` 为准。

## 必读顺序

1. **本文件**（团队通用工作流与安全）
2. **协作与安全底线**：`.cursor/rules/user-core-principles.mdc`
3. **改 `amap/`、`amap-uni/`、执行 goctl 或新增 RPC/HTTP 接口前**：Read `.cursor/skills/SKILL.md`

**优先级**：安全与授权 > 项目技术规范（SKILL.md）> 本文件工作流。

---

## 项目是什么

网约车业务 monorepo：Go **go-zero** 微服务（`amap/`）+ Vue3 H5 前端（`amap-uni/`，非 uni-app）。

联调依赖：MySQL、Redis、RabbitMQ、Etcd（`127.0.0.1:2379`）。HTTP 网关 `apiGateway` 监听 **8888**。

---

## 目录结构（关键路径）

```
amap-lq/
├── AGENTS.md                           ← 本文件（跨工具入口）
├── .cursor/
│   ├── rules/user-core-principles.mdc  ← 协作 + 安全（Cursor alwaysApply）
│   └── skills/SKILL.md                 ← go-zero / goctl / 业务规范（权威）
├── amap/                               ← Go 微服务
│   ├── apiGateway/                     ← HTTP + WS 入口
│   ├── rpcUser/ rpcDriver/ rpcOrder/ rpcMap/
│   └── common/
└── amap-uni/                           ← Vue3 + Vite 前端
```

**禁止**在错误位置新建目录（尤其 `goctl -m` 产生的 `client/rpcxxx/` 等），详见 SKILL.md。

---

## 工作流（团队 superpowers 精简版）

| 场景 | 要求 |
|------|------|
| 新功能 / 大改 | 先澄清需求、出设计；用户确认后再写代码 |
| 实现 | 先 Read/Grep 调查；**未授权不改文件、不跑破坏性命令** |
| 后端 | 优先 TDD：先写失败测试，再最小实现，再重构 |
| 排错 | 先复现 → 定位根因 → 验证修复；禁止未验证就宣称修好 |
| 收尾 | 汇报变更清单；仅用户明确要求时才 commit / push |

---

## 测试与临时文件清理（硬规则）

为验证而**临时创建**的代码与编译产物，**测试/验证完成后必须自行删除**，不得遗留在工作区或提交进 Git。

### 必须清理

- 仅为本次验证新建的脚本、草稿、`main.go` 探针、一次性测试文件（**非**正式 `*_test.go` 套件）
- 验证时编译产生的二进制（如 `go build` 生成的 `.exe`、`apiGateway` 可执行文件等）
- 临时目录、`/tmp` 或仓库内临时文件夹
- 前端验证产生的临时 `dist/`、`.vite` 缓存（若仅为测试而生成且非用户要求保留）

### 不得删除

- 用户明确要求保留的文件
- 正式业务代码与**项目既有**测试文件（`internal/logic/*_test.go` 等）
- `goctl` / `protoc` 生成的规范产物（见 SKILL.md）
- 用户已 staged 或明确要求提交的改动

### 执行要求

1. 验证结束（成功或失败）后，**同一轮任务内**执行清理，再向用户汇报结果。
2. 清理后在回复中简要列出已删除路径（一行即可）。
3. 若无法安全判断某文件是否临时物，**先问用户**，不要误删。

---

## 禁止事项

- 禁止 `goctl` 加 **`-m`**
- 禁止未授权改文件、commit、push
- 禁止在 `apiGateway` 写业务逻辑（网关只做鉴权 + RPC 转发）
- 禁止把乘客/司机「我的订单列表」放进 `rpcOrder`（应在 `rpcUser` / `rpcDriver`）
- 禁止在文档或提交中写入密钥、token
- 禁止验证结束后遗留临时测试代码或编译产物（见上一节）

---

## 构建与验证

```bash
# 后端（在对应服务目录）
go build .

# 前端
cd amap-uni && npm run dev
cd amap-uni && npm run build
```

RPC / HTTP 代码生成命令见 `.cursor/skills/SKILL.md` §4。
