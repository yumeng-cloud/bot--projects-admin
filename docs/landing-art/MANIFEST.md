# LANDING GenerateImage 真图 · 还原包

> 供其他 AI / 同事按文件名还原插画。气质与禁项以 [`../LANDING_TRUE_ART_PROMPTS.md`](../LANDING_TRUE_ART_PROMPTS.md) 为准。  
> 落盘目录：`frontend/src/pages/Landing/assets/`（覆盖同名 PNG）。  
> **真图必须是 GenerateImage（或同等插画引擎）**；禁止 HTML/CSS→PNG、色块拼版冒充。

## 文件 ↔ 提示词

### 1. `landing-hero-board.png` · 16:9

```
Premium Chinese B2B SaaS landing hero illustration for product 育蒙协作看板, Feishu/Notion homepage quality. Layered composition: foreground human and agent role chips as one crew; midground three workstation nodes merging into a unified scheduling hub; soft depth; accent #5E6AD2; light gray #F2F4F8; polished marketing art not a raw screenshot. MUST show human+Agent one team and multi-machine scheduling. FORBIDDEN: plain three-column kanban screenshot only, teaching classroom, homework, school, Feishu logo, dark neon cyber.
```

### 2. `landing-why-compare.png` · 4:3

```
Split comparison: left chaotic overlapping chat windows; right calm seat roster crew panel. Soft flat Feishu-Notion style, purple #5E6AD2, light background. No classroom, no dark neon.
```

### 3. `landing-team-board.png` · 4:3

```
Humans and simple agent avatars collaborating around one shared workspace (board fragment + context), product 育蒙协作看板, quiet premium marketing illustration, light theme #5E6AD2. Not a full kanban screenshot dump. No school theme, no Feishu logo.
```

### 4. `landing-multi-machines.png` · 16:9

```
Three labeled workstations 审图 写码 主持 flowing into hub 统一调度. Soft flat premium SaaS marketing, light, #5E6AD2. Distributed crew not session handoff. No neon.
```

### 5. `landing-seats-keys.png` · 4:3

```
Professional seat occupancy table with human/agent tags plus one on-demand key card. Clean Feishu-Notion settings aesthetic, light, #5E6AD2. No QR wall, no classroom seating chart.
```

### 6. `landing-flow-steps.png` · 16:9

```
Horizontal four-step premium process: 建班底 → 各机认领 → 看板推进 → 统一调度. Soft line, numbered, #5E6AD2, light marketing infographic. No textbook timeline.
```

## 共用风格锁

| 要 | 禁 |
| --- | --- |
| 浅色 `#F2F4F8` / 白表面；主色 `#5E6AD2` | 暗黑、径向 glow、赛博霓虹 |
| GenerateImage 插画级软扁平 / 轻等距营销合成 | HTML/CSS→PNG、色块拼版、平面信息图冒充插画 |
| 专业协作产品感（班底、席位、调度） | 教学作业皮（教与学/老师/课件/作业） |
| 副题：人与 Agent 一支队 · 多机分工 · 统一调度 | 飞书 / Lark / Notion 品牌 Logo |

## 验收指针

- Hero 第一眼：敢用的产品品牌页，不是弱三列看板演示
- 叙事可读：人+Agent / 多机 / 统一调度
- 专册全文：`docs/LANDING_TRUE_ART_PROMPTS.md` · 视觉基线：`docs/SITE_LANDING_VISUAL_BASELINE.md`
