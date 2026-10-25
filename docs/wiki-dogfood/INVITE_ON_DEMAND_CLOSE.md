# INVITE_ON_DEMAND 关刀纪要

> 2026-10-25 · 产品体验 · 狗粮 Wiki

## 目标

入场码是钥匙不是海报——要一把才开一把。禁建队/全班默认堆 `ymt_`；列表禁裸 `team:uuid:slug`；有 K 查用、无则生成、在席禁顶替。

## 人话验收

| 行 | 结果 |
| --- | --- |
| 建队默认 | 不自动弹堆码 / 不 auto issueCrewPack |
| 按需一码 | 显式发一把 → 仅 1 枚；禁静默加堆 / `invites/batch` |
| 列表 | 人话席名；脏串永不渲染 |
| K 状态机 | 有 `ymk_` → 查用；有 `ymt_` → 复用；无 → 签发一枚 |
| 活跃席 | 禁 mint 顶替（UI 藏入口 · Go **409**） |
| 全班包 | 二级 + 勾选空席确认后再签（不物理删） |

## 锚

| 项 | SHA / 证据 |
| --- | --- |
| I-M1 通道 | `22cc85e` |
| I-R1～R4 FE | `b0a4d3d` |
| I-G1 Go | `a034a15` |
| Soft† IQ1-3/5 | `1257830` |
| Lookup UI | `7f1e576` |
| 在席查用（IQ1-4 B） | `d99e14b` |
| I-Q1 PASS†（IQ1-1～6 + API） | `5a9fc06` · 纪要 `INVITE_ON_DEMAND-I-Q1-qa-20261025.txt` |
| I-Q1 箱截（节选） | `INVITE_ON_DEMAND-I-Q1-retest-IQ1-3-20261025.png` · `…-IQ1-4-ymk-d99e14b-20261025.png` · `…-IQ1-4-ymt-20261025.png` · `…-IQ1-4-pm-create-20261025.png` · `…-IQ1-5-20261025.png` |
| I-P1 关单 | 本页 |

## 非目标（仍冻）

Soft 三件不解冻；不改 SEAT 三态 Tag / SEAT_REFILL 补位主路径；不新开 `invites/batch`；不实现角色锚模型档位配置。
