# n9e (Nightingale) 功能差距分析

> 目标：快速达到 n9e 能力水平，再用 Go 扩展 n9e 没有的功能
> 分析日期：2026-04-29 | SREAgent v1.16.2 vs n9e v8.x

---

## 总览

| 领域 | SREAgent 现状 | n9e 对标 | 差距 |
|------|-------------|---------|------|
| 告警引擎 | 完整，含 Pipeline | 完整，含自愈 | Pipeline 是优势 |
| 告警规则 | CRUD + 导入导出 | 含规则模板、SLO | 缺模板系统 |
| 仪表盘 | V2 面板系统 | 8+ 面板类型 | 缺多种图表 |
| 数据源 | 4 种 | 7+ 种 | 差 ES/Loki/ClickHouse |
| 通知渠道 | 飞书/邮件/Webhook | 9 种 | 差钉钉/企微/Telegram |
| RBAC | 5 角色 | Busi Group 中心化 | 权限粒度不足 |
| 值班排班 | 完整 | 内置值班 | 持平 |
| 主机管理 | 无 | Categraf + 自动发现 | 完全缺失 |
| SLO/SLI | 无 | 内置 SLO 管理 | 完全缺失 |
| 自监控 | 无 | 内置指标暴露 | 完全缺失 |

---

## 1. 告警系统 (Alerting)

### 已有
- 规则 CRUD、Prometheus 格式导入导出、多数据源
- 状态机：firing → ack → assign → resolve → close
- 静默/抑制、心跳检测、分组通知、升级策略
- 可编程处理链 (Event Pipeline) — **SREAgent 独有优势**

### 缺失
- **告警规则模板** — n9e 支持预定义规则模板，用户只需选 + 改阈值
- **SLO/SLI 管理** — n9e 内置 SLO 定义、错误预算计算、燃尽图
- **P0-P4 标准严重等级** — 目前只有 critical/warning/info
- **告警关联/聚合** — n9e 支持事件关联、根因推断
- **自愈动作** — n9e 支持告警触发后自动执行脚本/回调

### 优先级
| 功能 | 影响 | 工作量 |
|------|------|--------|
| P0-P4 严重等级 | 中 | 小（纯前端 label 重命名 + 迁移） |
| 规则模板 | 高 | 中 |
| SLO/SLI | 高 | 大 |
| 告警关联 | 中 | 大 |
| 自愈动作 | 低 | 中 |

---

## 2. 仪表盘 (Dashboards)

### 已有
- Dashboard V2：面板级仪表盘、变量模板 (query/custom/textbox/constant)
- ECharts 时间序列图表 + 表格视图
- Legend 格式化、DataZoom、Tooltip cross 指针

### 缺失（n9e 面板类型）
- **Stat 面板** — 单值大数字 + 阈值颜色（最常用的概览面板）
- **Gauge 面板** — 半圆/圆形仪表盘
- **Bar 面板** — 柱状图
- **Pie 面板** — 饼图
- **Heatmap 面板** — 热力图
- **Text/Markdown 面板** — 富文本说明面板
- **面板拖拽布局** — 目前是固定布局，需要 Grid Layout 拖拽
- **仪表盘导入导出 JSON** — 社区分享用
- **仪表盘收藏/标签** — 组织用

### 优先级
| 功能 | 影响 | 工作量 |
|------|------|--------|
| Stat 面板 | 极高 | 小（1个组件） |
| Gauge 面板 | 高 | 小 |
| 面板拖拽 Grid 布局 | 高 | 中 |
| Bar/Pie 面板 | 中 | 小 |
| 导入导出 JSON | 中 | 小 |
| Heatmap/Text | 低 | 中 |

---

## 3. 数据源 (Datasources)

### 已有
- Prometheus / VictoriaMetrics / VictoriaLogs / Zabbix
- Instant/Range Query、LogsQL、标签代理
- 健康检查、版本发现

### 缺失（n9e 原生支持）
- **Elasticsearch** — 日志查询 + 聚合
- **Loki** — 日志查询（Grafana 风格）
- **ClickHouse** — 高性能时序 + 日志分析
- **MySQL/TDW** — 结构化数据查询
- **InfluxDB** — 时序数据
- **多数据源混合查询** — 同一面板内跨数据源查询

### 优先级
| 功能 | 影响 | 工作量 |
|------|------|--------|
| Elasticsearch | 高 | 中 |
| Loki | 中 | 中 |
| ClickHouse | 中 | 大 |
| InfluxDB | 低 | 小 |
| 跨源查询 | 低 | 大 |

---

## 4. 通知渠道 (Notification)

### 已有
- 飞书 (Lark)：Webhook + Bot API + DM + 卡片模板
- 邮件：SMTP 真实发送
- Webhook：自定义 HTTP POST
- 订阅规则：用户自主订阅

### 缺失（n9e 支持渠道）
- **钉钉 (DingTalk)** — 机器人 Webhook + 卡片消息
- **企业微信 (WeCom)** — 机器人 + 应用消息
- **Telegram** — Bot API
- **Slack** — Webhook + Block Kit
- **短信 (SMS)** — 通过第三方 API
- **电话** — 语音通知
- **通知模板变量** — n9e 有更丰富的 go template 变量

### 优先级
| 功能 | 影响 | 工作量 |
|------|------|--------|
| 钉钉 | 极高 | 小（类似飞书 Webhook） |
| 企业微信 | 高 | 中 |
| Telegram | 低 | 小 |
| Slack | 低 | 小 |

---

## 5. 主机/目标管理 (Target Management)

### 已有
- **无** — 这是 SREAgent 最大缺口之一

### n9e 对标
n9e 的核心差异化功能就是 **Categraf 采集器 + 主机管理**：
- Categraf 轻量 Agent，采集 metrics/logs/traces
- 自动注册目标主机到中心
- 主机分组、标签、元数据
- Agent 心跳 + 离线告警
- 批量下发采集配置
- 插件热加载

### 建议
**不需要移植 Categraf**。改为：
1. 复用现有的 Prometheus/VM 数据源体系，通过服务发现识别目标
2. 创建「目标分组」表（类似 n9e 的 Target + Busi Group 关系）
3. 通过 Prometheus API 反向拉取 targets 列表展示

### 优先级
| 功能 | 影响 | 工作量 |
|------|------|--------|
| 目标对象表 + 分组 | 高 | 中 |
| Prom SD 目标同步 | 高 | 中 |
| Agent 心跳/离线 | 中 | 小 |
| 采集配置管理 | 低 | 大 |

---

## 6. RBAC 权限

### 已有
- 5 角色：admin / team_lead / member / viewer / global_viewer
- 路由级权限中间件
- 业务分组 (Biz Group) — 有树形结构 + match_labels

### 缺失
- **Busi Group 中心化权限** — n9e 的权限模型是用户属于 Busi Group，在 Group 内为 admin/member/viewer。SREAgent 目前的 Biz Group 更像标签作用域，没有和用户深度绑定
- **数据源级别权限** — n9e 按 Busi Group 限制可见的数据源
- **仪表盘权限** — 目前只有 public/private，没有按 Group 共享
- **操作审计增强** — n9e 有更完整的审计链路

### 优先级
| 功能 | 影响 | 工作量 |
|------|------|--------|
| Busi Group 用户绑定 | 高 | 中 |
| 数据源/仪表盘 Group 权限 | 中 | 中 |
| 审计增强 | 低 | 小 |

---

## 7. 自监控 (Self-monitoring)

### 缺失
- 无 `/metrics` 端点暴露自身指标
- 无健康检查仪表盘
- 无内部告警规则

### n9e 对标
- n9e 默认暴露 Go runtime metrics
- 内置自监控仪表盘
- 内置系统级告警规则

### 优先级
| 功能 | 影响 | 工作量 |
|------|------|--------|
| /metrics 端点 (Prometheus) | 高 | 小 |
| 自监控仪表盘 | 中 | 中 |
| 系统告警规则 | 中 | 小 |

---

## 8. 前端 UI/UX

### 已有
- Naive UI + 自定义 CSS Token 主题系统
- Dark/Light 双主题
- 中英文 i18n
- 玻璃态设计（侧栏/顶栏/登录页）

### 缺失/改进
- **全局搜索** — n9e 有 Ctrl+K 搜索（SREAgent 有 CommandPalette 但功能简单）
- **快捷操作** — n9e 列表页支持 hover 露出操作按钮
- **数据源类型图标** — Prom/VM/ES 等用品牌图标区分
- **页面加载骨架屏** — n9e 用 Skeleton 替代空白 loading
- **移动端适配** — 基础但可用

---

## 9. SREAgent 独有优势（超越 n9e）

这些是 SREAgent 已有但 n9e 没有或不如的：

| 功能 | 说明 |
|------|------|
| **可编程告警处理链** | DAG 可视化编辑器 + 5 种处理器，远超 n9e 的固定通知流程 |
| **AI 告警分析** | LLM 生成告警分析报告 + SOP 建议，n9e 无此能力 |
| **飞书深度集成** | Bot 指令回调、卡片交互、DM 通知，比 n9e 更完整 |
| **告警操作页面** | Token 认证的 HTML 页面，从通知卡片直接操作，n9e 无 |
| **升级策略** | 多步骤 + 多目标类型 (user/team/schedule)，n9e 较简单 |
| **现代前端** | Vue 3 + Naive UI + 玻璃态设计语言，n9e 前端较传统 |

---

## 10. 实施路线图

### Phase 1: 快速补洞（1-2 周） — 目标 v1.18
```
✅ Pipelines 侧栏菜单            (done)
✅ 硬编码颜色 → CSS token        (done)
✅ 死代码清理                     (done)
□ 钉钉通知渠道                    (新增)
□ Stat + Gauge 面板              (新增)
□ /metrics 端点                  (新增)
□ P0-P4 严重等级标准化           (改造)
□ 告警规则模板                    (新增)
```

### Phase 2: 核心补齐（2-4 周） — 目标 v1.20
```
□ Elasticsearch 数据源           (新增)
□ 面板拖拽 Grid 布局             (新增)
□ Busi Group 用户绑定            (改造)
□ 目标对象管理 (Prom SD 同步)    (新增)
□ 仪表盘导入导出 JSON            (新增)
□ 企业微信通知                    (新增)
□ 自监控仪表盘                    (新增)
□ 骨架屏 loading                 (新增)
```

### Phase 3: 超越 n9e（4-8 周） — 目标 v1.22+
```
□ SLO/SLI 管理系统               (新增 — n9e 有)
□ Bar/Pie/Heatmap 面板           (新增)
□ 跨数据源混合查询               (新增)
□ 告警关联/智能降噪              (增强 — 超越 n9e)
□ AI 根因分析增强               (增强 — n9e 无)
□ 移动端 PWA                     (新增)
```
