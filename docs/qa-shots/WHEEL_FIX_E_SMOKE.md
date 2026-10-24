# WHEEL_FIX · E1–E2 全链路 smoke（QA）

> 权威：`docs/NEXT_WAVE_DEBT_WHEEL_FIX.md` @`8af223e`  
> 证据名：`docs/qa-shots/WHEEL_FIX-smoke-YYYYMMDD.txt`  
> **分段交到即苛刻，禁闲等。** 测绿直推 main。

## E1 覆盖链
登录 → WS → 队 → 看板 Issue → 归档 → **解档** → 切队 → Bot 席抽核 → 超管用户台

| 段 | 依赖 | 状态 |
|----|------|------|
| 登录/me | 现网 | 可先跑 |
| WS 开通/进看板 | A1/A2 | 可先跑；a3 去 seed 等 React |
| 队+看板 Issue | 现网 | 可先跑 |
| 归档 | ISSUE_ARCHIVE 已关 | 可先跑 |
| 解档 | **B1/B2** | 等交 |
| Relations 剔归档边 | **B3** | 等交 |
| 切队+Bot席+pins | C1–C3 | 可先回归 |
| 超管用户搜索/所属WS | **D1/D2** | 等交；现网超管台可轻抽 |

## E2
无阻塞 Defect；软注后置。
