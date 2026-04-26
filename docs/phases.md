# Phase 追踪 & QA 修复汇总

> 从 CHANGELOG.md 分离，记录内部开发阶段和质量修复。

---

## Phase 记录

| 阶段 | 内容 | 状态 |
|------|------|:----:|
| Phase 0 | Cleanup（删除遗留文件、修复 Dockerfile/K8s 配置） | ✅ |
| Phase 1 | CI/CD 完整文档 | ✅ |
| Phase 2 | Redis 引擎状态持久化 | ✅ |
| Phase 3 | Keycloak OIDC + RBAC 权限 | ✅ |
| Phase 4 | 核心模块补完（Subscribe/Notify 管道） | ✅ |
| Phase 5 | 前端 UI 全面改版（7 个子阶段） | ✅ |
| Phase 6 | API 文档（120+ 端点） | ✅ |
| Phase 7 | QA 多角色验证（14 后端 + 16 前端修复） | ✅ |
| Phase 8 | 上下文压缩 + 文档更新 | ✅ |

## QA 修复汇总 (Phase 7)

### 后端（14 项）
- RequireRole / GetCurrentUserID 不安全类型断言 → comma-ok
- ChangePassword 修改了管理员自己的密码 → 改用 URL :id 参数
- OIDC callback 无 CSRF state 验证 → 添加 state cookie 验证
- OIDC Secure cookie flag 硬编码 → 从 TLS/X-Forwarded-Proto 推导
- OIDC JWT 通过 query param 传递 → 改为 URL fragment
- Redis 在 HTTP Server 之前关闭 → 调换顺序
- zap.Fatal 在 goroutine 中阻止优雅关闭 → zap.Error + os.Exit
- StateEntry 缺少 Annotations → 添加字段

### 前端（16 项）
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
