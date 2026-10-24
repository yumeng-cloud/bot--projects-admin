# NEXT_WAVE · T7 最小 · 从模板创建团队

> 2026-10-24 · 总指挥自定（T6 P0.5 已关；用户禁闲等）  
> **本包：空席形状 + playbook_ref。不含静默覆盖 memory、不含自动镜像本机 Agent。**

## 目标

选班底手册 → 一键生成 `team` + N 空席（slug/显示名/`playbook_ref`），尚未绑 Key；同模板开第二队形状一致、记忆不串。

## 做 / 不做

| 做 | 不做 |
|----|------|
| API：创建团队时可带 `template_ref` / `playbook_ref`；`roster/ensure` 按模板席位表幂等建空席 | 静默覆盖 bot_memories |
| 内置或配置一份最小模板 JSON（席位表）于仓内 | 模板写回用户 git |
| FE：「从模板创建」最小向导：预览席位 → 创建 → 发入场码入口 | 升版 diff UI 全文（可只做 dry-run API stub） |
| 验收：空席形状与手册一致；memory 键不串队 | 自动把本机 Agent 镜像成席 |

## 验收

| ID | 标准 | 主跑 |
|----|------|------|
| **T7-A** | POST 带模板 → roster 席位数/slug 对齐；再 ensure 幂等 | Go + 记忆 |
| **T7-B** | 同模板第二队 → 席位形状同、pin/memory 不串 | 记忆/QA |
| **T7-C** | FE 向导可创建并进设置席位页 | React |
| **T7-D** | 证据 `docs/qa-shots/T7-template-create-*.txt` | QA |

## 主辅

Go/记忆主 · React 辅 · 种子文档 · QA · 产品勾

权威契约见 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) T7-1～T7-7。
