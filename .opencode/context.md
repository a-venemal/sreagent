# SREAgent — 持久会话上下文

> 本文件由 OpenCode 自动维护，供未来会话快速恢复上下文，无需重新探索代码库。
> 最后更新：2026-04-04（Phase 8 — 全部 8 阶段完成后的最终状态）

---

## 项目概览

**模块路径**：`github.com/sreagent/sreagent`
**Go 版本**：1.25.0（go.mod，运行时 1.25.5）
**前端**：Vue 3 + TypeScript + Naive UI + Vite 6
**数据库**：MySQL 8.0（GORM v2 + golang-migrate）
**缓存/状态**：Redis 7（引擎状态持久化 + 节流 + Stream）
**认证**：JWT HS256 + 可选 Keycloak OIDC（go-oidc/v3）
**定位**：面向 SRE/运维团队的智能告警管理平台（告警生命周期 + OnCall + AI 分析 + 飞书集成）

---

## 完成的阶段（全部 8 个阶段已完成）

| 阶段 | 内容 | 状态 |
|------|------|------|
| Phase 0 | Cleanup（删除遗留文件、修复 Dockerfile/K8s 配置、清理前端） | ✅ |
| Phase 1 | CI/CD 完整文档（`docs/ci-deploy.md`） | ✅ |
| Phase 2 | Redis 引擎状态持久化（StateStore 接口 + Redis Hash 实现） | ✅ |
| Phase 3 | Keycloak OIDC + RBAC 权限（RequireRole 应用到所有路由） | ✅ |
| Phase 4 | 核心模块补完（Subscribe/Notify 管道接入、AlertRuleHistory） | ✅ |
| Phase 5 | 前端 UI 全面改版（7 个子阶段：组件提取、composable、OIDC、页面拆分、RBAC UI） | ✅ |
| Phase 6 | API 文档（`docs/api.md`，~87 个端点） | ✅ |
| Phase 7 | QA 多角色验证（14 个后端 + 16 个前端问题已修复） | ✅ |
| Phase 8 | 上下文压缩 + 文档更新（本文件） | ✅ |

---

## 架构决策记录（ADR）

### ADR-1：AI 和 Lark Bot 配置从静态文件移入数据库
- 新增 `system_settings` 表（`000002_system_settings` 迁移），`SystemSettingService` 管理读写
- `sync.RWMutex` + 30s TTL 内存缓存，写操作立即失效
- Sentinel 模式：空字符串的敏感字段不会覆盖现有值

### ADR-2：删除 docker-compose 文件
- 部署目标是 Kubernetes，`docker-compose.yml` 和 `deploy/docker-compose.yml` 均已删除

### ADR-3：K8s Secret 保留四项
- `db-password`、`redis-password`、`jwt-secret`、`secret-key`（AES-GCM 主密钥）
- AI/Lark 凭据加密后存于 DB，不出现在 Secret/ConfigMap 中

### ADR-4：golang-migrate 是 Schema 唯一真实来源
- GORM `AutoMigrate` 仅作安全网，主迁移通过嵌入 SQL 文件执行

### ADR-5：AES-256-GCM 加密敏感配置字段
- 环境变量：`SREAGENT_SECRET_KEY`（64 位十六进制 = 32 字节）
- 加密字段：`ai.api_key`、`lark.app_secret`、`lark.verification_token`、`lark.encrypt_key`
- 存储格式：`enc:<base64(12字节nonce + GCM密文)>`
- 无密钥时明文存储 + WARN 日志

### ADR-6：多数据源规则评估路由
- Prometheus/VictoriaMetrics（PromQL）、Zabbix（JSON-RPC）、VictoriaLogs（LogsQL）
- `rule_eval.go:executeQuery()` 按 `datasource.Type` 分发

### ADR-7：Redis 引擎状态持久化
- `StateStore` 接口 + `StateEntry` 结构体（含 Labels、Annotations、Status、时间戳、EventID）
- Redis Hash 实现：`engine:state:{ruleID}` → fingerprint → JSON
- 所有状态转换点（pending/firing/resolved/recovery_hold/nodata）均调用 persist
- Redis 不可用时优雅降级

### ADR-8：Keycloak OIDC 可选集成
- `OIDCConfig` 在 config.go 中，`cfg.OIDC.Enabled` 控制启用
- Authorization Code Flow → ID Token 验证 → 可配置的 role claim path → 自动创建/更新用户
- CSRF state cookie 验证，Secure flag 从 TLS/X-Forwarded-Proto 推导
- JWT token 通过 URL fragment 传回前端（非 query param，避免 Referer 泄露）

### ADR-9：RBAC 三级权限方案
- `adminOnly`：admin
- `manage`：admin, team_lead
- `operate`：admin, team_lead, member
- `authenticated`：任何登录用户（viewer、global_viewer 可读）
- GET 端点仅需 JWT，写操作需 RequireRole
- Webhook 端点（`/webhooks/alertmanager`、`/lark/event`）无认证

---

## 目录结构

```
cmd/server/main.go              # 入口（~450 行），手动 DI wiring
internal/
  config/config.go              # Viper 配置 + OIDCConfig
  model/                        # GORM 模型
  handler/                      # Gin Handler（含 oidc.go）
  service/                      # 业务逻辑（含 oidc.go ~383 行）
  repository/                   # 数据访问层（含 alert_rule_history.go）
  middleware/auth.go            # JWT + RequireRole（comma-ok 安全断言）
  router/router.go              # 路由注册 + RequireRole 应用
  pkg/
    dbmigrate/                  # golang-migrate runner + embed SQL
    datasource/                 # Prom/VM/VLogs/Zabbix 客户端
    lark/                       # 飞书 Webhook + 卡片模板
    redis/                      # Redis Client + RedisStateStore
    errors/                     # 结构化错误码
  engine/
    evaluator.go                # AlertEvaluator + RuleEvaluator（~318 行）
    rule_eval.go                # 状态机 + 持久化调用
    state_store.go              # StateStore 接口 + StateEntry
    suppression.go              # LevelSuppressor
    escalation_executor.go      # EscalationExecutor（~260 行）
web/src/
  api/                          # request.ts（401 拦截 + Vue Router 跳转）、index.ts
  components/common/            # KVEditor、PageHeader、SeverityTag、StatusTag
  composables/                  # useCrudModal、usePaginatedList
  pages/
    settings/                   # Index.vue + 6 个子组件
    schedule/                   # Index.vue + 4 个子组件
    alerts/                     # rules/ events/ history/ mute/
    notification/               # Rules、Media、Templates、Subscribe、AlertChannels
    dashboard/ datasources/
  stores/auth.ts                # token/user/role + canManage/canOperate + role 持久化
  router/index.ts               # OIDC hash fragment 拦截 + role guard
  i18n/                         # zh-CN.ts (~760 行) + en.ts (~743 行)，locale 持久化
  utils/                        # alert.ts、format.ts
  styles/global.css             # CSS 变量、AI 风格基础样式
  types/index.ts                # TypeScript 接口定义（~390 行）
deploy/
  docker/Dockerfile             # 多阶段构建（Go 1.25 + Node 20 → Alpine）
  kubernetes/                   # namespace / mysql / redis / app
docs/
  architecture.md               # 架构概览
  ci-deploy.md                  # CI/CD 完整文档
  api.md                        # REST API 参考（~87 个端点）
  product-design.md             # 产品设计文档
```

---

## 认证/RBAC 最终状态

- **本地认证**：JWT HS256，bcrypt 密码，24h 过期，无 refresh token
- **OIDC 认证**：可选 Keycloak 集成（Authorization Code Flow + PKCE-ready）
  - 4 个端点：`/auth/oidc/login`、`/auth/oidc/callback`、`/auth/oidc/token`、`/auth/oidc/config`
  - 自动用户创建/更新，可配置 role claim 映射
- **5 角色**：`admin`、`team_lead`、`member`、`viewer`、`global_viewer`
- **RequireRole 已应用到所有路由**（Phase 3 完成）
  - 安全的 comma-ok 类型断言（Phase 7 修复）
- **默认管理员**：`admin/admin123`（首次启动种子）
- **前端 RBAC**：`canManage`/`canOperate` 计算属性控制按钮/菜单可见性

---

## 告警引擎最终状态

- **AlertEvaluator** 管理 RuleEvaluator goroutine 池（一个规则一个协程）
- **状态机**：inactive → pending → firing → recovery_hold → resolved / nodata
- **Redis 持久化**：StateStore 接口，Redis Hash 实现（`engine:state:{ruleID}`）
- **LevelSuppressor**：基于严重级别的去重
- **EscalationExecutor**：独立 goroutine，定时检查并执行升级策略
- **Mute Rules**：`IsAlertMuted()` 在 `SetOnAlert` 回调中调用，通知前拦截
- **通知管道完整链路**：
  - `RouteAlert` → v1 策略管道 → `processSubscriptions()` → `FindSubscriptions` → `ProcessEvent`
  - 支持 lark_webhook / lark_bot / email / custom_webhook

---

## 通知渠道配置格式

### lark_webhook / lark_bot
```json
{"webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"}
```

### email
```json
{
  "smtp_host": "smtp.example.com",
  "smtp_port": 587,
  "smtp_tls": true,
  "username": "alert@example.com",
  "password": "secret",
  "from": "SREAgent <alert@example.com>",
  "recipients": ["ops@example.com"]
}
```

### custom_webhook
```json
{
  "url": "https://hooks.example.com/alert",
  "method": "POST",
  "headers": {"Authorization": "Bearer xxx"},
  "timeout_seconds": 10
}
```

---

## 错误码约定

| 错误码 | 含义 |
|--------|------|
| `0` | 成功 |
| `10001` | 参数错误 |
| `10002` | 请求处理失败 |
| `10200` | 权限不足（RequireRole 拒绝） |
| `40001` | 未授权（JWT 验证失败） |
| `50001` | 数据库错误 |
| `50003` | 外部 API 错误 |

---

## 配置覆盖关系

```
configs/config.yaml
  ← SREAGENT_* 环境变量覆盖（28 个 Viper 绑定变量）
  ← 手动 os.Getenv：SREAGENT_SECRET_KEY / SREAGENT_DB_DEBUG / CORS_ALLOWED_ORIGINS
  ← K8s Secret 挂载为环境变量

AI / Lark 配置：
  ← AES-GCM 加密存储在 system_settings 表
  ← Web UI 设置页管理
  ← SystemSettingService.GetAIConfig / GetLarkConfig（30s 缓存）

OIDC 配置：
  ← config.yaml 的 oidc 节 / SREAGENT_OIDC_* 环境变量
  ← 字段：enabled, issuer_url, client_id, client_secret, redirect_url, scopes, role_claim
```

---

## CI/CD 概要

`.github/workflows/docker-build.yml`（3 个 job）：
- `test`：`go test ./...`
- `typecheck`：`npm run typecheck`
- `build-and-push`：buildx multi-arch，Push `main` → `:latest`，Tag `v*` → `:vX.Y.Z` + `:X.Y` + `:latest`，PR → `:pr-N`（仅构建）

详见 `docs/ci-deploy.md`。

---

## QA 修复汇总（Phase 7）

### 后端（14 项，关键/高/中均已修复）
- RequireRole / GetCurrentUserID 不安全类型断言 → comma-ok
- ChangePassword 修改了管理员自己的密码 → 改用 URL `:id` 参数
- OIDC callback 无 CSRF state 验证 → 添加 state cookie 验证
- OIDC Secure cookie flag 硬编码 false → 从 TLS/X-Forwarded-Proto 推导
- OIDC JWT 通过 query param 传递 → 改为 URL fragment
- Redis 在 HTTP Server 之前关闭 → 调换顺序
- `zap.Fatal` 在 goroutine 中阻止优雅关闭 → `zap.Error` + `os.Exit`
- StateEntry 缺少 Annotations → 添加字段

### 前端（16 项，关键/高/中均已修复）
- OIDC token 拦截更新为 hash fragment
- fetchProfile catch-all logout → 仅 401 时 logout
- 401 拦截器改用 Vue Router + 去重
- Schedule setInterval 泄露 → 生命周期管理
- MainLayout fetchProfile 移入 onMounted
- Login.vue 支持 redirect query param
- Settings 路由添加 role guard
- XSS（v-html）→ pre + text
- i18n locale 持久化到 localStorage
- Auth store role 持久化

---

## 已知遗留事项

| ID | 位置 | 问题 | 状态 |
|----|------|------|------|
| DESIGN | `larkbot.go:SendMessage` | 始终发到 DefaultWebhook，chatID 未使用 | 已知限制 |
| LOW | QA 低优先级项（4 项） | 日志级别/注释/性能优化 | 可后续迭代 |

---

## 数据库表结构概要

### system_settings（migration 000002）
```sql
CREATE TABLE system_settings (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  group_name VARCHAR(64) NOT NULL,
  key_name VARCHAR(64) NOT NULL,
  value TEXT NOT NULL,
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY idx_group_key (group_name, key_name)
);
```

**group 取值**：`ai`（provider/api_key[加密]/base_url/model/enabled）、`lark`（app_id/app_secret[加密]/default_webhook/verification_token[加密]/encrypt_key[加密]/bot_enabled）

---

## 数据模型关系

```
DataSource ──1:N── AlertRule ──1:N── AlertEvent ──1:N── AlertTimeline
                      │                                    │
                      ├── AlertRuleHistory (CRUD 审计)      └── 12 种 action type
                      └── BizGroup (作用域)

Team ──1:N── TeamMember ──N:1── User (含 OIDCSubject 字段)
Team ──1:N── Schedule ──1:N── ScheduleParticipant
                  └── ScheduleOverride / Shift

EscalationPolicy ──1:N── EscalationStep（有运行时 Executor）

NotifyRule (v2) ── match labels → NotifyMedia
MuteRule ── match labels → 抑制通知（已接入 RouteAlert）
SubscribeRule ── match labels → 额外接收人（已接入 processSubscriptions）

SystemSetting ── group+key KV（AI/Lark 配置，AES-GCM 加密）
```
