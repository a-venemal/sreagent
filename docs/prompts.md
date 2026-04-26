# AI 提示词模板 (Prompts)

> 本文件提供结构化提示词模板，目标是最小化 token 消耗。
> 原则：**只给必要上下文，用路径代替内容，先方案后代码。**

---

## 对话规范（每次会话生效）

以下规则自动生效，无需每次重复：

1. **语言**：中文对话和交互
2. **上下文**：只加载 CLAUDE.md，按需读取 docs/ 和 MODULES.md
3. **代码引用**：用 `文件路径:行号`，不要粘贴大段代码
4. **方案优先**：复杂任务先给方案，确认后再写代码
5. **范围限定**：每次只改一个模块，不要跨模块批量修改
6. **完成标准**：`go build` 通过 + 更新 CHANGELOG.md

---

## 新增功能

### 模板

```
开发 [模块名] 的 [功能名]。

当前状态：参考 MODULES.md 中 [模块名] 的状态
目标：[一句话描述]
涉及文件：
- model: internal/model/xxx.go
- handler: internal/handler/xxx.go
- service: internal/service/xxx.go
- 前端: web/src/pages/xxx/Index.vue

约束：
1. 只改上述文件
2. 遵循 CLAUDE.md 代码规范
3. 完成后更新 CHANGELOG.md

请先给实现方案。
```

### 示例

```
开发 notification 的企业微信通知功能。

当前状态：参考 MODULES.md 中 notification 的状态
目标：新增 wechat_work 通知渠道类型
涉及文件：
- model: internal/model/notify_media.go（新增 MediaTypeWechatWork）
- service: internal/service/notify_media.go（新增 sendWechatWork）
- handler: 无需改动（复用现有 CRUD）
- 前端: web/src/pages/notification/Media.vue（表单新增类型选项）

约束：
1. 只改上述文件
2. 遵循 CLAUDE.md 代码规范
3. 完成后更新 CHANGELOG.md

请先给实现方案。
```

---

## 修复 Bug

### 模板

```
Bug: [一句话描述现象]
文件: [路径:行号]
错误: [粘贴错误日志，不要超过 10 行]

请：
1. 分析根因
2. 给修复方案
3. 修复
4. 确认 go build 通过
```

### 示例

```
Bug: 告警规则创建后分组等待时间不生效
文件: internal/service/alert_group.go:45
错误: group_wait_seconds 始终为 0，即使 DB 中有值

请：
1. 分析根因
2. 给修复方案
3. 修复
4. 确认 go build 通过
```

---

## 生成测试骨架

### 模板

```
为 [文件路径] 生成测试骨架。

要求：
1. 每个导出函数至少 1 个测试（正常 + NotFound）
2. 函数体用 t.Skip("TODO") 占位
3. 遵循 Test{Struct}_{Method}_{Scenario} 命名
4. 参考 docs/testing.md 的规范

输出：直接写入 [文件路径]_test.go
```

### 示例

```
为 internal/service/alert_rule.go 生成测试骨架。

要求：
1. Create、GetByID、Update、Delete、List 每个至少 2 个测试
2. 函数体用 t.Skip("TODO") 占位
3. 遵循 Test{Struct}_{Method}_{Scenario} 命名
4. 参考 docs/testing.md 的规范

输出：直接写入 internal/service/alert_rule_test.go
```

---

## 代码审查

### 模板

```
审查 [文件路径]。

检查项（按优先级）：
1. 安全漏洞（SQL 注入、XSS、硬编码密码）
2. 错误处理（未处理的 error、panic 风险）
3. 性能（N+1 查询、内存泄漏）
4. 规范一致性（是否遵循 CLAUDE.md）

输出：按严重程度排序，每项给 文件:行号 + 修改建议。
```

---

## 数据库迁移

### 模板

```
为 [功能名] 添加数据库迁移。

当前最高迁移号：[查看 internal/pkg/dbmigrate/migrations/]
表名：[table_name]
变更：[ADD COLUMN / CREATE TABLE / ...]

约束：
1. 单语句模式（每个 .sql 文件一条语句）
2. 文件名：{next_number}_{description}.{up|down}.sql
3. 同时生成 up 和 down
4. 更新 model/ 对应文件
```

---

## 前端页面

### 模板

```
开发 [模块名] 的前端页面。

参考：
- API: docs/api.md 中 [模块名] 的端点
- 类型: web/src/types/index.ts 中的 [TypeName]
- Composable: useCrudModal / usePaginatedList

目标：[描述页面功能]
涉及文件：
- web/src/pages/[module]/Index.vue
- web/src/api/index.ts（如需新增 API 调用）
- web/src/i18n/zh-CN.ts + en.ts（新增 i18n key）

约束：
1. 用 Naive UI 组件
2. 用 composable 减少重复代码
3. 禁止 v-html
```

---

## 性能优化

### 模板

```
优化 [模块/函数] 的性能。

当前问题：[描述性能瓶颈]
相关文件：[路径]
指标：[当前耗时/内存/查询次数]

请：
1. 分析瓶颈
2. 给优化方案（含预期收益）
3. 实施
4. 说明如何验证效果
```

---

## 上下文压缩（对话过长时）

```
当前对话上下文太长。请：
1. 总结已完成的工作（文件列表 + 变更摘要）
2. 列出待完成的任务
3. 生成 < 300 字的上下文摘要

我会在新对话中用这个摘要继续。
```

---

## 项目健康度检查

```
对项目进行健康检查：

1. 运行 go build && go vet，报告结果
2. 对照 MODULES.md，哪些模块缺测试？
3. 有没有死代码（未使用的导出函数/文件）？
4. docs/ 中的文档是否与代码一致？

输出：按优先级排序的修复建议。
```

---

## Token 节省技巧

| 做法 | 说明 |
|------|------|
| 用文件路径代替文件内容 | "参考 `internal/service/alert_rule.go:45`" 而不是粘贴代码 |
| 只给相关文件 | 不要整个项目丢给 AI，AI 会自己找依赖 |
| 模块化开发 | 每次只改一个模块 |
| 写清需求再开始 | 需求不清导致反复修改更费 token |
| 用 docs 代替口述 | "参考 docs/alert-engine.md" 比口述省 10 倍 token |
| 复用本文件模板 | 不要每次重写提示词 |
| 阶段性总结 | 完成一个功能后更新文档，下次不用重新解释 |
| 控制对话长度 | 超过 20 轮考虑用"上下文压缩"模板开新对话 |
