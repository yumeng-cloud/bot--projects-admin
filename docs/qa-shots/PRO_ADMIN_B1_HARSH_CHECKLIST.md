# PRO_ADMIN · B1 P0 无项目空态苛刻清单（QA）

> 2026-10-24 22:05 Asia/Shanghai · QA  
> 权威：`docs/NEXT_WAVE_DEBT_PRO_ADMIN.md` **B1**（总指挥 P0）  
> 现网烂点：`ProjectRedirect` 主 CTA「去设置」→ 踢进设置迷宫，无法一键建项目进视图。  
> **等 React 热修后立刻验**；勿盲测。

## 现网缺口（基线 FAIL）

| # | 现象 | 代码锚 |
|---|------|--------|
| GAP-1 | 新 WS 零项目进 `/w/:wsId/board`（及 issues/cycles/timeline 同构）→ Result「无法进入项目视图」 | `frontend/src/pages/ProjectRedirect.tsx` |
| GAP-2 | 主按钮 **「去设置」** `href=/w/:wsId/settings`（禁作主 CTA） | 同文件 `extra` |
| GAP-3 | 无本页「创建项目」弹层/`POST /projects`；建完不能直进该项目看板 | — |

## 红线验收（修完必过）

| # | 步骤 | PASS |
|---|------|------|
| B1-a | 新建空 WS（零项目）→ 打开 `/w/{ws}/board` | |
| B1-b | 主文案含「暂无项目」；**无** seed/技术口吻；标题不挡主 CTA | |
| B1-c | **主按钮=「创建项目」**：本页打开表单/弹层（非跳转设置） | |
| B1-d | 「去设置」若有 → **仅文字链/次要**，非 `type=primary` | |
| B1-e | 填 name + key_prefix → `POST /workspaces/{id}/projects` 201 → **直进** `/w/{ws}/projects/{id}/board` | |
| B1-f | 同构：`/issues` `/cycles` `/timeline` 空态一致（主 CTA 同规） | |
| B1-g | 已有项目时仍自动 Navigate 到首项目对应 tab（回归） | |

## 证据
- 主：`docs/qa-shots/PRO_ADMIN-B1-qa-retest-YYYYMMDD.txt`
- 箱截：`pro-b-empty-no-project.png`（对齐债 B4；本项可先单张）
- 可选 curl：建项目前后 `GET …/projects` 条数 0→1

## 关报
短报产品+总指挥：Overall · B1 表 · 证据 · HEAD。测绿 push origin/main（站立令）。产品勾 B1 后可并 G1 抽。
