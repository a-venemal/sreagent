# Monitoring Trading Repository

基于 **VMAgent + git-sync + VictoriaMetrics + VMAlert + Alertmanager** 的监控与告警一体化配置仓库。
使用 file_sd 实现动态服务发现与自动热更新，覆盖 ~305 条告警规则。

> **多集群联邦架构**：多个 K8s 集群（public-monitor / business-public / win-k8s）通过联邦/remote-write 写入同一个 VMS，所有告警规则均已针对多集群场景做过验证，通过 `project` 标签区分不同集群来源。

## 📁 目录结构

```
monitoring-trading/
├── alerts/                          # 告警规则（VMAlert / Prometheus rules 格式）
│   ├── node-exporter/               # Linux 主机告警（6 文件，~52 条规则）
│   │   ├── cpu.yaml
│   │   ├── memory.yaml
│   │   ├── disk.yaml
│   │   ├── network.yaml
│   │   ├── filesystem.yaml
│   │   └── system.yaml
│   ├── windows-exporter/            # Windows 主机告警（5 文件，31 条规则）
│   │   ├── cpu.yaml
│   │   ├── memory.yaml
│   │   ├── disk.yaml
│   │   ├── network.yaml
│   │   └── system.yaml
│   ├── kubernetes/                  # K8s 基础设施告警（9 文件，~88 条规则）
│   │   ├── apiserver.yaml
│   │   ├── container.yaml
│   │   ├── coredns.yaml
│   │   ├── kube-controller-manager.yaml
│   │   ├── kube-etcd.yaml
│   │   ├── kube-scheduler.yaml
│   │   ├── kube-state-metrics.yaml
│   │   ├── kubelet.yaml
│   │   └── resource-reservation.yaml    # 资源预留与调度告警（新增）
│   ├── middleware/                   # 中间件告警（5 文件，68 条规则）
│   │   ├── etcd.yaml
│   │   ├── kafka.yaml
│   │   ├── nacos.yaml
│   │   ├── rabbitmq.yaml
│   │   └── rocketmq.yaml
│   ├── database/                    # 数据库告警（4 文件，~64 条规则）
│   │   ├── clickhouse.yaml
│   │   ├── elasticsearch.yaml
│   │   ├── mongodb.yaml
│   │   └── redis.yaml
│   ├── probe/                       # 探测告警（2 文件，12 条规则）
│   │   ├── blackbox-http.yaml
│   │   └── blackbox-tcp.yaml
│   └── templates/                   # 通知模板
│       └── feishu-card.tmpl         # 飞书卡片消息模板
├── kubernetes/                      # K8s 集群部署配置
│   ├── public-monitor/              # 监控集群（vmagent / vmalert / alertmanager / blackbox）
│   ├── share-public-business/       # 公共业务集群（kube-prometheus-stack + blackbox TCP 探测）
│   ├── bigdata-flink/               # Flink 大数据集群（kube-prometheus-stack，remote write 到 VMS）
│   └── win-k8s/                     # Windows K8s 集群（VictoriaMetrics stack）
└── targets/                         # 采集目标配置（vmagent file_sd）
    └── <biz_project>/               # 业务线（ts / cc / mdc / cpp / metatradertools / product-center）
        └── <tenant>/                # 租户
            ├── host/                # 主机监控（node-exporter）
            ├── kubernetes/          # K8s 集群监控（Prometheus federation）
            ├── probe/               # L4 层探测（blackbox TCP 端口连通性）
            ├── domain/              # L7 层域名监控（blackbox HTTP/HTTPS + 证书有效期）
            ├── db/                  # 数据库 exporter
            └── midd/                # 中间件 exporter
```

## 🏗️ 架构概览

```
Git 仓库 (monitoring-trading)
    ├── alerts/     <-- 告警规则 (~305条)
    ├── targets/    <-- 采集目标
    └── kubernetes/ <-- K8s 部署配置

         | git-sync (30s)
         v

VMAgent --采集指标--> VictoriaMetrics (VMS)
                           ^
VMAlert --评估规则---------+
    |
    | 触发告警
    v
Alertmanager --按 biz_project 分流--> webhook-feishu --> 飞书群
    |                                                    ├── ts (交易系统)
    | 降噪/抑制/聚合                                       ├── cc (清算系统)
    |                                                    ├── mdc (行情数据)
    |                                                    ├── cpp (C++服务)
    |                                                    ├    ├── rps (RPS / minirps / official-rps)
    |                                                    ├    └── ocs (OCS 服务)
    |                                                    ├── metatradertools
    |                                                    ├── product-center
    |                                                    └── infra (基础设施)
```

**核心组件：**

| 组件 | 作用 |
|------|------|
| **VMAgent** | 通过 file_sd 采集 targets/ 下所有监控目标，remote_write 到 VMS |
| **VictoriaMetrics (VMS)** | 统一时序数据存储 |
| **VMAlert** | 加载 alerts/ 下所有告警规则，定期评估并发送到 Alertmanager |
| **Alertmanager** | 按 `biz_project` 标签路由告警，降噪/抑制/聚合后发送到 webhook |
| **webhook-feishu** | 将告警转为飞书卡片消息，投递到对应业务飞书群 |
| **blackbox-exporter** | TCP 端口探测（公共节点），服务 ts/cc/metatradertools/product-center/cpp 业务线 |
| **blackbox-exporter-mdc** | TCP 端口探测（MDC 独占节点），服务 mdc 业务线（01sec/03sec） |
| **git-sync** | 每 30s 同步 Git 仓库到 Pod 本地卷 |

## 🚨 告警规则

### 告警统计

| 类别 | 目录 | 规则数 |
|------|------|--------|
| Linux 主机 | `node-exporter/` | ~52 |
| Windows 主机 | `windows-exporter/` | 31 |
| K8s 基础设施 | `kubernetes/` | ~88 |
| 中间件 | `middleware/` | 68 |
| 数据库 | `database/` | ~64 |
| 探测 | `probe/` | 12 |
| **总计** | | **~315** |

### 告警分级（P0-P3）

| 级别 | 含义 | 处理要求 | 示例 |
|------|------|---------|------|
| **P0** | Critical | 需要立即处理 | 节点 Down、OOM Kill、磁盘只读、K8s 组件不可用 |
| **P1** | Warning/High | 需要关注可能需要处理 | CPU>90%、内存<5%、磁盘空间<10%、Pod CrashLoop |
| **P2** | Warning/Medium | 需要关注 | iowait>30%、磁盘空间<20%、证书即将过期、网络高错误率 |
| **P3** | Info | 信息通知 | 负载偏高、内存增长趋势、上下文切换高、uptime 超长 |

### 标签设计

每条告警规则包含以下标签，供 Alertmanager 路由和飞书卡片展示：

```yaml
labels:
  severity: P0/P1/P2/P3                                # 告警级别
  category: cpu/memory/disk/network/filesystem/system   # 告警维度
  alert_type: threshold/status/event/trend/prediction   # 告警类型
```

告警触发时，以下业务标签从时间序列自动继承到告警（通过 PromQL 表达式保留），供 Alertmanager 路由和飞书卡片展示：

- **强制保留**：`project`、`biz_project`、`tenant`、`maintainer`（所有告警规则）
- **按类型保留**：`service`（node-exporter / 中间件 / 数据库 / 探测）、`cluster`（中间件 / 数据库）、`namespace`（K8s 资源）

### 部署方式

VMAlert 通过 `-rule` 参数加载告警规则文件：

```
-rule=/rules/alerts/node-exporter/*.yaml
-rule=/rules/alerts/windows-exporter/*.yaml
-rule=/rules/alerts/kubernetes/*.yaml
-rule=/rules/alerts/middleware/*.yaml
-rule=/rules/alerts/database/*.yaml
-rule=/rules/alerts/probe/*.yaml
```

### 新增：资源预留与调度告警（resource-reservation.yaml）

**文件位置**：`alerts/kubernetes/resource-reservation.yaml`

**监控场景**：
- 🎯 **资源预留率过高**：requests/allocatable > 85%，新 Pod 无法调度
- 🎯 **资源碎片化**：预留率高但实际使用率低，资源浪费
- 🎯 **超大预留 Pod**：单个 Pod 预留超过节点容量的 50%，影响调度灵活性
- 🎯 **集群容量不足**：集群整体预留率过高，需扩容

**告警规则**：

| 告警名称 | 级别 | 触发条件 | 说明 |
|---------|------|---------|------|
| `KubeNodeCPUReservationCritical` | P0 | CPU 预留率 >90% | 新 Pod 无法调度 |
| `KubeNodeCPUReservationHigh` | P1 | CPU 预留率 >80% | 预警 |
| `KubeNodeMemoryReservationCritical` | P0 | 内存预留率 >90% | 新 Pod 无法调度 |
| `KubeNodeMemoryReservationHigh` | P1 | 内存预留率 >80% | 预警 |
| `KubeNodeCPUFragmentation` | P1 | 预留>80% 且 使用<50% | 资源浪费 |
| `KubeNodeMemoryFragmentation` | P1 | 预留>80% 且 使用<50% | 资源浪费 |
| `KubePodOversizedCPURequest` | P2 | 单 Pod CPU>节点 50% | 调度不灵活 |
| `KubePodOversizedMemoryRequest` | P2 | 单 Pod 内存>节点 50% | 调度不灵活 |
| `KubeClusterCPUReservationHigh` | P1 | 集群 CPU>85% | 需扩容 |
| `KubeClusterMemoryReservationHigh` | P1 | 集群内存>85% | 需扩容 |

**核心指标**：
- `kube_pod_container_resource_requests`: Pod 请求的资源量
- `kube_pod_container_resource_limits`: Pod 限制的资源
- `kube_node_status_allocatable`: 节点可分配资源
- `kube_node_status_capacity`: 节点总容量

**PromQL 示例**（节点 CPU 预留率）：
```promql
sum by (node, project, biz_project, tenant, maintainer) (
  kube_pod_container_resource_requests{resource="cpu", unit="core"}
)
/
sum by (node, project, biz_project, tenant, maintainer) (
  kube_node_status_allocatable{resource="cpu", unit="core"}
) * 100
```

## 📬 告警链路

### 1. 规则评估（VMAlert）

VMAlert 定期从 VictoriaMetrics 查询告警规则中定义的 PromQL 表达式。当条件满足持续时间（`for`）后，告警变为 firing 状态，发送到 Alertmanager。

### 2. 路由分流（Alertmanager）

Alertmanager 按 `biz_project` 标签将告警路由到对应的 webhook receiver：

| biz_project | 飞书群 | 说明 |
|-------------|--------|------|
| `ts` | 交易系统告警群 | 交易系统 |
| `cc` | 清算系统告警群 | 清算系统 |
| `mdc` | 行情数据告警群 | 行情数据中心 |
| `cpp` | C++服务告警群 | C++ 服务 |
| `metatradertools` | MetaTradeTools 告警群 | MetaTradeTools |
| `product-center` | 产品中心告警群 | 产品中心 |
| `infra` | 基础设施告警群 | 基础设施（默认） |

### 3. 降噪与抑制（inhibit_rules）

Alertmanager 配置了多层抑制规则，避免告警风暴。所有规则的 `equal` 均包含 `project`，防止不同集群间误抑制。

| 层级 | Source 告警 | Target 告警 | 匹配维度 |
|------|------------|------------|---------|
| 主机级 severity | P0 | P1/P2/P3 | `biz_project` + `category` + `instance` + `project` |
| 主机级 severity | P1 | P2/P3 | `biz_project` + `category` + `instance` + `project` |
| 容器级 severity | P0 (category=container) | P1/P2/P3 (category=container) | `biz_project` + `namespace` + `pod` + `container` + `project` |
| 容器级 severity | P1 (category=container) | P2/P3 (category=container) | `biz_project` + `namespace` + `pod` + `container` + `project` |
| 主机 Down | `NodeExporterDown` | 所有告警 | `biz_project` + `instance` + `project` |
| K8s 节点异常 | `KubeNodeNotReady` | category=container / pod | `biz_project` + `node` + `project` |
| 中间件 Down | `KafkaExporterDown` / `RedisDown` / `MongoDBDown` / `RabbitMQDown` / `NacosDown` / `RocketMQExporterDown` | 对应 category | `biz_project` + `instance` + `project` |
| ES 状态 | `ElasticsearchClusterRed` | `ElasticsearchClusterYellow` | `biz_project` + `instance` + `project` |
| HTTP 探测失败 | `BlackboxHttpProbeFailed` | 延迟/状态码告警 | `biz_project` + `instance` + `project` |
| TCP 探测失败 | `BlackboxTcpProbeFailed` | 延迟告警 | `biz_project` + `instance` + `project` |

> 所有规则均包含 `biz_project` + `project`，前者防止多业务线共用主机/中间件时的跨业务线误抑制，后者防止多集群间的误抑制。

### 4. 飞书卡片通知（webhook-feishu）

使用 `alerts/templates/feishu-card.tmpl` 模板渲染飞书互动卡片，每行展示一个信息：

| 行 | 内容 | 标签来源 |
|----|------|---------|
| 状态行 | 触发中/已恢复 \| P0 \| category | `severity`、`category` |
| 摘要 | 告警摘要 | annotations.summary |
| 集群 | 数据来源集群 | `project` |
| 业务线 | 所属业务 | `biz_project` |
| 租户 | 租户标识 | `tenant` |
| 服务 | 服务类型 | `service` |
| 集群标识 | 中间件/数据库集群 | `cluster` |
| 命名空间 | K8s namespace | `namespace` |
| Pod | Pod + 容器 | `pod`、`container` |
| 主机 | 主机名 + 实例地址 | `hostname`、`instance` |
| 详情 | 告警描述 | annotations.description |
| 时间 | 开始 → 结束/持续中 | StartsAt、EndsAt |
| 按钮 | 操作手册 / 告警中心 | annotations.runbook_url |

> **注意**：`maintainer` 标签不在卡片中展示，但会传递到 Alertmanager 用于路由和分组。

## 🏷️ 标签规范（强制）

每个 target group **必须**包含：

| Label | 说明 | 示例值 | 必填 |
|-------|------|--------|------|
| `biz_project` | 业务线 | `mdc` / `ts` / `cc` / `cpp` | ✅ 所有目录 |
| `tenant` | 租户 | `mdc-01pri` / `public-dfau-trading-system` | ✅ 所有目录 |
| `service` | 服务类型 | `kafka` / `redis` / `etcd` / `nacos` | ✅ 所有目录 |
| `cluster` | 集群标识 | `mdc-01pri-kafka` / `dfid-ts-etcd` | ✅ midd/ 和 db/ |
| `hostname` | 主机名 | `prd-bss-rocket-rabbitmq-p1` | ✅ host/ 和 probe/ |
| `host_ip` | 主机IP | `10.1.87.58` | ✅ midd/ 和 db/ |
| `maintainer` | 维护团队 | `sre-t4` / `dba` | ✅ 所有目录 |

**自动注入标签（无需在 targets 里写）：**
- `project`：集群来源（由各 scrape job 的 relabel_configs 注入，或由联邦 target 文件定义。值为 `public-monitor` / `business-public` / `win-k8s`）
- `job`：由 VMAgent 自动添加（node-exporter / middleware_kafka / database_redis 等）
- `host_ip`：主机 IP（由 relabel_configs 从 `__address__` 中提取）
- `target_ip` / `target_port`：TCP 探测时自动提取（probe/ 目录）
- `domain` / `scheme`：HTTP 探测时自动提取（domain/ 目录）

**不同目录标签要求：**
- **`host/`**：`biz_project` + `tenant` + `service` + `hostname` + `maintainer`
- **`midd/`**（中间件）：`biz_project` + `tenant` + `service` + `cluster` + `hostname` + `host_ip` + `maintainer`
- **`db/`**（数据库）：`biz_project` + `tenant` + `service` + `cluster` + `hostname` + `host_ip` + `maintainer`
- **`probe/`**：`biz_project` + `tenant` + `service` + `hostname` + `maintainer`
- **`domain/`**：`biz_project` + `tenant` + `maintainer`（URL 已表明监控对象）
- **`kubernetes/`**：`biz_project` + `tenant` + `maintainer`

## 🔄 biz_project 索引

| biz_project | 说明 | 飞书告警群 |
|-------------|------|-----------|
| `ts` | 交易系统（Trading System） | feishu-ts |
| `cc` | 清算系统（Clearing Center） | feishu-cc |
| `mdc` | 行情数据中心（Market Data Center） | feishu-mdc |
| `cpp` | C++ 服务（含 blackarrow / rps / ocs / minirps / official-rps ） | feishu-cpp-rps / feishu-cpp-ocs / feishu-cpp-blackarrow |
| `metatradertools` | MetaTradeTools | feishu-metatradertools |
| `product-center` | 产品中心 | feishu-product-center |
| `infra` | 基础设施（K8s 系统组件 / 监控组件） | feishu-infra |

**tenant（租户）完整列表：**

| biz_project | tenant | 来源 | 飞书群 |
|-------------|--------|------|--------|
| `ts` | `public-dfid-trading-system` | targets/ | feishu-ts |
| `ts` | `public-dpsl-trading-system` | targets/ | feishu-ts |
| `ts` | `public-dfau-trading-system` | targets/ | feishu-ts |
| `ts` | `private-trading-system-dfid` | targets/ | feishu-ts |
| `ts` | `trading-system-dpen` | targets/ | feishu-ts |
| `cc` | `public-dfid-clearing-system` | targets/ | feishu-cc |
| `cc` | `public-dpsl-clearing-system` | targets/ | feishu-cc |
| `cc` | `cc-au-pri-v2` | targets/ | feishu-cc |
| `cc` | `cc-hk-v2` | targets/ | feishu-cc |
| `mdc` | `mdc-01pri` | targets/ | feishu-mdc |
| `mdc` | `mdc-02pri` | targets/ | feishu-mdc |
| `mdc` | `mdc-03pri` | targets/ | feishu-mdc |
| `mdc` | `public-mdc-01sec` | targets/ | feishu-mdc |
| `mdc` | `public-mdc-03sec` | targets/ | feishu-mdc |
| `mdc` | `intrade-datafeeds` | targets/ | feishu-mdc |
| `cpp` | `blackarrow` | targets/ + win-k8s | feishu-cpp-blackarrow |
| `cpp` | `rps` | targets/ | feishu-cpp-rps |
| `cpp` | `ocs` | targets/ + win-k8s | feishu-cpp-ocs |
| `cpp` | `minirps` | targets/ + win-k8s | feishu-cpp-rps |
| `cpp` | `official-rps` | targets/ + bigdata-flink + win-k8s | feishu-cpp-rps |
| `metatradertools` | `metatradertools` | targets/ | feishu-metatradertools |
| `metatradertools` | `mt-command-engine` | business-public | feishu-metatradertools |
| `product-center` | `product-center` | targets/ | feishu-product-center |

## 📝 Targets 文件格式

### node-exporter（主机监控）

`targets/mdc/mdc-01pri/host/kafka.yaml`

```yaml
# Service: kafka
- targets:
    - 10.10.10.12:9900
  labels:
    hostname: mdc01-kfk-p1
    biz_project: mdc
    tenant: mdc-01pri
    service: kafka
    maintainer: sre-t4
```

### redis database（数据库监控）

`targets/mdc/mdc-01pri/db/redis.yaml`

```yaml
# Redis Database Monitoring via redis_exporter (multi-target mode)
- targets:
    - redis://172.31.23.174:6379
  labels:
    hostname: ma03-rds-p1
    biz_project: mdc
    tenant: mdc-01pri
    service: redis
    cluster: mdc-01pri-redis
    maintainer: sre-t4
```

### etcd middleware（中间件监控）

`targets/ts/public-dfid-trading-system/midd/etcd.yaml`

```yaml
# Etcd Monitoring - Native Metrics on port 2379
- targets:
    - 10.1.87.58:2379
  labels:
    hostname: prd-dfid-ts-etcd-p1
    host_ip: 10.1.87.58
    biz_project: ts
    tenant: public-dfid-trading-system
    service: etcd
    cluster: dfid-ts-etcd
    maintainer: sre-t4
```

### blackbox-exporter（TCP 端口探测）

`targets/mdc/public-mdc-03sec/probe/mongodb.yaml`

```yaml
# MongoDB TCP Port Probe
- targets:
    - 10.1.80.40:27017
  labels:
    biz_project: mdc
    tenant: public-mdc-03sec
    service: mongo
    maintainer: sre-t4
    hostname: pc-arb02-mdb-p1
```

### blackbox-exporter（域名/HTTPS 监控）

`targets/mdc/public-mdc-01sec/domain/mdc-webapi.yaml`

```yaml
# 内网入口（使用集群 DNS）
- targets:
    - https://api-mdc-01sec.finpoints.tech
  labels:
    biz_project: mdc
    tenant: public-mdc-01sec
    maintainer: sre-t4
    network: internal

# 外网入口（使用公网 DNS 1.1.1.1 / 8.8.8.8）
- targets:
    - https://mdc-client01s.finpoints.com
  labels:
    biz_project: mdc
    tenant: public-mdc-01sec
    maintainer: sre-t4
    probe_module: http_non_5xx_public_dns
    network: external
```

**内外网双栈说明：**
- **无 `probe_module` 标签**：路由到 `blackbox-exporter:9115`（使用集群 DNS）
- **`probe_module: http_non_5xx_public_dns`**：路由到 `blackbox-exporter-external:9115`（使用公网 DNS）
- **`network` 标签**：区分内外网监控（`internal` / `external`）

## 🔍 Prometheus 查询示例

```promql
# 查询所有 kafka 服务
{service="kafka"}

# 查询特定集群
up{service="kafka", cluster="mdc-01pri-kafka"}

# 查询某个 tenant 的所有 redis
redis_up{biz_project="mdc", tenant="mdc-01pri", service="redis"}

# 区分多个 RabbitMQ 集群
rabbitmq_queue_messages{service="rabbitmq", cluster="rps-rabbitmq-01"}
rabbitmq_queue_messages{service="rabbitmq", cluster="rps-rabbitmq-02"}

# 按 maintainer 过滤
up{maintainer="sre-t4"}

# 组合查询：查询 ts 业务线的所有 etcd 集群
etcd_server_has_leader{biz_project="ts", service="etcd"}
```

## 📦 部署说明

### 前置准备

1. **创建 Git PAT（Personal Access Token）**
   - GitLab: User Settings -> Access Tokens -> `read_repository`

2. **创建 K8s Secret**

```bash
# Git PAT
kubectl -n prd-trading-infra-monitor create secret generic git-sync-token \
  --from-literal=token='glpat-xxxxxxxxxxxxxxxxxxxxx'

# VMS 认证
kubectl -n prd-trading-infra-monitor create secret generic vms-auth \
  --from-literal=username=admin \
  --from-literal=password='xxxxxxxxxxxxxxxxxxxxxxxxx'
```

### 部署组件

**监控集群（public-monitor）：**

```bash
# 1. VMAgent + git-sync（采集 + 同步）
kubectl apply -f kubernetes/public-monitor/deployment/vmagent-deployment.yaml

# 2. VMAlert（告警评估）
kubectl apply -f kubernetes/public-monitor/deployment/vmalert-deployment.yaml

# 3. Alertmanager（告警路由 + PVC 持久化静默/抑制状态）
kubectl apply -f kubernetes/public-monitor/deployment/alertmanager-deployment.yaml

# 4. alertmanager-webhook-feishu（飞书通知）
kubectl apply -f kubernetes/public-monitor/deployment/alertmanager-webhook-feishu-deployment.yaml
```

**公共业务集群（share-public-business）：**

```bash
# 5. VMAgent + git-sync（blackbox TCP 采集）
kubectl apply -f kubernetes/share-public-business/deployment/vmagent-deployment.yaml

# 6. Blackbox Exporter（TCP 探测，公共 + MDC 两套）
kubectl apply -f kubernetes/share-public-business/deployment/blackbox-exporter-deployment.yaml
```

> **Blackbox TCP 探测拆分说明：**
> 业务集群存在 MDC 独占节点（污点 `mdc-01-hk-sec` / `mdc-03-hk-sec` / `price-core`）和公共节点（污点 `internet` 或无污点），因此 blackbox-exporter 拆分为两套：
>
> | 组件 | 调度节点 | 容忍污点 | 探测目标 |
> |------|---------|---------|---------|
> | `blackbox-exporter` (replicas=3) | 公共节点 | `internet` | ts / cc / metatradertools / product-center / cpp |
> | `blackbox-exporter-mdc` (replicas=2) | MDC 独占节点 | `price-core` / `mdc-01-hk-sec` / `mdc-03-hk-sec` | mdc (01sec / 03sec) |

### 验证部署

```bash
# 查看 Pod 状态（监控集群）
kubectl -n prd-trading-infra-monitor get pod -l app=vmagent-file-sd
kubectl -n prd-trading-infra-monitor get pod -l app=vmalert
kubectl -n prd-trading-infra-monitor get pod -l app=alertmanager

# 查看 Pod 状态（业务集群 - blackbox TCP 探测）
kubectl -n prd-trading-infra-monitor get pod -l app=blackbox-exporter -o wide      # 公共
kubectl -n prd-trading-infra-monitor get pod -l app=blackbox-exporter-mdc -o wide  # MDC

# 查看 git-sync 日志
kubectl -n prd-trading-infra-monitor logs -l app=vmagent-file-sd -c git-sync --tail=50

# 查看 vmagent 日志
kubectl -n prd-trading-infra-monitor logs -l app=vmagent-file-sd -c vmagent --tail=50

# 确认 targets 已挂载
kubectl -n prd-trading-infra-monitor exec -it deploy/vmagent-file-sd -c vmagent -- \
  ls -R /sd/targets/
```

## 📊 vmagent 配置（scrape.yml）

vmagent 的 scrape 配置固定在 ConfigMap 中，只引用 file_sd。每个 job 通过 `relabel_configs` 注入 `project: public-monitor` 标签。

### Job 说明总览

| Job 名称 | 监控对象 | 文件路径 | metrics_path | 特殊配置 |
|---------|---------|----------|--------------|---------|
| `vmagent` | vmagent 自身 | static_configs | `/metrics` | - |
| `node-exporter` | 主机监控 | `host/*.yaml` | `/metrics` | 自动提取 host_ip |
| `middleware_kafka` | Kafka | `midd/kafka.yaml` | `/metrics` | - |
| `middleware_etcd` | Etcd | `midd/etcd.yaml` | `/metrics` | - |
| `middleware_rocketmq` | RocketMQ | `midd/rocketmq.yaml` | `/metrics` | - |
| `middleware_rabbitmq` | RabbitMQ | `midd/rabbitmq.yaml` | `/metrics` | - |
| `middleware_nacos` | Nacos | `midd/nacos.yaml` | `/nacos/actuator/prometheus` | 特殊 metrics_path |
| `database_clickhouse` | ClickHouse | `db/clickhouse.yaml` | `/metrics` | - |
| `database_redis` | Redis | `db/redis.yaml` | `/scrape` | multi-target 模式 |
| `database_elasticsearch` | Elasticsearch | `db/elasticsearch.yaml` | `/probe` | multi-target + auth_module |
| `database_mongodb` | MongoDB | `db/mongodb.yaml` | `/metrics` | scrape_interval: 30s |
| `k8s-federate` | K8s 集群 | `kubernetes/*.yaml` | `/federate` | honor_labels + metric_relabel |
| `blackbox-http` | 域名/HTTPS | `domain/*.yaml` | `/probe` | 双 DNS 路由 |

**业务集群 VMAgent（share-public-business）：**

| Job 名称 | 监控对象 | 文件路径 | metrics_path | 指向 blackbox |
|---------|---------|----------|--------------|--------------|
| `blackbox-tcp-mdc` | MDC TCP 探测 | `mdc/*/probe/*.yaml` | `/probe` | `blackbox-exporter-mdc:9115` |
| `blackbox-tcp-public` | 公共 TCP 探测 | `{ts,cc,metatradertools,product-center,cpp}/*/probe/*.yaml` | `/probe` | `blackbox-exporter:9115` |

> **注意**：监控集群配置以 `kubernetes/public-monitor/deployment/vmagent-deployment.yaml` 中的 ConfigMap 为准，业务集群配置以 `kubernetes/share-public-business/deployment/vmagent-deployment.yaml` 为准。

## 🔀 Job 名称统一（metric_relabel_configs）

VMS 汇聚了三个数据来源（`public-monitor`、`business-public`、`win-k8s`），不同来源的同类指标可能有不同的 job 名称。通过 metric_relabel_configs 统一：

| 原始 job（联邦源） | 统一为 | 说明 |
|-------------------|--------|------|
| `expose-node-metrics` | `node-exporter` | kube-prometheus-stack 默认名 |
| `expose-kubelets-metrics` | `kubelet` | kubelet 指标 |
| `expose-kubernetes-metrics` / `kubernetes` | `apiserver` | K8s API Server |
| `expose-kube-cm-metrics` | `kube-controller-manager` | Controller Manager |
| `expose-kube-etcd-metrics` | `kube-etcd` | Etcd 指标 |
| `expose-kubernetes-metrics` | `kube-state-metrics` | kube-state-metrics 指标 |
| `prometheus-io-scrape`（nginx_ingress_controller_*） | `ingress-nginx` | 按指标名拆分 |
| `prometheus-io-scrape`（coredns_*） | `node-local-dns` | 按指标名拆分 |

> **cAdvisor 重复采集说明**：kube-prometheus-stack 同时通过 kubelet `/metrics/cadvisor` 和独立 cadvisor DaemonSet 采集容器指标，导致指标重复。已在 `share-public-business-values.yaml` 的 `remoteWrite.writeRelabelConfigs` 中通过以下规则 drop 掉 cadvisor DaemonSet 自身的多余指标：
> ```yaml
> - sourceLabels: [service, metrics_path]
>   regex: "cadvisor;/metrics/cadvisor"
>   action: drop
> ```

## 🏷️ business-public 业务标签注入（writeRelabelConfigs）

业务集群（share-public-business）的 Prometheus 通过 `remoteWrite.writeRelabelConfigs` 在写入 VMS 时自动注入业务标签，配置文件为 `kubernetes/share-public-business/kube-prometheus-stack/share-public-business-values.yaml`。

### 注入逻辑

所有标签都基于 `namespace` 匹配，按以下优先级执行：

**1. `project` 固定注入**
```yaml
- targetLabel: project
  replacement: business-public
```

**2. `biz_project` 按 namespace 正则匹配**

| namespace 正则 | biz_project |
|---------------|-------------|
| `.*trading.*` / `public-d.*-dolphinscheduler` / `public-d.*-apisix-gateway` | `ts` |
| `metatradertools` | `metatradertools` |
| `mt-command-engine` | `metatradertools` |
| `pc-dolphinscheduler` / `product-center` | `product-center` |
| `.*clearing.*` / `public-d.*-makerv2` / `public-d.*-tradev2` / `public-d.*-takerv2` | `cc` |
| `public-mdc.*` | `mdc` |
| `default` / `ingress-nginx` / `kube-system` / `monitor` / `prd-trading-infra-monitor` / `velero` / `logging-vector` | `infra` |

**3. `tenant` 按 namespace 精确匹配**

| namespace 正则 | tenant |
|---------------|--------|
| `private-trading-system-dfid` | `private-trading-system-dfid` |
| `trading-system-dpen` | `trading-system-dpen` |
| `public-dfau-trading.*` | `public-dfau-trading-system` |
| `public-dfid-trading.*` | `public-dfid-trading-system` |
| `public-dpsl-trading.*` | `public-dpsl-trading-system` |
| `public-dfid-clearing.*` 等 | `public-dfid-clearing-system` |
| `public-dpsl-clearing.*` 等 | `public-dpsl-clearing-system` |
| `public-mdc01sec-.*` | `public-mdc-01sec` |
| `public-mdc03sec-.*` | `public-mdc-03sec` |

**4. `maintainer` 统一注入 `sre-t4`**（覆盖所有业务 namespace）

**5. MDC 独占节点标签覆盖**（优先级最高，兜底配置）

node-exporter 的 relabeling 通过节点 label（`mdc_01_hk_sec` / `mdc_03_hk_sec`）识别 MDC 独占节点，将 `biz_project` 和 `tenant` 覆盖为正确的 MDC 值，最后通过 `labeldrop` 清理中间标签 `_biz_project_raw` / `_tenant_raw`。

**6. Drop 规则**

| sourceLabels | 正则 | 说明 |
|-------------|------|------|
| `namespace` | `p4b.*\|website\|kibana\|odoo\|rdms-fp\|...` | 非业务 namespace 直接丢弃 |
| `service` + `metrics_path` | `cadvisor;/metrics/cadvisor` | 丢弃 cadvisor DaemonSet 重复指标 |

## 🏷️ bigdata-flink 集群标签注入（writeRelabelConfigs）

配置文件：`kubernetes/bigdata-flink/kube-prometheus-stack/prometheus-values.yaml`

| namespace 正则 | biz_project | tenant |
|---------------|-------------|--------|
| `official-rps` | `cpp` | `official-rps` |
| `default\|ingress-nginx\|kube-system\|monitor\|velero\|logging-vector` | `infra` | `infra` |

> 其余 namespace 的指标通过 `keep` 规则过滤，不写入 VMS。

## 🏷️ win-k8s 集群标签注入（inlineUrlRelabelConfig）

配置文件：`kubernetes/win-k8s/victoria-metrics-k8s-stack/win-k8s-values.yaml`

使用 VMAgent `inlineUrlRelabelConfig`，注入逻辑如下：

| namespace 正则 | biz_project | tenant |
|---------------|-------------|--------|
| `ba-mt5-gateway` | `cpp` | `blackarrow` |
| `minirps` | `cpp` | `minirps` |
| `official-rps` | `cpp` | `official-rps` |
| `ocs` | `cpp` | `ocs` |
| `mttools` | `cpp` | `ocs` |
| `default\|ingress-nginx\|istio-system\|kube-system\|velero\|monitoring` | `infra` | `infra` |

> Drop 规则：部分废弃节点通过 `instance`/`node` 标签过滤丢弃。

## 🚀 快速开始

### 1. 添加新的监控目标

```bash
# 1. 克隆仓库
git clone https://git.finpoints.tech/devops/monitoring-trading.git
cd monitoring-trading

# 2. 按规范创建/编辑 targets 文件
mkdir -p targets/mdc/public-mdc-01sec/db
vim targets/mdc/public-mdc-01sec/db/mysql.yaml

# 3. 提交并推送
git add targets/
git commit -m "Add MySQL targets for public-mdc-01sec"
git push origin main
```

### 2. 验证配置生效

提交后等待 **30s ~ 1min**，在 VictoriaMetrics 查询：

```promql
up{biz_project="mdc", tenant="public-mdc-01sec"}
```

### 3. 热更新机制

- **无需重启 vmagent / vmalert**
- git-sync 每 30 秒拉取一次
- vmagent file_sd 每 30 秒刷新一次
- VMAlert 自动重新加载规则文件
- **提交后最多 1 分钟自动生效**

## 🔐 安全建议

1. 使用只读 PAT（避免误操作篡改仓库）
2. PAT 存储在 K8s Secret（不要写在代码/ConfigMap 里）
3. 通过 GitOps 流程管理变更（PR review）
4. 敏感信息（密码/token）不要写在 targets 文件里

## 🚨 注意事项

- YAML 格式必须正确（建议本地用 `yamllint` 验证）
- 删除 target 前确认无关联告警规则
- 生产变更建议在维护窗口进行
- 告警规则修改后注意检查 VMAlert 日志确认加载成功

### 多集群联邦 PromQL 设计规范

本仓库的 VMS 汇聚多个集群的数据，编写告警规则时必须遵守以下规范，否则可能出现告警永远不触发或产生错误结果：

1. **所有聚合操作必须包含 `project`**：`sum/avg/max/min/count by(...)` 中必须含 `project`，防止跨集群数据合并。

2. **跨 exporter join 必须使用 `on()` 明确指定 join key**：不同 exporter（如 cAdvisor vs kube-state-metrics）的标签集不完全一致，隐式匹配会导致结果错误。join key 必须包含 `project`。
   ```promql
   # 正确
   cadvisor_metric / on(namespace, pod, container, project) group_left(biz_project, tenant) ksm_metric
   # 错误（可能产生 >100% 的假值）
   cadvisor_metric / ksm_metric
   ```

3. **禁止使用裸 `absent()`**：`absent()` 在多集群下只要任一集群有数据就不触发，应改为按 `project` 分组检测：
   ```promql
   # 正确
   max by(project, biz_project, tenant, maintainer)(up{job="xxx"}) == 0
   # 错误（任一集群存活则不告警）
   absent(up{job="xxx"} == 1)
   ```

4. **分母必须有 `> 0` 护卫**：除法运算的分母必须有护卫条件防止除零，特别是节点 swap/进程数/文件描述符等可能为 0 的指标。

5. **`rate()` 只能用于 counter 类型指标**：不要对 gauge 类型（如 `workqueue_depth`）使用 `rate()`，应使用 `avg_over_time()` 或直接使用 gauge 值。

## 🛠️ 故障排查

### targets 没有生效？

1. 确认 git-sync 是否正常同步：
   ```bash
   kubectl -n prd-trading-infra-monitor logs -l app=vmagent-file-sd -c git-sync --tail=50
   ```

2. 确认 vmagent 能读取到文件：
   ```bash
   kubectl -n prd-trading-infra-monitor exec -it deploy/vmagent-file-sd -c vmagent -- \
     ls /sd/targets/mdc/public-mdc-03sec/db/
   ```

3. 在 VMS 查询是否有数据：
   ```promql
   up{biz_project="mdc", tenant="public-mdc-03sec"}
   ```

### 告警没有发出？

1. 检查 VMAlert 是否加载了规则：
   ```bash
   kubectl -n prd-trading-infra-monitor logs -l app=vmalert --tail=50
   ```

2. 检查 Alertmanager 是否收到告警：
   ```bash
   kubectl -n prd-trading-infra-monitor port-forward svc/alertmanager 9093:9093
   # 访问 http://localhost:9093/#/alerts 查看当前告警
   ```

3. 检查 webhook-feishu 日志：
   ```bash
   kubectl -n prd-trading-infra-monitor logs -l app=alertmanager-webhook-feishu --tail=50
   ```

### git-sync 拉取失败？

- 检查 PAT 是否正确/过期：
  ```bash
  kubectl -n prd-trading-infra-monitor get secret git-sync-token -o yaml
  ```
- 检查 Git 仓库地址是否正确
- 确认网络策略允许访问 Git 服务器

## 📚 参考资料

- [VictoriaMetrics 文档](https://docs.victoriametrics.com/)
- [VMAgent 配置](https://docs.victoriametrics.com/vmagent/)
- [VMAlert 配置](https://docs.victoriametrics.com/vmalert/)
- [Alertmanager 配置](https://prometheus.io/docs/alerting/latest/configuration/)
- [Prometheus file_sd](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#file_sd_config)
- [git-sync](https://github.com/kubernetes/git-sync)

## 🤝 贡献指南

1. 遵循目录结构（`targets/<biz_project>/<tenant>/<resource-type>/`）
2. 遵循标签规范（必须包含 `biz_project` / `tenant` / `maintainer`）
3. 一个 PR 只修改相关业务线配置
4. 提交前用 `yamllint` 验证格式
5. 在 commit message 里说明变更原因
6. 告警规则修改需注明影响的规则数量和级别

## 📧 联系方式

- **维护团队**：SRE Team 4
- **问题反馈**：提交 Issue 或联系 `SRE Team 4`
