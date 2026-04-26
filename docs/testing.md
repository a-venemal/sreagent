# 测试策略

## 测试金字塔

```
        /  E2E  \          少量，覆盖核心流程
       / 集成测试 \         中量，覆盖模块间交互
      /  单元测试   \       大量，覆盖核心逻辑
```

## 框架

- **Go**: `testify` (assert/require) + 标准 `testing`
- **前端**: Vitest + Vue Test Utils（待建立）
- **E2E**: Playwright（待建立）

## 运行方式

```bash
# Go 单元/集成测试
SREAGENT_TEST_DSN="user:pass@tcp(localhost:3306)/sreagent_test" make test

# 前端测试（待建立）
cd web && npx vitest run
```

## 测试分层

### 后端（每个模块）

| 文件 | 覆盖内容 |
|------|----------|
| `service/*_test.go` | 业务逻辑：CRUD、校验、边界条件 |
| `handler/*_test.go` | HTTP 接口：正常响应、400/404/403、请求体校验 |
| `engine/*_test.go` | 引擎：状态机转换、指纹生成、抑制逻辑 |
| `repository/*_test.go` | 数据层：SQL 正确性、约束、事务 |

### 前端（每个模块）

| 文件 | 覆盖内容 |
|------|----------|
| `*.test.ts` | API 调用 mock、响应处理 |
| `*.spec.vue` | 组件渲染、事件触发、props 传递 |

## 测试数据库

- 独立测试数据库 `sreagent_test`
- 环境变量 `SREAGENT_TEST_DSN` 控制
- 每个测试用 `CleanupDB()` 清理
- 未设置 DSN 时自动 skip（不影响 CI）

## 测试命名规范

```go
// 格式: Test{Struct}_{Method}_{Scenario}
func TestAlertChannelService_Create_Success(t *testing.T) { ... }
func TestAlertChannelService_Create_DuplicateName(t *testing.T) { ... }
func TestAlertChannelService_GetByID_NotFound(t *testing.T) { ... }
```

## 覆盖目标

| 优先级 | 模块 | 目标覆盖率 |
|:------:|------|:----------:|
| P0 | engine/ (状态机、抑制) | 80% |
| P0 | service/notification.go | 70% |
| P1 | service/ (所有 CRUD) | 60% |
| P1 | handler/ (所有接口) | 50% |
| P2 | repository/ | 40% |
| P2 | 前端组件 | 30% |

## 测试骨架生成

对任何新模块，先生成骨架再实现：

```bash
# 为 service/alert_rule.go 生成测试骨架
# 参考 docs/prompts.md 中的"生成测试骨架"模板
```
