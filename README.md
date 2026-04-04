# SREAgent

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-Proprietary-red?style=flat-square)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-multi--arch-2496ED?style=flat-square&logo=docker)](https://hub.docker.com/)

**面向 SRE/运维团队的智能告警管理平台**：统一告警生命周期管理、OnCall 值班调度、AI 辅助分析与飞书（Lark）深度集成。

---

## 目录

- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [构建镜像](#构建镜像)
- [Kubernetes 部署](#kubernetes-部署)
- [默认账号](#默认账号)
- [API 文档](#api-文档)
- [开发指南](#开发指南)
- [项目结构](#项目结构)

---

## 功能特性

### 告警管理
- **数据源接入** — 支持 Prometheus、VictoriaMetrics、VictoriaLogs、Zabbix，内置健康检查
- **告警规则引擎** — 内置 Go 评估引擎（不依赖外部 AlertManager），支持 PromQL/LogsQL，含防抖（`for_duration`）与留观（`recovery_hold`）机制
- **AlertManager Webhook 兼容** — 可直接接收来自 AlertManager/VMAlert 的标准 Webhook 推送
- **告警事件完整生命周期** — `firing → acknowledged → assigned → resolved → closed`，支持认领、分派、静默、评论
- **告警时间线（Timeline）** — 每条告警的完整操作审计记录
- **屏蔽规则（Mute Rules）** — 支持一次性与周期性时间窗口，按标签/级别/规则 ID 批量屏蔽（常用于维护窗口）
- **批量操作** — 批量认领、批量关闭告警事件

### 通知路由
- **告警频道（Alert Channels）** — 基于标签子集匹配，自动将告警推送到指定 Lark Webhook 群，含节流防刷屏
- **通知媒介（Notify Media）** — 支持 Lark Webhook、邮件、HTTP 回调、脚本，可发送测试消息
- **消息模板** — 使用 Go template 语法自定义 Lark 卡片、Markdown、纯文本消息格式
- **通知规则（Notify Rules）** — 支持 Pipeline 处理（Relabel、AI 摘要、自定义 Callback）
- **订阅规则（Subscribe Rules）** — 用户/团队可跨业务线订阅感兴趣的告警

### OnCall 值班调度
- **排班计划（Schedules）** — 日历视图，直接对人排班，支持日/周/自定义轮换
- **班次管理（Shifts）** — 精确到分钟的手动排班，支持自动生成未来 N 周班次
- **班次覆盖（Overrides）** — 节假日调班、临时换班，优先级高于普通班次
- **升级策略（Escalation Policies）** — 超时未认领自动升级通知范围，多步骤升级链配置
- **告警自动分派** — 新告警触发时根据标签匹配当前值班人，自动设置 `assigned_to`

### AI 辅助分析
- **告警报告生成** — 自动拉取数据源指标作为上下文，通过 LLM 生成分析报告并嵌入 Lark 卡片
- **SOP 推荐** — 根据告警上下文推荐处理步骤
- **多服务商支持** — OpenAI、Azure OpenAI、Ollama（本地）、自定义兼容接口（OneAPI/vLLM）

### 飞书（Lark）集成
- **Webhook 通知** — 发送富文本交互卡片，包含操作按钮（认领/静默/解决）
- **Lark Bot** — 支持事件回调，可通过飞书机器人与平台交互
- **免登录操作页** — 告警卡片中的按钮跳转至 `/alert-action/:token`，无需登录即可操作

### 组织与权限
- **RBAC** — `admin / team_lead / member / viewer` 四级角色
- **团队管理** — 支持标签关联，用于权限隔离与通知路由匹配
- **业务组（Biz Groups）** — 树形结构（`/` 分隔），如 `infrastructure/database`
- **虚拟用户** — 支持 `bot`（飞书机器人代理）和 `channel`（告警频道实体）类型
- **个人通知配置** — 每个用户可配置多个个人通知媒介（飞书个人 ID / 邮件 / Webhook）

### 平台能力
- **实时仪表盘** — 告警引擎状态、活跃告警统计、个人待处理告警
- **规则导入/导出** — 兼容 Prometheus YAML 格式（`groups: [{name, rules}]`）
- **自动数据库迁移** — 启动时通过 golang-migrate 自动完成建表和升级，零人工干预

---

## 技术栈

| 层次 | 技术 | 版本 |
|------|------|------|
| **后端语言** | Go | 1.24 |
| **HTTP 框架** | Gin | v1.10 |
| **ORM** | GORM | v2 |
| **数据库迁移** | golang-migrate | v4 |
| **配置管理** | Viper | v1.19 |
| **日志** | Zap | v1.27 |
| **认证** | golang-jwt/jwt | v5 |
| **数据库** | MySQL 8.0（OceanBase 兼容） | 8.0+ |
| **缓存** | Redis | 7.x |
| **前端框架** | Vue 3 + TypeScript | 3.5+ |
| **UI 组件库** | Naive UI | 2.x |
| **构建工具** | Vite | 6.x |
| **图表** | ECharts | 5.x |
| **容器** | Docker（多阶段构建，多架构） | — |
| **编排** | Kubernetes | — |

---

## 快速开始

### 方式一：Docker（单容器，外挂 MySQL + Redis）

确保你已有可用的 MySQL 8.0 和 Redis 7 实例，然后：

```bash
# 1. 克隆仓库
git clone <your-repo-url> sreagent
cd sreagent

# 2. 准备配置
cp configs/config.example.yaml configs/config.yaml
# 编辑 config.yaml，填写 database.password、redis.password、jwt.secret

# 3. 构建镜像
docker build -f deploy/docker/Dockerfile -t sreagent:latest .

# 4. 启动服务（首次启动自动完成数据库建表）
docker run -d --name sreagent \
  -p 8080:8080 \
  -v $(pwd)/configs/config.yaml:/app/configs/config.yaml:ro \
  sreagent:latest
```

**访问地址：**

| 服务 | 地址 |
|------|------|
| Web UI | `http://your-server-ip:8080` |
| API | `http://your-server-ip:8080/api/v1` |
| 健康检查 | `http://your-server-ip:8080/healthz` |

**常用操作：**

```bash
docker logs -f sreagent        # 实时查看服务日志
docker restart sreagent        # 修改配置后重启
docker rm -f sreagent          # 停止并删除容器
```

---

### 方式二：本地开发

#### 前置依赖

| 依赖 | 版本要求 |
|------|---------|
| Go | 1.24+ |
| Node.js | 20+ |
| MySQL | 8.0+ |
| Redis | 7+ |

#### 步骤

**1. 克隆仓库**

```bash
git clone <your-repo-url> sreagent
cd sreagent
```

**2. 准备配置文件**

```bash
cp configs/config.example.yaml configs/config.yaml
```

编辑 `configs/config.yaml`，至少填写数据库密码和 Redis 密码（或通过环境变量覆盖，见[配置说明](#配置说明)）。

**3. 启动依赖（MySQL + Redis）**

```bash
make docker-up
```

**4. 启动后端**

```bash
# 直接运行
make run

# 或使用 air 热重载（需先安装：go install github.com/air-verse/air@latest）
make dev
```

后端服务启动在 `http://localhost:8080`，首次启动会自动完成数据库建表。

**5. 启动前端**

```bash
make web-install   # 安装 npm 依赖
make web-dev       # 启动 Vite 开发服务器（含 API 代理）
```

前端开发服务器启动在 `http://localhost:3000`，API 请求自动代理到后端。

**6. 登录**

打开浏览器访问 `http://localhost:3000`，使用默认账号登录（见[默认账号](#默认账号)）。

---

## 配置说明

### 5.1 必须配置（启动所需）

以下三项是服务启动的必要配置。推荐通过**环境变量**注入，避免将密钥写入配置文件。

| 配置项（YAML 路径） | 环境变量 | 说明 | 示例 |
|---|---|---|---|
| `database.password` | `SREAGENT_DATABASE_PASSWORD` | MySQL 数据库密码 | `your-db-password` |
| `redis.password` | `SREAGENT_REDIS_PASSWORD` | Redis 密码（无密码时留空） | `your-redis-password` |
| `jwt.secret` | `SREAGENT_JWT_SECRET` | JWT 签名密钥，建议 32 字节以上随机字符串 | `openssl rand -hex 32` |

**其他常用配置项（`configs/config.yaml`）：**

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"           # 生产环境改为 "release"
  # external_base: "https://sreagent.example.com"  # 通知消息中的链接基础地址

database:
  host: "127.0.0.1"
  port: 3306
  username: "sreagent"
  database: "sreagent"

redis:
  host: "127.0.0.1"
  port: 6379
  db: 0

jwt:
  expire: 86400           # Token 有效期（秒）

engine:
  enabled: true
  sync_interval: 30       # 告警规则同步间隔（秒）
```

### 5.2 平台运行时配置（Web UI 配置）

以下配置**不在配置文件中**，而是存储在数据库，通过 **Web UI → 系统设置** 页面进行管理：

| 功能 | 配置入口 | 说明 |
|------|---------|------|
| **AI 配置** | 设置 → AI 配置 | 服务商（OpenAI/Azure/Ollama/自定义）、API Key、Base URL、模型名称 |
| **飞书机器人** | 设置 → 飞书机器人 | App ID、App Secret、Verification Token、Encrypt Key、默认 Webhook |
| **通知媒介** | 通知 → 通知媒介 | Lark Webhook URL、邮件 SMTP、HTTP 回调等 |
| **告警频道** | 通知 → 告警频道 | 标签匹配规则、关联通知媒介、节流配置 |

> **提示：** AI 和 Lark Bot 的敏感凭据均通过 Web UI 写入数据库，无需挂载配置文件或注入环境变量。

---

## 构建镜像

### 单架构构建

```bash
docker build -f deploy/docker/Dockerfile -t sreagent:latest .
```

### 多架构构建（linux/amd64 + linux/arm64）

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f deploy/docker/Dockerfile \
  -t your-repo/sreagent:latest \
  --push .
```

### CI/CD 自动构建

项目已配置 GitHub Actions（`.github/workflows/docker-build.yml`），自动触发规则：

| 触发条件 | 镜像标签 | 说明 |
|---------|---------|------|
| Push 到 `main` 分支 | `:latest` | 构建并推送到 Docker Hub |
| Push `v*` 格式 Tag | `:v1.2.3`、`:1.2`、`:latest` | SemVer 语义化标签 |
| PR 到 `main` | `:pr-<number>` | 仅构建验证，不推送 |

流水线包含：Go 单元测试 → 前端 TypeScript 类型检查 → 多架构镜像构建推送。

---

## Kubernetes 部署

所有 K8s 配置文件位于 `deploy/kubernetes/` 目录。

### 部署步骤

**第 1 步：创建命名空间**

```bash
kubectl apply -f deploy/kubernetes/00-namespace.yaml
```

**第 2 步：部署 MySQL**

```bash
kubectl apply -f deploy/kubernetes/mysql/
```

**第 3 步：部署 Redis**

```bash
kubectl apply -f deploy/kubernetes/redis/
```

**第 4 步：编辑 Secret（填入真实密码）**

编辑 `deploy/kubernetes/app/secret.yaml`，替换占位符：

```yaml
stringData:
  db-password: "your-real-db-password"
  redis-password: "your-real-redis-password"
  jwt-secret: "your-32-char-random-secret"    # openssl rand -hex 32
```

```bash
kubectl apply -f deploy/kubernetes/app/secret.yaml
```

**第 5 步：编辑 ConfigMap（填入访问域名和镜像名）**

编辑 `deploy/kubernetes/app/configmap.yaml`，修改两处：

```yaml
# 1. 改为你的实际对外访问地址（用于通知消息中的跳转链接）
external_base: "https://sreagent.your-domain.com"
```

编辑 `deploy/kubernetes/app/deployment.yaml`，修改镜像地址：

```yaml
image: your-dockerhub-username/sreagent:latest
```

```bash
kubectl apply -f deploy/kubernetes/app/configmap.yaml
```

**第 6 步：部署应用**

```bash
kubectl apply -f deploy/kubernetes/app/
```

**验证部署状态：**

```bash
# 查看 Pod 是否就绪
kubectl -n sreagent get pods

# 查看服务日志
kubectl -n sreagent logs -f deployment/sreagent

# 检查健康端点
kubectl -n sreagent port-forward svc/sreagent 8080:8080
curl http://localhost:8080/healthz
```

> **注意：** 告警引擎使用内存状态机，默认 `replicas: 1`。如需多副本水平扩展，需在引擎层引入分布式锁（Redis 互斥锁）。

---

## 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| `admin` | `admin123` | admin（全平台管理员） |

> **安全警告：** 首次登录后请**立即修改**默认密码。进入右上角头像 → 个人设置 → 修改密码。

---

## API 文档

所有 API 使用 `/api/v1` 前缀，除登录接口和 Webhook 外均需携带 JWT Token：

```
Authorization: Bearer <token>
```

**统一响应格式：**

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

**分页参数：** `?page=1&page_size=20`

### API 路由一览

| 模块 | 路径前缀 | 主要操作 |
|------|---------|---------|
| **认证** | `/api/v1/auth` | 登录、获取 Profile |
| **个人信息** | `/api/v1/me` | 更新资料、修改密码、个人通知配置 |
| **数据源** | `/api/v1/datasources` | CRUD、手动触发健康检查 |
| **告警规则** | `/api/v1/alert-rules` | CRUD、启用/禁用、规则导入/导出 |
| **告警事件** | `/api/v1/alert-events` | 列表、详情、认领/分派/解决/关闭/静默/评论、时间线、批量操作 |
| **屏蔽规则** | `/api/v1/mute-rules` | CRUD |
| **告警频道** | `/api/v1/alert-channels` | CRUD |
| **通知媒介** | `/api/v1/notify-media` | CRUD、发送测试 |
| **通知规则** | `/api/v1/notify-rules` | CRUD |
| **消息模板** | `/api/v1/message-templates` | CRUD、预览渲染 |
| **订阅规则** | `/api/v1/subscribe-rules` | CRUD |
| **通知渠道（旧）** | `/api/v1/notify-channels` | CRUD、发送测试 |
| **通知策略（旧）** | `/api/v1/notify-policies` | CRUD |
| **用户管理** | `/api/v1/users` | CRUD、启用/禁用、修改密码、创建虚拟用户 |
| **团队管理** | `/api/v1/teams` | CRUD、成员管理 |
| **业务组** | `/api/v1/biz-groups` | CRUD、树形列表、成员管理 |
| **排班计划** | `/api/v1/schedules` | CRUD、班次管理、当前值班人、Override、自动生成班次 |
| **升级策略** | `/api/v1/escalation-policies` | CRUD、步骤管理 |
| **AI** | `/api/v1/ai` | 生成报告、SOP 推荐、配置读写、连通性测试 |
| **飞书机器人** | `/api/v1/lark/bot` | 配置读写 |
| **告警引擎** | `/api/v1/engine/status` | 获取引擎运行状态 |
| **仪表盘** | `/api/v1/dashboard/stats` | 统计数据 |
| **Webhook** | `/webhooks/alertmanager` | 接收 AlertManager/VMAlert Webhook（无需认证） |
| **飞书事件回调** | `/lark/event` | 接收飞书机器人事件（无需认证，Token 验证） |
| **告警操作页** | `/alert-action/:token` | 免登录告警操作（Token 鉴权） |

### AlertManager Webhook 集成

将 AlertManager 或 VMAlert 的 Webhook 地址配置为：

```
http://<sreagent-host>:8080/webhooks/alertmanager
```

支持标准 AlertManager Webhook payload 格式，可直接与 Prometheus/VictoriaMetrics 告警系统对接。

---

## 开发指南

### 运行测试

```bash
# 运行所有 Go 单元测试
go test ./... -timeout 120s

# 带覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 前端构建

```bash
# 安装依赖
cd web && npm install

# 开发模式（含热重载）
npm run dev

# TypeScript 类型检查
npm run typecheck

# 生产构建（输出到 web/dist/）
npm run build
```

### 添加数据库迁移

迁移文件位于 `internal/pkg/dbmigrate/migrations/`，使用 golang-migrate 管理。新增迁移：

1. 按命名规范创建迁移文件：

```
migrations/
  000001_initial_schema.up.sql
  000001_initial_schema.down.sql
  000002_system_settings.up.sql
  000002_system_settings.down.sql
  000003_your_change.up.sql
  000003_your_change.down.sql
```

2. 服务启动时自动执行待执行的迁移（无需手动命令）。

3. 文件命名规范：`{6位序号}_{描述}.{up|down}.sql`，版本号零填充递增。

### Makefile 常用命令

```bash
make help          # 列出所有可用命令
make run           # 直接运行后端服务
make dev           # air 热重载模式
make build         # 编译 Go 二进制
make test          # 运行测试
make lint          # 运行 linter
make fmt           # 格式化代码
make docker-up     # 启动本地依赖（MySQL + Redis）
make docker-down   # 停止本地依赖
make docker-build  # 构建 Docker 镜像
make web-install   # 安装前端依赖
make web-dev       # 启动前端开发服务器
make web-build     # 构建前端生产包
```

---

## 项目结构

```
sreagent/
├── cmd/server/              # 应用入口（main.go）
├── internal/
│   ├── config/              # 配置结构体（Viper）
│   ├── model/               # GORM 数据模型
│   ├── handler/             # HTTP 处理器（Gin）
│   ├── service/             # 业务逻辑层
│   ├── repository/          # 数据访问层
│   ├── middleware/          # 中间件（JWT 认证、CORS、日志、限流）
│   ├── router/              # 路由注册（router.go）
│   └── pkg/
│       ├── datasource/      # 数据源健康检查客户端（Prometheus/VM/Zabbix/VLogs）
│       ├── dbmigrate/       # golang-migrate 运行器 + SQL 迁移文件
│       ├── lark/            # 飞书 Webhook 客户端 + 卡片模板
│       ├── redis/           # Redis 客户端封装
│       └── errors/          # 结构化错误类型
├── web/                     # Vue 3 前端
│   └── src/
│       ├── api/             # Axios API 客户端
│       ├── pages/           # 页面组件（Dashboard、Alerts、Schedule、Settings 等）
│       ├── stores/          # Pinia 状态管理
│       ├── router/          # Vue Router（含认证守卫）
│       └── types/           # TypeScript 类型定义
├── deploy/
│   ├── docker/              # Dockerfile + entrypoint.sh
│   └── kubernetes/          # K8s 清单文件（namespace、mysql、redis、app）
├── configs/
│   └── config.example.yaml  # 配置文件模板
├── .github/workflows/       # GitHub Actions CI/CD
└── Makefile
```

---

## License

内部项目，保留所有权利。
