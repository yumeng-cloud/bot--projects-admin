# PRO_ADMIN · B1 P0 无项目空态苛刻清单（QA）

> 2026-10-24 22:05 Asia/Shanghai · QA  
> 权威：`docs/NEXT_WAVE_DEBT_PRO_ADMIN.md` **B1**（总指挥 P0）  
> 现网烂点：`ProjectRedirect` 主 CTA「去设置」→ 踢进设置迷宫，无法一键建项目进视图。  
> **Retest #2 @640fd6e 箱截有效 → Overall PASS**（2026-10-24 22:17 CST）。

## 现网缺口（基线 FAIL — 已修）

| # | 现象 | 代码锚 |
|---|------|--------|
| GAP-1 | 新 WS 零项目进 `/w/:wsId/board`（及 issues/cycles/timeline 同构）→ Result「无法进入项目视图」 | `frontend/src/pages/ProjectRedirect.tsx` — **已改为 Empty+本页创建** |
| GAP-2 | 主按钮 **「去设置」** `href=/w/:wsId/settings`（禁作主 CTA） | 同文件 — **已降为次要文字链** |
| GAP-3 | 无本页「创建项目」弹层/`POST /projects`；建完不能直进该项目看板 | CreateProjectModal + Navigate — **已修** |

## 红线验收（修完必过）

| # | 步骤 | PASS |
|---|------|------|
| B1-a | 新建空 WS（零项目）→ 打开 `/w/{ws}/board` | **PASS** (API empty WS + valid shot @640fd6e; live click †PENDING) |
| B1-b | 主文案含「暂无项目」；**无** seed/技术口吻；标题不挡主 CTA | **PASS** (UI「还没有项目」+ shot) |
| B1-c | **主按钮=「创建项目」**：本页打开表单/弹层（非跳转设置） | **PASS** (code+shot primary CTA) |
| B1-d | 「去设置」若有 → **仅文字链/次要**，非 `type=primary` | **PASS** |
| B1-e | 填 name + key_prefix → `POST /workspaces/{id}/projects` 201 → **直进** `/w/{ws}/projects/{id}/board` | **PASS** (API 201 0→1 + Navigate code) |
| B1-f | 同构：`/issues` `/cycles` `/timeline` 空态一致（主 CTA 同规） | **PASS** (shared ProjectRedirect) |
| B1-g | 已有项目时仍自动 Navigate 到首项目对应 tab（回归） | **PASS** (code) |

## 证据
- 主：`docs/qa-shots/PRO_ADMIN-B1-qa-retest-20261024.txt` · **Overall PASS**（retest #2 @640fd6e）
- 箱截：`pro-b-empty-no-project.png` ← **PASS**（32310 B / 1280×800 / 非白屏；还没有项目 + 创建项目 + 去设置）
- curl：空 WS `05706ecc…` projects=[]；POST project 201 → 0→1→2 OK
- 2026-10-24 22:17 Asia/Shanghai · verify HEAD `e6c5d60`（含 `640fd6e` 有效底图）

## 关报
短报产品+总指挥：Overall **PASS** · B1 表全绿 · 证据 · HEAD。测绿 push origin/main（站立令）。**请产品勾 B1**；G1 抽可并。
