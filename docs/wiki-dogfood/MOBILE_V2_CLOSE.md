# MOBILE_V2 关刀纪要

> 2026-10-26 · 产品体验 · 狗粮 Wiki  
> **本刀：** 正常手机 App 新壳（废 Roku/货架；废旧 MOBILE_RESPONSIVE 当产品移动端）

## 目标

`frontend-mobile/`（Vite+React+TS）+ 共享 api-client；人端 `GET|PUT /me/session-context` 续现场；标准底栏/列表详情；能办事+好用顺手；IA 贴队席/Issue/Wiki/续现场。

## 人话验收

| 行 | 结果 |
| --- | --- |
| 窄视口登录→队席→Issue | 少步闭环 PASS† |
| session-context PUT / 覆盖 / PC 同账号 | PASS† |
| 登录后 hydrate 续现场 | PASS†（B2 复测 @`497f7a3`；不落 `/crew`） |
| 视觉 | V1 放行 · 正常手机浅色（非 Roku） |

## 锚

| 项 | SHA / 证据 |
| --- | --- |
| 债 / 架构 | `docs/NEXT_WAVE_DEBT_MOBILE_V2.md` |
| G1 session-context | `c9a6ac7` |
| R0/R1 壳 | `814245f` |
| R2 底栏+列表详情 | `11df4bc` |
| V0 三帧 | `fd2d788` |
| V1 放行 | `fd2d788`（债注 `243ecfe`） |
| B2 修 | `497f7a3` |
| Q1 初测 FAIL†（已清） | `0817676` / `38a9969` / 箱截 `d5bc922` |
| Q1 B2 复测 PASS† | `MOBILE_V2-MV2-Q1-retest-*` + `MOBILE_V2-MV2-Q1-retest-qa-20261026.txt` |
| P1 关单 | 本页 |

## 软非挡（另票 / 不挡关）

- 登录页上留白偏松  
- 队席副行 seat slug（A4 Soft†）  
- Capacitor 壳、未读计数 API —— **另票**

## 非目标（仍冻 / 未动）

Soft 三件；INVITE；LANDING（已另关）；人调 Bot handoff；Expo/RN/Flutter/Unity。
