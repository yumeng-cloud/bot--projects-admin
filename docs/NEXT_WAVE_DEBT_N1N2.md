# 下一波可演示 · 记债收口 N1/N2

> 2026-10-24 17:31 Asia/Shanghai · 产品体验定范围  
> 前提：Nits PASS；D1–D5 `4318273`；D7-min 回归 `0a0cdd2` PASS；GET invites `6750013`  
> **不做**：包装头/列模板再打磨（④ 与 Nits 已关）；J7 巡检例程/「立即推送」队列（另波）

---

## N1 · J1 入场码吊销（优先 · 可演示「收回码」）

| | |
| --- | --- |
| **谁** | **Go 后端** → 挂路由；**React 前端** → 设置班底列表动作 |
| **现状** | `botmem.RevokeInvite` + DB `revoked_at` 已有；**HTTP 未挂**；FE 已有 GET 末四列表，**无吊销钮** |
| **Go** | `POST /api/v1/workspaces/{ws}/teams/{team}/invites/{inviteID}/revoke`（Admin only；幂等已吊销→200/404 人话二选一，推荐已吊销仍 200 + status=revoked）；API.md 补一行；单测 + 箱 curl |
| **React** | `revokeTeamInvite` client；Bot 设置已签发列表每行：末四 + 态芯片 + **吊销**（Popconfirm）；吊销后列表变 `revoked`，再 redeem 人话「已吊销」 |
| **验收** | 箱 5174：签发→列表见 active 末四→吊销→态=已吊销；旧码兑换失败人话；截 `N1-revoke-list` |
| **成功** | J1 记债「HTTP/FE 未挂」清掉；产品勾 N1 |

## N2 · J7 下发 FE 最小（席位条「下发」）

| | |
| --- | --- |
| **谁** | **React 前端**（Webhook 主路径已 PASS，本波不改 Go 扇出） |
| **一句话** | 设置/班底 **占用中**席位行：次要动作「下发任务」→ 打开建 Issue（或抽屉），**assignee 预填该席 `bound_member_id`**，标题可空待填；保存即指派。**不做**「立即推送」按钮、不做巡检 UI |
| **验收** | 箱 5174：对占用席点下发→新建 Issue 负责人=该席 Bot 成员；保存后 Issue 列表可见；截 `N2-dispatch-seat` |
| **成功** | J7 记债「下发 FE」收一刀；巡检/立即推送仍 P0.5 另单 |

## 暂缓

| 项 | 原因 |
| --- | --- |
| 包装打磨 | P-pack / P-col + Nits 已 PASS，重开=空转 |
| J7 巡检例程 | 通道/例程域，非本波 FE 演示 |

## 派发

| # | 项 | 席位 |
| --- | --- | --- |
| 1 | N1 吊销 HTTP | **Go 后端** |
| 2 | N1 吊销 UI | **React 前端**（等路由或并行 mock 后接真） |
| 3 | N2 下发 FE | **React 前端**（可与 N1 UI 同波） |
