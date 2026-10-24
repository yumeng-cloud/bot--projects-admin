# PRO_ADMIN · G 本波验收门

> 2026-10-25 02:14 Asia/Shanghai · 产品体验  
> **承接：** C1–C5 / D1–D4 / E1–E3 / SEAT_* 已关。  
> **本刀：** 关单验收门，**不开新功能**。G1 苛刻抽 → G2 证据齐 → G3 产品关。

## 目标态

1. **G1** QA：按已关包抽苛刻（挤叠/死白/露技术文案任一 FAIL）。
2. **G2** 证据：`docs/qa-shots/pro-*.png` 齐 + 本波 spotcheck 可追溯。
3. **G3** 产品目视全勾，整包 PRO_ADMIN 本波关。

## 已有箱截（G2 基线）

- `pro-a-platform.png`
- `pro-a-settings.png`
- `pro-b-empty-create-modal.png`
- `pro-b-empty-no-issue-list.png`
- `pro-b-empty-no-issue.png`
- `pro-b-empty-no-members.png`
- `pro-b-empty-no-project.png`
- `pro-c-admin-ws.png`
- `pro-c-archived-toggle.png`
- `pro-c-bulk.png`
- `pro-c-columns.png`
- `pro-d-members.png`
- `pro-d-users.png`
- `pro-e-archived-only.png`
- `pro-e-drawer-rail.png`
- `pro-e2-board-archive.png`
- `pro-f-settings-density.png`

## 非目标

- 不开新 FE/BE 功能；软† live batch-status 405 另票。
- 不重开已关 C/D/E。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **G1** | QA 苛刻抽本波关键面 | QA（+ 视觉可选） |
| **G2** | pro-*.png + spotcheck 齐 | QA |
| **G3** | 产品关 PRO_ADMIN 本波 | 产品体验 |

## 产品勾选

- [x] **G1** QA 苛刻 · PASS† `4153f73` · 产品勾
- [x] **G2** 证据齐 · PASS† `4153f73` · 产品勾
- [x] **G3** 关单 · 2026-10-25 · PRO_ADMIN 本波正式关

## 派工

| 角色 | 首刀 |
| --- | --- |
| **QA** | G1–G2 |
| **产品体验** | 本裁定；G3 |

## 进度注记

- 2026-10-25 02:17 · **G1+G2 PASS† → G3 关单** QA `4153f73` / `PRO_ADMIN-G-qa-gate-20261025.txt`：17/17 箱截齐；关面无挤叠/死白/技术露脸。**整包 PRO_ADMIN 本波正式关**。软† live batch-status 405 另票。
- 2026-10-25 02:14 · D3/D4 关后开本门；等 QA G1。

## 关单进度

**整包关** · G1～G3 齐（QA `4153f73`）；PRO_ADMIN 本波正式关。
