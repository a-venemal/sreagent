# SREAgent UI 重构方案 — v2.0 "Aurora × Soft Futurism"

> 目标：把 SREAgent 的 UI 从"专业但常规"升级为**酷、深邃、丝滑、现代**，对标
> Linear / Vercel / Raycast / Arc Browser / Anthropic 官网 / visionOS 的视觉语言。
>
> 执行原则：**不推倒重来**。现有 design token 系统与 Naive UI 主题保留，在其上
> 叠加新层次。每一步都能独立上线验收，若方向不对随时回滚。

---

## 版本路线

| Step | 目标 | 规模 | 推荐模型 |
|------|------|------|---------|
| **Step 1** | 视觉底座：Aurora 背景、Spotlight、GlowCard、AnimatedNumber、登录页极光 | ~8 文件 | Sonnet 4.6 |
| **Step 2** | 结构重构：Icon Rail 侧栏、⌘K 命令面板、Bento 仪表盘 | ~12 文件 | Opus 4.7（⌘K/Bento），Sonnet（Icon Rail） |
| **Step 3** | 原生级动效：View Transitions API、磁吸按钮、3D tilt、滚动揭示 | ~6 文件 | Opus 4.7 |

---

## 视觉语言约束（三步通用）

- **暗色优先**：深度层次 `#05070a → #0b0e14 → #121722 → #192030` + 流动极光
- **品牌色对**：`#18a058`（翡翠绿）× `#06b6d4`（青蓝），高亮点缀 `#a78bfa`（紫）、`#f472b6`（粉）
- **圆角偏大**：card 16px / modal 22px / button 10px
- **玻璃层**：`backdrop-filter: blur(20px) saturate(160%)` + 半透明
- **动效品味**：出场 `cubic-bezier(0.22,1,0.36,1)` 320ms；强调 `cubic-bezier(0.34,1.56,0.64,1)` overshoot；hover 180ms
- **禁用原则**：不用强烈发光、不用彩虹渐变满铺、不用 emoji 装饰、不用 `v-html`
- **可访问**：`prefers-reduced-motion` 下所有装饰动画静止

---

# Step 1 — 视觉底座

**目标**：只动视觉层，结构不变。改完后用户一打开就感觉"**哇不一样**"。

## 1.1 新增设计 Token（`web/src/styles/global.css`）

在 `:root` 内追加：

```css
/* --- Aurora / 极光 --- */
--sre-aurora-1: #18a058;   /* 翡翠绿 */
--sre-aurora-2: #06b6d4;   /* 青蓝 */
--sre-aurora-3: #a78bfa;   /* 紫 */
--sre-aurora-4: #f472b6;   /* 粉 */
--sre-aurora-blur: 140px;
--sre-aurora-opacity: 0.38;

/* --- Conic border (rotating rainbow) --- */
--sre-conic-brand: conic-gradient(
  from var(--sre-conic-angle, 0deg),
  #18a058, #06b6d4, #a78bfa, #f472b6, #18a058
);

/* --- Spotlight cursor --- */
--sre-spotlight-size: 520px;
--sre-spotlight-opacity: 0.12;

/* --- Noise texture (inline SVG data URL) --- */
--sre-noise-url: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='200' height='200'><filter id='n'><feTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='2'/><feColorMatrix values='0 0 0 0 1  0 0 0 0 1  0 0 0 0 1  0 0 0 0.35 0'/></filter><rect width='100%' height='100%' filter='url(%23n)' opacity='0.5'/></svg>");

/* --- Glow shadows --- */
--sre-shadow-glow-critical: 0 0 0 1px rgba(239,68,68,0.4), 0 0 32px rgba(239,68,68,0.28);
--sre-shadow-glow-success:  0 0 0 1px rgba(16,185,129,0.4), 0 0 32px rgba(16,185,129,0.22);
--sre-shadow-soft-xl: 0 40px 80px -24px rgba(0,0,0,0.55), 0 8px 24px rgba(0,0,0,0.25);
```

在 `body.light-theme` 内追加：
```css
--sre-aurora-opacity: 0.22;
--sre-spotlight-opacity: 0.06;
```

## 1.2 新增关键帧与工具类（`global.css` 尾部）

```css
/* Aurora drift — 整个极光背景缓慢漂移 */
@keyframes sre-aurora-drift {
  0%, 100% { transform: translate3d(0, 0, 0) rotate(0deg); }
  33%      { transform: translate3d(3%, -2%, 0) rotate(3deg); }
  66%      { transform: translate3d(-2%, 2%, 0) rotate(-2deg); }
}

/* Conic border rotate — 驱动 --sre-conic-angle */
@property --sre-conic-angle {
  syntax: '<angle>';
  inherits: false;
  initial-value: 0deg;
}
@keyframes sre-conic-rotate {
  to { --sre-conic-angle: 360deg; }
}

/* 玻璃更强版（用于卡片） */
.surface-glass-strong {
  background: color-mix(in srgb, var(--sre-bg-card) 62%, transparent);
  backdrop-filter: saturate(170%) blur(22px);
  -webkit-backdrop-filter: saturate(170%) blur(22px);
  border: 1px solid var(--sre-border-strong);
}

/* Conic 渐变边框 */
.conic-border {
  position: relative;
  isolation: isolate;
}
.conic-border::before {
  content: '';
  position: absolute;
  inset: -1px;
  border-radius: inherit;
  padding: 1px;
  background: var(--sre-conic-brand);
  -webkit-mask:
    linear-gradient(#000 0 0) content-box,
    linear-gradient(#000 0 0);
  -webkit-mask-composite: xor;
          mask-composite: exclude;
  animation: sre-conic-rotate 8s linear infinite;
  opacity: 0.7;
  pointer-events: none;
  z-index: -1;
}
.conic-border--critical::before {
  background: conic-gradient(from var(--sre-conic-angle,0deg), #ef4444, #f59e0b, #ef4444);
  animation-duration: 4s;
  opacity: 1;
}

/* Noise overlay */
.noise-overlay {
  position: relative;
}
.noise-overlay::after {
  content: '';
  position: absolute;
  inset: 0;
  background-image: var(--sre-noise-url);
  opacity: 0.04;
  mix-blend-mode: overlay;
  pointer-events: none;
  border-radius: inherit;
  z-index: 1;
}
body.light-theme .noise-overlay::after { opacity: 0.02; }

/* 3D tilt 容器（JS 会写入 --tilt-x --tilt-y） */
.tilt {
  transform:
    perspective(800px)
    rotateX(calc(var(--tilt-y, 0) * 1deg))
    rotateY(calc(var(--tilt-x, 0) * 1deg));
  transition: transform 160ms var(--sre-ease-out);
  transform-style: preserve-3d;
  will-change: transform;
}
```

## 1.3 新增组件

### `web/src/components/common/AuroraBackground.vue`
- 固定定位 `position: fixed; inset: 0; z-index: -2; pointer-events: none;`
- 3 层 `<div>` 形成的 blur orb，每层 `filter: blur(var(--sre-aurora-blur))`
- 分别用 aurora-1/2/3 色，分别位于左上、右上、中下
- 每层 `animation: sre-aurora-drift 18s/22s/26s ease-in-out infinite`
- 暗色下 opacity 使用 `--sre-aurora-opacity`
- 最上层 `::after` 铺 noise texture（opacity 0.04）
- 单独一层 `linear-gradient(180deg, transparent 0%, var(--sre-bg-page) 85%)` 做底部渐隐

### `web/src/components/common/SpotlightCursor.vue`
- `position: fixed; pointer-events: none; z-index: -1;`
- 监听 `window` 的 `mousemove`，通过 CSS 变量 `--mx --my` 更新位置
- 背景：`radial-gradient(circle var(--sre-spotlight-size) at var(--mx) var(--my), color-mix(in srgb, var(--sre-brand-accent) 100%, transparent) 0%, transparent 60%)`
- 只在暗色模式显示（`body:not(.light-theme) &`）
- 支持 `prefers-reduced-motion` 关闭

### `web/src/components/common/AnimatedNumber.vue`
- props: `value: number`, `duration: number = 900`, `decimals: number = 0`, `suffix?: string`
- 使用 `requestAnimationFrame` 做 ease-out 缓动
- watch value 变化时重启动画
- 添加 `.number-display` class 保证字体对齐

### `web/src/components/common/GlowCard.vue`
- 替代 `n-card` 的高价值卡片包装
- props: `variant: 'default' | 'critical' | 'success'`, `interactive?: boolean`, `tilt?: boolean`, `glow?: boolean`
- 内置：
  - `.surface-glass-strong` 玻璃底
  - `.noise-overlay` 颗粒
  - 可选 `.conic-border` / `.conic-border--critical`
  - tilt: mousemove 计算 `--tilt-x --tilt-y` + 离开重置
  - glow: 根据 variant 用 `--sre-shadow-glow-critical/success`

## 1.4 集成

### `web/src/App.vue`
```vue
<template>
  <NConfigProvider ...>
    <AuroraBackground />
    <SpotlightCursor />
    <NMessageProvider ...>
      ...
    </NMessageProvider>
  </NConfigProvider>
</template>
```

### `web/src/layouts/MainLayout.vue`
- header 的 `.surface-glass` 改成 `.surface-glass-strong`
- sider 容器加 `noise-overlay` class
- logo 旁的品牌点加 `pulse` 呼吸动画

### `web/src/pages/dashboard/Index.vue`
- 4 个 stat card 改用 `<GlowCard>`：
  - 活跃告警卡：`variant="critical"` `glow` `tilt`
  - 其他三张：`variant="default"` `tilt`
- stat 数字改用 `<AnimatedNumber :value="..." />`
- section-title 上方加 `.eyebrow` overline 小字

### `web/src/pages/Login.vue`
- 移除原有两个 orb，改用 `<AuroraBackground />` 组件（复用）
- 登录卡片：`surface-glass-strong + conic-border + noise-overlay`
- 登录按钮：`tilt` 类 + `sre-shadow-glow-brand` hover
- logo 下方加 eyebrow 文字："SRE Alert Intelligence"

## 1.5 验收标准
1. 暗色模式打开仪表盘，能看到背景流动极光（不刺眼，低饱和）
2. 鼠标移动在暗色模式下有柔和光晕跟随
3. stat 卡数字刷新时有 tween 递增（不是瞬间闪）
4. 活跃告警卡周围有缓慢旋转的 conic 红橙描边
5. hover stat 卡有轻微 3D 倾斜（最大 6°）
6. 登录页像"太空仪表盘"而不是"后台登录"
7. `prefers-reduced-motion: reduce` 下所有装饰动画消失
8. `npm run build` 通过，无警告
9. Light mode 下极光极淡，不影响白底可读性

## 1.6 文件清单（Step 1）
- **修改**：
  - `web/src/styles/global.css`
  - `web/src/App.vue`
  - `web/src/layouts/MainLayout.vue`
  - `web/src/pages/Login.vue`
  - `web/src/pages/dashboard/Index.vue`
- **新增**：
  - `web/src/components/common/AuroraBackground.vue`
  - `web/src/components/common/SpotlightCursor.vue`
  - `web/src/components/common/AnimatedNumber.vue`
  - `web/src/components/common/GlowCard.vue`

---

# Step 2 — 结构重构（Icon Rail + ⌘K + Bento）

**目标**：让 UI 的「**信息架构与交互方式**」达到 2026 现代产品水平。

## 2.1 Icon Rail 侧栏

### 改造点
- 默认宽度 72px（只显图标 + label 小字，类似 Arc Browser / Linear）
- 悬停到菜单 group 上，右侧弹出 **flyout panel** 展示子菜单（绝对定位，玻璃底，250ms spring 入场）
- 点击 flyout 之外区域/ESC 关闭
- active 状态：左侧 3px 垂直 gradient 条 + 图标高亮
- 折叠/展开按钮改为图标排顶部的小型 toggle，保留用户可完全展开回 240px 的偏好
- 底部 persistent：版本号（小字，tertiary 色）+ 设置齿轮

### 文件
- 重构：`web/src/layouts/MainLayout.vue` 的 sider 部分
- 新增：`web/src/components/common/IconRailMenu.vue`
  - props: `options: MenuOption[]`, `activeKey: string`
  - events: `@select`
  - slot: `#item` 可自定义渲染
- 新增样式：`global.css` `.icon-rail` `.icon-rail-flyout` class

### 注意
- 移动端（< 768px）退回原 240px drawer 风格
- 保留键盘导航（方向键 + Enter）

## 2.2 ⌘K 命令面板（最高光功能）

### 交互规格
- 全局快捷键 `⌘K` / `Ctrl+K` 唤起
- 居中 modal，宽度 `min(640px, 90vw)`，`surface-glass-strong` + `conic-border`
- 顶部搜索框（大号字，autoFocus），无图标装饰，纯文本
- 下方按「**分组列表**」展示候选：
  - **Navigate** — 所有路由（规则、告警、静默、仪表盘…）
  - **Actions** — 新建规则、切换主题、切换语言、登出、复制当前页面链接
  - **Search** — 告警（动态 fetch，前 5 条）、规则（动态 fetch）
  - **Recent** — 最近打开的 3 条（localStorage）
- 键盘导航：上下箭头移动高亮，Enter 执行，Esc 关闭
- 模糊匹配：用 `fuse.js`（~8KB gzip，加入依赖）或自写 substring + 拼音首字母
- 结果项右侧显示 hint 标签（如 `Ctrl+1`、`Page`、`Action`）
- 无结果时显示"Try different keywords"提示

### 文件
- 新增：`web/src/components/common/CommandPalette.vue`
- 新增：`web/src/composables/useCommandPalette.ts`
  - 导出 `open()`, `close()`, `registerAction()`, `registerSearchProvider()`
  - 全局 `visible` ref
- 在 `App.vue` 挂载 `<CommandPalette />`（或 `MainLayout.vue`）
- `web/package.json` 增加依赖：`fuse.js@^7.0.0`

### i18n
- 新 key：`command.title`, `command.placeholder`, `command.navigate`, `command.actions`, `command.search`, `command.recent`, `command.noResults`

## 2.3 Bento Dashboard

### 布局
```
┌──────────────────────────────────────┬───────────────┐
│ HERO: 活跃告警大卡                   │ MTT:          │
│  - 总数 (AnimatedNumber 巨大号字)    │  MTTA/MTTR    │
│  - 按 severity 分布小条图            │  P50/P95      │
│  - 趋势 spark mini line              │  (tall card)  │
├─────────────┬──────────────┬─────────┤               │
│ DS count    │ Resolved 24h │ Rules   │               │
│ (small)     │ (small)      │ (small) │               │
├─────────────┴──────────────┴─────────┼───────────────┤
│ SEVERITY DONUT                       │ MTTR TREND    │
├──────────────────────────────────────┴───────────────┤
│ RECENT ALERTS TABLE                                  │
└──────────────────────────────────────────────────────┘
```

CSS Grid：
```css
grid-template-columns: 1.2fr 1fr 1fr 1fr;
grid-template-rows: 220px 120px 280px auto;
grid-template-areas:
  "hero hero hero mtt"
  "s1 s2 s3 mtt"
  "donut donut trend trend"
  "recent recent recent recent";
gap: 20px;
```

响应式：< 1280px 退化为 2 列；< 768px 单列。

### 文件
- 改写：`web/src/pages/dashboard/Index.vue`
- 新增：`web/src/pages/dashboard/widgets/HeroAlertsCard.vue`
- 新增：`web/src/pages/dashboard/widgets/MttMetricsCard.vue`
- 新增：`web/src/pages/dashboard/widgets/SeverityDonut.vue`
- 新增：`web/src/pages/dashboard/widgets/MttrTrend.vue`
- 新增：`web/src/pages/dashboard/widgets/RecentAlertsTable.vue`

每个 widget 都是 `<GlowCard>` 包裹。

### 验收
1. 仪表盘一眼抓住重点（Hero 卡占视觉主体）
2. 每张卡尺寸不同，有节奏感，不再是"四个一样大的 stat 格"
3. 响应式折叠优雅

## 2.4 Step 2 文件清单
- **修改**：
  - `web/src/layouts/MainLayout.vue`
  - `web/src/App.vue`
  - `web/src/pages/dashboard/Index.vue`
  - `web/src/styles/global.css`
  - `web/src/i18n/en.ts`, `zh-CN.ts`
  - `web/package.json`
- **新增**：
  - `web/src/components/common/IconRailMenu.vue`
  - `web/src/components/common/CommandPalette.vue`
  - `web/src/composables/useCommandPalette.ts`
  - `web/src/pages/dashboard/widgets/` 5 个 widget 文件

---

# Step 3 — 原生级动效

**目标**：让每一次交互都"**爽**"。

## 3.1 View Transitions API

- 在 `App.vue` 的 `<router-view>` 外层用 `document.startViewTransition()` 包裹路由切换
- 退回到当前 Vue `<transition>` 作为 Firefox/Safari 的 fallback（Feature detect `'startViewTransition' in document`）
- CSS：
  ```css
  ::view-transition-old(root),
  ::view-transition-new(root) {
    animation-duration: 320ms;
    animation-timing-function: var(--sre-ease-out);
  }
  ::view-transition-old(root) {
    animation-name: sre-vt-fade-out;
  }
  ::view-transition-new(root) {
    animation-name: sre-vt-fade-in;
  }
  ```
- 给关键元素加 `view-transition-name`，让它们跨页持续：
  - PageHeader: `view-transition-name: page-header`
  - Hero dashboard card → alert detail hero: `view-transition-name: hero-<id>`

### 文件
- `web/src/App.vue`（router guard 注入 startViewTransition）
- `web/src/styles/global.css`（加 VT 样式）

## 3.2 磁吸按钮指令

### 文件：`web/src/directives/magnetic.ts`
```ts
export default {
  mounted(el: HTMLElement) {
    const rect = () => el.getBoundingClientRect()
    const onMove = (e: MouseEvent) => {
      const r = rect()
      const x = (e.clientX - r.left - r.width / 2) * 0.18
      const y = (e.clientY - r.top - r.height / 2) * 0.18
      el.style.transform = `translate(${x}px, ${y}px)`
    }
    const onLeave = () => { el.style.transform = '' }
    el.addEventListener('mousemove', onMove)
    el.addEventListener('mouseleave', onLeave)
    el.__magnetic__ = { onMove, onLeave }
  },
  unmounted(el: any) {
    el.removeEventListener('mousemove', el.__magnetic__.onMove)
    el.removeEventListener('mouseleave', el.__magnetic__.onLeave)
  },
}
```

- 在 `main.ts` 注册：`app.directive('magnetic', magnetic)`
- 使用：`<n-button v-magnetic type="primary">...</n-button>`
- 应用到：登录按钮、仪表盘 CTA、⌘K 打开按钮等关键 CTA

## 3.3 3D Tilt（已在 Step 1 GlowCard 预埋）
Step 3 精修：
- 使用 `mousemove` + `requestAnimationFrame` 节流
- 离开 card 时 reset 用 spring（`transition: transform 400ms cubic-bezier(0.34,1.56,0.64,1)`）
- 跨所有 `GlowCard` 打开 tilt 开关

## 3.4 Scroll Reveal

### 文件：`web/src/composables/useScrollReveal.ts`
- 自动 IntersectionObserver
- 观察带 `data-reveal` 属性的元素
- 元素进入视窗 + 60% 时切 `data-reveal="in"`
- CSS:
  ```css
  [data-reveal="out"] { opacity: 0; transform: translateY(24px); }
  [data-reveal="in"]  {
    opacity: 1; transform: translateY(0);
    transition: opacity 500ms var(--sre-ease-out), transform 500ms var(--sre-ease-spring);
  }
  ```
- 在长页面（告警规则列表、告警历史、仪表盘 chart section）用

## 3.5 Skeleton Shimmer 精修

已有 `.shimmer`，新增组合：
- `<SkeletonCard>` 组件：3 行灰块 + 圆形头像占位，整体扫光
- 替换所有 `n-spin` 全局 mask 为 skeleton（更现代）
- 文件：`web/src/components/common/SkeletonCard.vue`

## 3.6 Step 3 文件清单
- **修改**：
  - `web/src/App.vue`
  - `web/src/main.ts`
  - `web/src/styles/global.css`
  - `web/src/components/common/GlowCard.vue`（tilt 精修）
- **新增**：
  - `web/src/directives/magnetic.ts`
  - `web/src/composables/useScrollReveal.ts`
  - `web/src/components/common/SkeletonCard.vue`

---

# 验证流程（每步结束）

```bash
# 本地开发
cd web && npm run dev      # http://localhost:3000

# 类型 & 构建
cd web && npm run build    # 必须通过，无 warning

# Docker 全量构建（上 tag 前）
docker build -f deploy/docker/Dockerfile -t sreagent:uinext .

# 回归：至少验
# 1. 登录页动画
# 2. 仪表盘加载首屏
# 3. 深色/浅色切换
# 4. 菜单折叠展开
# 5. prefers-reduced-motion 模式
```

# Commit 规范

```
feat(ui): Step 1 — Aurora visual foundation
feat(ui): Step 2 — Icon Rail sidebar + Command Palette + Bento dashboard
feat(ui): Step 3 — View Transitions + magnetic + tilt + scroll reveal
```

每步打独立 tag：`v1.8.0-ui.1`, `v1.8.0-ui.2`, `v1.8.0-ui.3`，最后合并为 `v1.8.0`。

---

# 参考灵感

| 产品 | 借鉴点 |
|------|--------|
| Linear | 键盘优先、克制色彩、速度感 |
| Vercel Dashboard | Aurora 背景、Bento 网格 |
| Raycast | ⌘K 命令面板交互细节 |
| Arc Browser | Icon Rail 侧栏 + flyout |
| Anthropic.com | 深色极光 hero、微粒纹理 |
| visionOS | 玻璃层叠、柔光、大圆角 |
| Framer 官网 | 3D tilt、磁吸、丝滑缓动 |
