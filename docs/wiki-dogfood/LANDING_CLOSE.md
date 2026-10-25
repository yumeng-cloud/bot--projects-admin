# LANDING / SITE_LANDING 关刀纪要

> 2026-10-25 · 产品体验 · 狗粮 Wiki

## 目标

未登录默认页 = 完整官网（安静好用、非赛博）；登录后进后台；`/join` 深链不破。气质与锁文案见债册；视觉跟专册。

## 人话验收

| 行 | 结果 |
| --- | --- |
| 未登录 `/` | 完整 Landing（非 sole Login 墙） |
| 登录后 | 进后台 Home（未卡官网） |
| `/join` | 深链可达、不破 |
| 多机叙事 | 正文含认领 + 统一调度；换机续聊仅否定句 |
| 眉题 / 主标 / CTA | 纯文眉题禁 pill · 育蒙协作看板 · 登录 |

## 关键

| 项 | SHA / 证据 |
| --- | --- |
| FE L-R1/R2/R3 加厚 | `b7b20ea`（CSS 旁证 `083c3cc`） |
| SL-V1/V2 箱截 + SL-V3 总监放行 | `e027f5a` |
| 勾债 SL-V1/V2/V3 | `b680ed8` |
| L-Q1 PASS† | `35cd80a` · 纪要 `LANDING-L-Q1-qa-20261025.txt` |
| L-Q1 箱截 | `LANDING-L-Q1-home-unauth-20261025.png` · `LANDING-L-Q1-home-scroll-20261025.png` · `LANDING-L-Q1-after-login-20261025.png` · `LANDING-L-Q1-join-20261025.png` |
| L-P1 关单 | 本页 |

## 软非挡（另票）

全局 Inter CDN（Landing 页未外挂；全站既有）——**另票**，不挡本刀关单。

## 非目标（仍冻）

Soft 三件不解冻；不改 INVITE_ON_DEMAND / SEAT_* 业务逻辑；不做定价/注册漏斗；不重做 LOGIN_REFRESH。
