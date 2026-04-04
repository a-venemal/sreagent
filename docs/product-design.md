# SREAgent 平台产品设计文档 v1.0

> 版本：v1.0 · 更新时间：2026-04-03
> 定位：面向 SRE/运维团队的智能运维平台

---

## 平台定位

面向 SRE/运维团队的智能运维平台，核心能力：**统一告警管理 + OnCall 值班调度 + AI 辅助分析 + Lark 深度集成**。

---

## 核心概念关系图

```
数据源 ──产生──→ 告警规则 ──触发──→ 告警事件
                                      │
                    ┌─────────────────┼─────────────────────┐
                    ▼                 ▼                       ▼
              告警频道           值班排班                  屏蔽规则
         （Lark群/邮件列表）    （当前值班人）             （匹配则静默）
                    │                 │
                    ▼                 ▼
              通知媒介          分派给值班人 ──→ 用户个人通知媒介
           （发送Lark卡片）
                    │
                    ▼
              AI 分析报告（嵌入卡片）
```

---

## 一、数据源管理

### 功能概述

接入外部监控/日志系统，作为告警规则的查询目标。

### 支持类型

| 类型 | 用途 | 健康检查端点 |
|------|------|-------------|
| Prometheus | 指标查询（PromQL） | `/-/healthy` |
| VictoriaMetrics | 指标查询（PromQL 兼容） | `/health` |
| VictoriaLogs | 日志查询（LogsQL） | `/health` |
| Zabbix | 传统监控 | JSON-RPC `apiinfo.version` |

### 字段说明

- **名称**：平台内唯一标识
- **地址**：数据源 HTTP 端点
- **认证方式**：None / Basic / Bearer Token / API Key
- **标签**：`business_line=payment`、`env=prod` 等，用于后续告警路由匹配
- **健康检查间隔**：定期探活（秒）

### 与其他模块的联动

```
数据源 ──被引用──→ 告警规则（每条规则绑定一个数据源）
数据源 ──被拉取──→ AI 分析（分析时从数据源拉取关联指标）
数据源状态 ──影响──→ 告警引擎（数据源不健康时跳过该数据源下的规则评估）
```

---

## 二、告警管理

### 2.1 告警规则

#### 功能概述

定义「什么情况下产生告警」，内置 Go 评估引擎，不依赖外部 AlertManager。

#### 规则字段

| 字段 | 说明 | 示例 |
|------|------|------|
| 数据源 | 从哪里查询 | Production VM |
| 表达式 | PromQL / LogsQL | `avg(cpu_usage) > 90` |
| 持续时长 | 条件持续多久才告警（防抖） | `5m` |
| 留观时长 | 恢复后继续观察多久才关闭 | `2m` |
| 级别 | critical / warning / info | `critical` |
| 标签 | 用于路由匹配 | `business_line=payment` |
| 注解 | 告警描述模板 | `summary: "CPU高"` |
| NoData | 数据缺失时是否告警 | 开/关 |
| 评估间隔 | 多久查询一次 | `60s` |
| 业务组 | 归属的业务组 | `DBA/MySQL` |

#### 告警规则状态机

```
创建 → enabled（正常评估）
             ├─ 手动禁用 → disabled（停止评估）
             └─ 手动静默 → muted（评估但不产生事件）
```

#### 规则导入/导出

- **导入**：兼容 Prometheus YAML 格式（`groups: [{name: xx, rules: [...]}]`）
- **导出**：生成标准 Prometheus rules YAML

#### 与其他模块的联动

```
告警规则 ──评估触发──→ 告警事件（引擎发现条件满足时创建事件）
告警规则.标签 ──参与匹配──→ 告警频道（labels 子集匹配）
告警规则.标签 ──参与匹配──→ 通知规则（决定发给谁）
告警规则.标签 ──参与匹配──→ 屏蔽规则（匹配则不产生事件）
告警规则.标签 ──参与匹配──→ 值班排班（决定分派给哪个班次的值班人）
```

---

### 2.2 活跃告警

#### 功能概述

展示当前正在发生的告警，支持认领、分派、静默、解决操作。

#### 告警事件生命周期

```
firing（告警中）
    │
    ├── 手动认领 → acknowledged（已认领，by 某人）
    │       │
    │       └── 分派给他人 → assigned（已分派，to 某人）
    │
    └── 自动分派（OnCall 命中）→ assigned（is_dispatched=true）

acknowledged / assigned
    │
    └── 标记解决 → resolved（已解决，记录 resolution）
                        │
                        └── 关闭 → closed（归档）

firing（任何时候）
    └── 静默 → silenced（在静默时间内不重复通知）

firing（数据恢复）
    └── 自动恢复 → resolved（AlertManager/引擎回调）
```

#### 视图模式（基于角色）

| 视图 | 可见人群 | 过滤逻辑 |
|------|---------|---------|
| 我的告警 | 所有人 | `assigned_to = 我` OR `acked_by = 我` |
| 未分派 | 所有人 | `assigned_to IS NULL AND acked_by IS NULL AND status=firing` |
| 全局告警 | admin / global_viewer | 无过滤 |

#### 告警时间线（Timeline）

每条告警有完整操作记录：

```
2026-04-01 10:00:05  created      — 告警引擎触发
2026-04-01 10:02:30  dispatched   — 自动分派给 张三（当前值班）
2026-04-01 10:05:00  acknowledged — 张三 认领
2026-04-01 10:08:00  commented    — 张三："正在排查，CPU使用率95%"
2026-04-01 10:25:00  resolved     — 张三："扩容了2个Pod，已恢复"
2026-04-01 10:30:00  closed       — 系统自动关闭
```

#### 与其他模块的联动

```
告警事件创建 ──触发──→ OnCall 引擎（查当前值班人）→ 自动设 assigned_to
告警事件创建 ──触发──→ 通知引擎（匹配告警频道 + 通知规则）→ 发 Lark 卡片
告警事件创建 ──触发──→ AI Pipeline（拉指标 → LLM 分析 → 嵌入卡片）
静默告警    ──检查──→ 通知引擎（silenced_until 内不重复通知）
```

---

### 2.3 历史告警

#### 功能概述

已归档的告警（resolved/closed），支持多维筛选和统计。

#### 筛选维度

- 时间范围（1h / 6h / 24h / 7d / 30d / 自定义）
- 级别（critical / warning / info）
- 状态（resolved / closed）
- 来源数据源
- 处理人（认领人 / 分派人）
- 标签筛选（如 `business_line=payment`）

#### 核心指标

每条历史告警展示：

- 触发时间 → 恢复时间（持续时长）
- 认领人 + 认领耗时（MTTA）
- 解决人 + 解决耗时（MTTR）
- 触发次数（fire_count）
- AI 分析报告（如已生成）

---

### 2.4 屏蔽规则

#### 功能概述

在特定条件下阻止告警事件产生，常用于维护窗口、已知问题等场景。

#### 屏蔽条件（AND 关系）

| 维度 | 说明 | 示例 |
|------|------|------|
| 标签匹配 | 告警必须包含这些 labels | `instance=db-01` |
| 级别过滤 | 只屏蔽指定级别 | `critical,warning` |
| 规则 ID | 只屏蔽指定规则 | 规则 ID 列表 |
| 一次性时间窗口 | 在这段时间内生效 | `2026-04-01 02:00 ~ 06:00` |
| 周期性时间窗口 | 每周/每天某时间段 | 每天 `02:00-06:00`，周一至周五 |
| 时区 | 时间判断基准时区 | `Asia/Shanghai` |

#### 屏蔽规则 vs 告警静默的区别

| 维度 | 屏蔽规则 | 告警静默 |
|------|---------|---------|
| 作用对象 | 告警规则维度（评估阶段） | 单个告警事件 |
| 配置方式 | 独立管理，可复用 | 在告警详情页操作 |
| 粒度 | 批量屏蔽多条 | 针对某一条 |
| 常见场景 | 维护窗口、已知问题 | 临时忽略某条告警 |

---

## 三、通知管理

### 3.1 告警频道

#### 功能概述

虚拟的「接收端」——代表一个 Lark 群、邮件列表等。告警事件匹配到频道的 labels 时，自动推送通知。

#### 与「通知规则」的区别

```
告警频道  = 无需订阅，自动推送给指定媒介（如 Lark 群）
通知规则  = 需要有人/团队关联，描述「谁」通过「什么媒介」收告警
```

#### 字段说明

| 字段 | 说明 |
|------|------|
| 名称 | 如「支付业务线 Lark 告警群」 |
| 匹配标签 | `business_line=payment,env=prod`（全匹配） |
| 级别过滤 | `critical,warning` |
| 通知媒介 | 指向一个已配置的 NotifyMedia |
| 消息模板 | 可选，覆盖媒介默认模板 |
| 推送间隔 | 相同告警最小推送间隔（防刷屏），单位：分钟 |

#### 推送流程

```
告警事件触发
    → 遍历所有启用的告警频道
    → 检查 event.labels ⊇ channel.match_labels（子集匹配）
    → 检查 event.severity ∈ channel.severities
    → 检查节流（上次推送 + throttle_min > 现在 → 跳过）
    → 构建通知卡片（含 AI 分析）
    → 通过 media 发送
```

---

### 3.2 通知媒介

#### 功能概述

定义「怎么发」——具体的发送方式和连接参数。

#### 支持类型

| 类型 | 配置参数 | 用途 |
|------|---------|------|
| `lark_webhook` | `webhook_url` | 发到飞书群 |
| `email` | `smtp_host/port/username/password/from` | 发邮件 |
| `http` | `method/url/headers/body` | 自定义 HTTP 回调 |
| `script` | `path/args` | 执行本地脚本 |

#### 内置媒介

系统启动时自动创建「默认 Lark Webhook」和「默认邮件」两个媒介模板（`is_builtin=true`，不可删除）。

#### 测试功能

每个媒介支持「发送测试」，验证配置是否正确。

---

### 3.3 消息模板

#### 功能概述

用 Go template 语法自定义通知消息格式，不同媒介可用不同模板。

#### 可用变量

```
{{.AlertName}}    告警名称
{{.Severity}}     级别
{{.Status}}       当前状态
{{.Labels}}       标签 map（{{index .Labels "instance"}}）
{{.Annotations}}  注解 map
{{.FiredAt}}      触发时间
{{.Duration}}     持续时长
{{.Value}}        当前指标值
{{.RuleName}}     规则名称
{{.EventID}}      事件 ID（用于生成操作链接）
```

#### 模板类型

- `text`：纯文本，邮件正文等
- `markdown`：Markdown 格式
- `lark_card`：飞书交互卡片 JSON，支持按钮操作

---

### 3.4 通知规则

#### 功能概述

定义更复杂的通知路由逻辑，支持事件 Pipeline 处理（Relabel / AI 分析）。

#### 与告警频道的协作

```
告警频道  ——适合——→ 简单场景：某标签的告警发到某个群
通知规则  ——适合——→ 复杂场景：需要 Pipeline 处理、多媒介、条件分支
```

#### Pipeline 步骤（JSON 配置）

```json
[
  {"type": "relabel",    "config": {"source_label": "instance", "target_label": "host"}},
  {"type": "ai_summary", "config": {"only_critical": true}},
  {"type": "callback",   "config": {"url": "https://your-hook.com"}}
]
```

---

### 3.5 订阅规则

#### 功能概述

让用户/团队「订阅」不属于自己业务线的告警。

#### 使用场景示例

平台团队想收到所有 critical 级别告警（不管是哪个业务线），可以创建一条订阅规则：

- `match_labels: {}`（空 = 全匹配）
- `severities: critical`
- `notify_rule: 平台告警通知规则`

---

## 四、值班管理

### 4.1 排班计划

#### 核心设计理念

**排班直接到人**，不强制依赖团队。每个班次（OnCallShift）明确指定：谁在什么时间段值班，处理哪些级别的告警。

#### Schedule（排班计划）字段

| 字段 | 必填 | 说明 |
|------|:----:|------|
| 名称 | ✓ | 如「支付业务线 OnCall」 |
| 归属团队 | ✗ | 可选，用于权限和组织管理 |
| 时区 | ✓ | 班次时间的基准时区 |
| 告警级别过滤 | ✗ | 此排班只处理哪些级别的告警（默认全部） |
| 轮换类型 | ✗ | `daily` / `weekly` / `custom`，仅用于自动生成班次 |
| 交接时间 | ✗ | 自动生成时的交接时刻，如 `09:00` |

#### OnCallShift（班次）字段

| 字段 | 说明 |
|------|------|
| 值班人 | 具体某个用户（human / bot / channel 均可） |
| 开始时间 | 精确到分钟的 datetime |
| 结束时间 | 精确到分钟的 datetime |
| 告警级别过滤 | 覆盖排班计划的默认过滤，此班次只处理这些级别 |
| 来源 | `manual`（手动创建）/ `rotation`（自动生成） |
| 备注 | 可选 |

#### 日历视图

- 周视图（7 列 × 24 小时）
- 每个班次渲染为时间块，颜色按人区分
- 当前时间红线标记
- 点击空白区域 → 创建班次弹窗
- 点击班次块 → 编辑/删除

#### 自动生成班次

基于轮换配置（成员列表 + 轮换周期）自动生成未来 N 周的班次，生成后可手动微调。

#### 班次覆盖（Override）

在已有班次上创建临时覆盖（如节假日调班）：

- 指定时间段 + 替代值班人 + 原因
- Override 优先级高于普通班次

#### 告警分派逻辑

```
告警事件产生（labels: {business_line: payment, severity: critical}）
    ↓
查找所有启用的 Schedule
    ↓
筛选：schedule.labels 与告警 labels 匹配的 Schedule
    ↓
查找：当前时间点覆盖的 OnCallShift（优先查 Override）
    ↓
检查：shift.severity_filter 是否包含 critical（空 = 全包含）
    ↓
命中   → alert_event.oncall_user_id = 张三
          alert_event.is_dispatched = true
          通过张三的个人通知媒介推送
未命中 → is_dispatched = false → 进入「未分派」列表（所有人可见）
```

#### Schedule 与告警频道的关系

```
Schedule  决定「谁来处理这条告警」（分派给人）
告警频道   决定「发到哪个群通知」（推送给渠道）

两者并行不冲突：
  同一条告警 → 发到 payment Lark 群（频道） + 分派给张三（班次）
```

---

### 4.2 升级策略

#### 功能概述

当告警未在规定时间内响应时，自动升级通知范围。

#### 升级链配置示例

```
步骤 1：等待 5 分钟未认领
步骤 2：通知当前值班人的上级（team_lead）
步骤 3：再等 10 分钟
步骤 4：通知全组成员 + 发送到紧急频道
```

---

## 五、系统设置

### 5.1 用户管理

#### 账号类型

| 类型 | 说明 | 可登录 |
|------|------|:------:|
| `human` | 真实用户 | ✓ |
| `bot` | 飞书机器人（代表一个 Lark 群接收告警） | ✗ |
| `channel` | 告警频道实体（代表邮件组等） | ✗ |

#### 角色权限

| 角色 | 能力 |
|------|------|
| `admin` | 全平台管理权限 |
| `global_viewer` | 查看所有告警（不能操作） |
| `team_lead` | 管理自己团队的规则、排班、成员 |
| `member` | 查看/操作分派给自己的告警 |
| `viewer` | 只读 |

#### 个人设置（右上角头像下拉）

| 功能 | 说明 |
|------|------|
| 头像 | 选择预设 Emoji 头像 |
| 基本信息 | 修改显示名、邮件、手机 |
| 修改密码 | 需要验证旧密码 |
| 通知设置 | 配置多个个人通知媒介（Lark 个人 ID / 邮件 / Webhook） |

---

### 5.2 团队管理

团队是组织层，主要作用：

- 告警规则/排班的归属标签
- 成员权限批量授权
- 通知路由的匹配条件

---

### 5.3 业务组

树形结构（`/` 分隔），例如：

```
infrastructure
infrastructure/network
infrastructure/database
payment
payment/order
payment/refund
```

告警规则可归属业务组，屏蔽规则可按业务组批量屏蔽。

---

### 5.4 AI 配置

| 字段 | 说明 |
|------|------|
| 启用 | 开/关 AI 分析功能 |
| 服务商 | `openai` / `azure` / `ollama` / `custom` |
| API Key | 访问密钥 |
| Base URL | 可指向自建兼容 API（如 OneAPI、vLLM） |
| 模型 | `gpt-4o` / `qwen` / `llama3` 等 |

#### AI 分析链路

```
告警事件触发
    → AlertContextBuilder:
        ① 从数据源拉取告警表达式的最近 30 分钟数据
        ② 根据 labels 拉取关联指标（CPU/内存/Pod 等）
    → 组装 Prompt（中文）
    → 调用 LLM API
    → 解析返回的 JSON：
        {
          "summary":          "payment-api-01 CPU 持续高位",
          "probable_causes":  ["流量突增", "内存泄漏"],
          "impact":           "支付链路响应延迟升高",
          "recommended_steps":["查看近期部署", "top 查看热点进程"]
        }
    → 嵌入 Lark 告警卡片
```

---

### 5.5 飞书机器人

#### 配置参数

| 字段 | 说明 |
|------|------|
| App ID | 飞书开放平台应用 ID |
| App Secret | 应用密钥 |
| Verification Token | 事件回调验证 |
| Encrypt Key | 消息加密密钥 |

#### 事件回调地址

在飞书开放平台配置：

```
http://your-server:8080/lark/event
```

---

## 六、仪表盘

### 当前登录用户视角

| 模块 | 内容 |
|------|------|
| 告警引擎状态 | 运行中/停止，当前评估规则数，活跃告警数 |
| 我的活跃告警 | `assigned_to = 我` 的告警列表 |
| 最近告警 | 最近 10 条告警（按角色过滤） |
| 统计卡片 | 活跃告警数 / 数据源数 / 今日已解决 / 规则总数 |

### Admin / global_viewer 视角（额外显示）

- 全局告警趋势图
- 各业务线告警分布
- OnCall 当前值班人总览

---

## 七、功能间完整联动路径

### 路径一：新告警产生 → 通知 → 处理 → 关闭

```
1. 告警引擎评估 alert_rules
   → 条件满足 + 持续 for_duration
   → 检查屏蔽规则（mute_rules）是否命中 → 命中则跳过
   → 创建 alert_events（status=firing）

2. 告警事件创建后并行触发：

   A. OnCall 分派
      - 查 schedules（标签匹配）→ 查当前 oncall_shift
      - 命中 → event.oncall_user_id = 值班人
      - 用值班人的 user_notify_configs 推送个人通知

   B. 频道推送
      - 遍历 alert_channels（标签子集匹配）
      - 检查节流 → 通过 notify_media 发送

   C. AI 分析
      - AlertContextBuilder 拉取指标
      - LLM 生成分析报告
      - 结果嵌入 Lark 卡片一起发送

3. 值班人收到 Lark 卡片：
   [认领告警] 按钮 → 跳转 /alert-action/{token}（免登录页）
   → 选择操作（认领/静默/解决）
   → 提交 → 更新 event.status + 写 timeline

4. 平台操作（非卡片）：
   - 活跃告警 → 认领 / 分派 / 静默 / 解决 / 评论
   - 所有操作写入 alert_timelines

5. 告警恢复：
   - 引擎检测到条件不满足 + 超过 recovery_hold
   - 更新 event.status = resolved
   - 写 timeline: "Auto-resolved"
   - 可手动关闭 → closed
```

### 路径二：配置告警频道（新业务线接入）

```
1. 数据源    → 添加新数据源（如 staging VictoriaMetrics）

2. 告警规则  → 创建规则，绑定上述数据源
               labels: {business_line: order, env: staging}

3. 通知媒介  → 配置 Lark Webhook（order 业务测试群）

4. 告警频道  → 创建频道
               match_labels: {business_line: order, env: staging}
               media → 上述 Lark Webhook

5. 告警触发时，自动推送到 order 测试群

6. （可选）值班管理 → 创建排班，添加 order 业务负责人的班次
              → 告警同时分派给值班人个人
```

### 路径三：配置值班排班

```
1. 系统设置 → 用户管理 → 确认成员已存在（或创建虚拟账号）

2. 值班管理 → 新建排班
   - 名称:       "支付业务线 OnCall"
   - 时区:       Asia/Shanghai
   - 告警级别过滤: critical（只有 critical 告警才分派给值班人）
   - 归属团队:   可选

3. 点击日历空白区域 → 创建班次
   - 选人:   张三
   - 时间:   2026-04-07 09:00 ~ 2026-04-08 09:00
   - 级别过滤: 空（继承排班计划的 critical）

4. （可选）轮换成员 Tab → 配置参与者顺序
   → 点「生成排班」按钮 → 自动生成未来 4 周班次

5. 告警触发 → 引擎查当前活跃班次
   → 匹配 labels + 级别 → 分派给张三
   → 张三的 user_notify_configs（如飞书个人消息）收到通知
```

---

## 八、当前已知缺口（待实现或需讨论）

| 功能 | 现状 | 优先级 |
|------|------|:------:|
| Lark Bot 指令（@机器人） | 框架存在，指令未完整实现 | 中 |
| 告警统计报表 | 无 | 低 |
| MTTR/MTTA 统计 | 无 | 中 |
| 升级策略实际执行 | 模型存在，执行逻辑未接 | 中 |
| 告警频道卡片更新 | 告警状态变更后更新已发的 Lark 卡片 | 高 |
| 告警降噪/聚合 | 同类告警批量处理 | 中 |
| SOP 知识库 | AI 推荐操作步骤的上下文记忆 | 低 |
| 操作审计日志 | 谁在什么时间做了什么 | 中 |
| 多租户隔离 | 当前所有用户共享数据 | 低 |

---

*文档持续更新，以实际代码实现为准。*
