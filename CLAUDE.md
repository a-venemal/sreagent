# CLAUDE.md — SREAgent

> **v1.10.0** | Go 1.25 + Gin + Vue 3 + MySQL 8 + Redis 7

## 代码约定

**后端分层**: `handler` → `service` → `repository` → `model`（严格单向）

- Handler: `func (h *XxxHandler) Method(c *gin.Context)`，响应用 `Success(c, data)` / `Error(c, err)`
- GetCurrentUserID: `id, ok := h.GetCurrentUserID(c)`（comma-ok 断言）
- RBAC: `adminOnly`(admin) / `manage`(admin,team_lead) / `operate`(admin,team_lead,member)
- 迁移: `internal/pkg/dbmigrate/migrations/{序号}_{描述}.{up|down}.sql`，**单语句**，禁止 SET NAMES
- 加密: AES-256-GCM，`SREAGENT_SECRET_KEY`（64位hex），格式 `enc:<base64(nonce+密文)>`
- 日志: `zap`，goroutine 内用 `zap.Error` 不用 `zap.Fatal`

**前端**: Vue 3 + Naive UI + Pinia + `useCrudModal`/`usePaginatedList` composable

## 目录

```
cmd/server/main.go           # 入口 + DI wiring
internal/
  model/ (22) handler/ (30) service/ (29) repository/ (21)
  engine/ (6)                # 告警引擎：evaluator + rule_eval + suppression + heartbeat + escalation
  middleware/ (3)            # JWT / CORS / Logger
  router/router.go           # 120+ 端点
  pkg/                       # dbmigrate / datasource / lark / redis / errors
web/src/                     # Vue 3 前端
```

## 错误码

`0` 成功 | `10001` 参数错 | `10002` 业务错 | `10200` 权限不足 | `40001` 未授权 | `50001` DB错 | `50003` 外部API错

## 环境变量

`SREAGENT_DATABASE_PASSWORD` / `SREAGENT_REDIS_PASSWORD` / `SREAGENT_JWT_SECRET` / `SREAGENT_SECRET_KEY`

## 开发命令

`make run` | `make dev` | `make test` | `make lint` | `make web-dev` | `make docker-up`

## 数据模型

```
DataSource ─1:N─ AlertRule ─1:N─ AlertEvent ─1:N─ AlertTimeline
Team ─1:N─ TeamMember ─N:1─ User
EscalationPolicy ─1:N─ EscalationStep
NotifyRule / MuteRule / InhibitionRule / SubscribeRule ── match labels → NotifyMedia
```

## 对话规范（自动生效）

1. 用 `file:line` 引用代码，**不要粘贴大段内容**
2. 先方案后代码，确认后再实现
3. 每次只改一个模块
4. 完成后 `go build` 通过 + 更新 CHANGELOG.md
5. 超过 20 轮对话考虑开新会话
