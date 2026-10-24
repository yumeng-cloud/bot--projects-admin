# 冷灰清爽气质 · 体验验收（P0）

> 状态：**C1–C5 全部已勾关单**（产品体验验收 2026-10-24 · Asia/Shanghai）  
> 用户反馈：背景不高档、不清爽——是**整体设计**（背景分层 / 宽窄节奏 / 字号），不是单点。  
> 仓库：`/workspace/ai-collab-board`

---

## 与既有文档的关系

| 文档 | 职责 | 本改是否动 |
|------|------|------------|
| [`UI_BREATHING_P0.md`](./UI_BREATHING_P0.md) | **呼吸感尺寸**：sider 240、content padding、rail 312、drawer 940、项高 36… | **保留**，不改尺寸 token |
| [`UI_MODERN_P0.md`](./UI_MODERN_P0.md) | **去老气**：细边框、淡品牌选中、禁蓝块粗边条、近无阴影、U5 滚动 | **仍有效**；本改只换冷灰分层与字 |
| **本文 `UI_CLEAN_P0`** | **气质色 + 字体**：冷灰画布 / 侧栏 chrome / 白工作面；Inter + 中文系统栈 | 权威色板 |

一句话：**尺寸跟呼吸，气质换冷灰分层。**

---

## 目标色板（light）

| Token | 值 | 用途 |
|-------|-----|------|
| `--acb-bg-layout` | `#F7F8FA` | 画布 / 主区底 |
| `--acb-bg-sider` | `#F2F4F8` | 侧栏表面（与主区白分层） |
| `--acb-bg-container` | `#FFFFFF` | 容器 / 卡 / 工作面 |
| `--acb-bg-elevated` | `#FCFCFD` | 抬升面（表头、轻面板） |
| `--acb-border` | `#E6E8EE` | 主边框 |
| `--acb-border-secondary` | `#EEF0F4` | 次级更淡冷边 |
| `--acb-text` | `#2A3140` | 正文 |
| `--acb-text-secondary` | `#4A5263` | 次级 |
| `--acb-text-tertiary` | `#8A93A6` | 三级 / meta |
| `--acb-primary` | `#5E6AD2` | 主色（保持）；选中/chip = 低饱和淡靛 |

禁止：厚阴影、彩虹卡、Ant 实心蓝块选中。

### 旧 → 新（light 对比）

| 角色 | 旧 | 新 |
|------|----|----|
| 画布 | `#F7F8F8` | `#F7F8FA` |
| 侧栏 | `#FFFFFF`（与主区同白） | `#F2F4F8` |
| 容器 | `#FFFFFF` | `#FFFFFF` |
| 抬升 | （无 / 白） | `#FCFCFD` |
| 主边框 | `#E4E5E9` | `#E6E8EE` |
| 次边框 | `#EEF0F3` | `#EEF0F4` |
| 正文 | `#1A1C23` | `#2A3140` |
| 次级字 | `#5C6370` | `#4A5263` |
| 三级字 | `#8B929E` | `#8A93A6` |
| 主色 | `#5E6AD2` | `#5E6AD2`（不变） |

---

## 字体

```css
--acb-font-sans: Inter, -apple-system, BlinkMacSystemFont, "Segoe UI",
  "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
```

- 入口：`frontend/index.html` Google Fonts Inter 400/500/600/700（preconnect + css2）
- `body` 与 `linearTheme.token.fontFamily` 均用 `var(--acb-font-sans)`
- 层级：标题 **15–16** / 正文 **13–14** / meta **12**；标题略紧 `letter-spacing`

---

## 分层原则

```
侧栏冷灰 chrome (#F2F4F8)
  └── 纯白工作面 (#FFFFFF) 叠在冷灰画布 (#F7F8FA) 上
```

Ant Layout：`siderBg` → `#F2F4F8`，`bodyBg` / `colorBgLayout` → `#F7F8FA`，`headerBg` → `#FFFFFF`。  
CSS：`.ant-layout-sider { background: var(--acb-bg-sider) }`。

---

## 必改文件

1. `docs/UI_CLEAN_P0.md`（本文）
2. `frontend/src/index.css` — 全部 `--acb-*` 色；sider 背景；`--acb-font-sans`
3. `frontend/src/theme/linearTheme.ts` — token 对齐；Layout siderBg/bodyBg/headerBg
4. `frontend/index.html` — Inter（已挂则勿重复打架）
5. `board.css` / `timeline.css` / `issue-detail.css` — 若写死旧灰则跟变量（当前已多用 `var(--acb-*)`）

---

## 验收（产品体验 · C1–C5）

- [x] **C1** 画布 `#F7F8FA`
- [x] **C2** 侧栏冷灰分层 `#F2F4F8`
- [x] **C3** Inter 字阶
- [x] **C4** 对比干净
- [x] **C5** 与呼吸感尺寸一体

### 落地核对（UI 视觉）

- [x] 色板写入本文并落到 `:root` + `linearTheme`
- [x] 侧栏表面 `#F2F4F8`，主区白工作面分层
- [x] 呼吸尺寸未改（240 / padding / rail / drawer）
- [x] `cd frontend && pnpm build` 通过（2026-10-24 · Asia/Shanghai）


---

## 勾选证据（2026-10-24 · Asia/Shanghai · 产品体验）

| ID | 结果 | 证据 |
|----|------|------|
| **C1** | 通过 | `:root --acb-bg-layout: #F7F8FA`；`linearTheme` `colorBgLayout` / Layout `bodyBg` |
| **C2** | 通过 | `--acb-bg-sider: #F2F4F8`；`.ant-layout-sider { background: var(--acb-bg-sider) }`；`siderBg` token |
| **C3** | 通过 | `index.html` Google Fonts Inter；`--acb-font-sans`；`body` + `linearTheme.fontFamily` |
| **C4** | 通过 | 正文 `#2A3140` / 次 `#4A5263` / 三级 `#8A93A6`；边框 `#E6E8EE`；近无阴影保留 |
| **C5** | 通过 | 呼吸感尺寸变量未回退窄轨（sider 240、content 28×20、rail 312、drawer 940） |

