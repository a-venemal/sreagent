# 架构设计

> 最后更新：2026-04-26（v1.10.0）

## 系统架构

```
                                ┌─────────────────────────────────┐
                                │         Vue 3 Frontend          │
                                │  Naive UI + TypeScript + Pinia  │
                                └──────────────┬──────────────────┘
                                               │ HTTP REST
                                ┌──────────────▼──────────────────┐
                                │        API Layer (Gin)          │
                                │  JWT Auth / OIDC / RBAC / CORS  │
                                └──────────────┬──────────────────┘
                     ┌─────────────────────────┼─────────────────────────┐
                     │                         │                         │
          ┌──────────▼────────┐   ┌────────────▼──────────┐  ┌──────────▼────────┐
          │  DataSource Svc   │   │   Alert Engine        │  │  OnCall Svc       │
          │  Prom/VM/VLogs    │   │  Evaluator + FSM      │  │  Schedule/Escalate│
          │  Zabbix           │   │  Heartbeat/Suppress   │  │                   │
          └──────────┬────────┘   └────────────┬──────────┘  └──────────┬────────┘
                     │                         │                        │
          ┌──────────▼─────────────────────────▼────────────────────────▼──┐
          │                          Redis 7                               │
          │  引擎状态持久化 (Hash per rule) · 节流                         │
          └──────────────────────────────┬─────────────────────────────────┘
                                         │
          ┌──────────────────────────────▼─────────────────────────────────┐
          │                       MySQL 8.0                                │
          │  22 张表 · golang-migrate 管理 · GORM v2 ORM                  │
          └────────────────────────────────────────────────────────────────┘

                     ┌───────────────────────────────────┐
                     │       外部集成                     │
                     │  Lark Bot+Hook · LLM API          │
                     │  Email SMTP · Custom Webhooks      │
                     │  Keycloak OIDC                    │
                     └───────────────────────────────────┘
```

## 关键架构决策

| 决策 | 原因 |
|------|------|
| AI/Lark 配置存 DB（system_settings），AES-256-GCM 加密 | 密钥不出现在 ConfigMap/Secret |
| golang-migrate 是 schema 唯一来源，GORM AutoMigrate 只作安全网 | 迁移可审计可回滚 |
| Redis Hash 持久化引擎状态 | 重启后恢复飞行中告警，Redis 不可用时降级到纯内存 |
| OIDC 配置存 DB，启动时合并 configmap | 运行时配置无需重启 |
| RBAC 三级权限（adminOnly/manage/operate） | 精细权限控制 |
| 多数据源路由（Prom/VM/VLogs/Zabbix） | 支持异构监控体系 |

## 告警引擎状态机

```
inactive → pending（for_duration）→ firing → recovery_hold → resolved
                                        └── nodata
```

- 每规则一个 goroutine，Evaluator 管理协程池
- LevelSuppressor 基于严重级别去重
- HeartbeatChecker 心跳超时检测
- EscalationExecutor SLA 超时自动升级
- AlertGroupManager group_wait/interval 通知分组

## 通知管道

```
Engine fires → AlertGroupManager → Inhibition → Mute → RouteAlert
  → v1 策略管道 (NotifyChannel + NotifyPolicy)
  → v2 规则管道 (NotifyRule → NotifyMedia)
  → 订阅管道 (SubscribeRule → NotifyRule)
  → SendNotification (lark_webhook / lark_bot / email / webhook / script)
```
