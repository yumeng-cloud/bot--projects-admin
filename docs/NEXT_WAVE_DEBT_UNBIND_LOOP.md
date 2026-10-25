# 解绑/结绑闭环 · UNBIND_LOOP

> 2026-10-25 08:30 Asia/Shanghai · 产品体验主持拍板  
> **承接：** 用户实操堵点（总指挥转）：后台已把 Bot 清完，再解绑仍提示「不能解绑 / 只能解绑自己绑定的席位」。要**整条使用闭环**先梳清再修；**QA 全链路自测过关再交，禁止再让用户测**。睡眠站立仍适用——不问用户。  
> **并行：** LOGIN_REFRESH 不挡本刀。  
> **仍冻：** Soft 三件（V2 静默补 claim / V6 多签 / 嵌套 revoke）。

## 根因（产品认定）

1. Admin `DELETE members`（设置·移除成员）**未级联**清 `team_seats.bound_member_id` / claim → 席上残留**指向已删 member 的脏绑**。  
2. 原 Bot 再 `seat leave` / `unbind`：`bound_member_id != 当前 member` → 硬 `403`「只能解绑自己绑定的席位」。  
3. 服务端对「席已无 bound」其实已幂等成功；堵点在**脏绑 + 人话硬拦**，不是缺 leave API。

## 状态机（目标态）

```
open ──(redeem/bind/enter)──► active(bound=member)
  ▲                              │
  │                              ├─(本人 leave/unbind)──► left(bound=∅) ──可补位/重绑
  │                              ├─(Admin 停工 stop)───► stopped(bound=∅) ──可补位
  │                              ├─(Admin 离职解绑)────► left(bound=∅)
  │                              └─(Admin 移除该 Bot 成员)─► **级联** left(bound=∅)+release claim
  │
left/stopped/open ◄── 补位仅显式补位码/合法 force（**不**自动；沿用任用冻 b6614ff）
```

任用硬规则（不变）：

- 活跃席 **禁**顶替；新人只进空席或新开席；仅 `stopped|left` 可补位。  
- 通码 **禁**静默补 stopped（V2 Soft 仍冻）。

## 拍死规则

| # | 规则 |
| --- | --- |
| **U1** | **谁绑：** redeem/bind/enter 成功的 member（持 `ymk_` 或当次 bootstrap）。 |
| **U2** | **谁解：** ① 绑定本人 `unbind`/`seat leave`；② Admin `stop`/`unbind`；③ Admin **移除成员**必须级联解席（见 U3）。 |
| **U3** | **Admin 移除成员 → 级联：** 凡 `bound_member_id = 该 member` 的席：清绑、`status=left`、release 其 active claim；人话审计 `seat.unbind` 或 `member.remove_cascade`。成员行删除成功与席清理同事务或等价原子。 |
| **U4** | **幂等离席：** 席已无 `bound` → `unbind`/`leave` **200 成功**（保持现状），CLI/FE 视为「已解除」。 |
| **U5** | **脏绑修复：** 席 `bound_member_id` 指向**已不存在的 member** → 视为失效绑定：`unbind`（本人鉴权或 Admin）**允许清绑成功**，人话「席位绑定已失效，已清理」；**禁止**再甩「只能解绑自己绑定的席位」。 |
| **U6** | **他人存活占用：** 席 bound 为**仍存在**的其他 member → 仍 `403`，人话改为：「席位由「{显示名}」占用；请对方离席，或让管理员停工/离职解绑」。 |
| **U7** | **本地收尾：** leave/unbind 成功或 U4/U5「已失效已清理」→ CLI/通道清本地队·席缓存；FE Toast 同句；**不要**卡死要求用户再点一次。 |
| **U8** | **重绑：** 席 `left|stopped|open` 且无 bound → 新人持合法码/空席路径进入；活跃仍禁顶替。 |

## 非目标

- 不解冻 V2 静默补 / V6 / 嵌套 revoke。  
- 不改入场主航线（HUMAN H0/H1）。  
- 不做「确认顶替」FE。  
- 登录页焕新另票并行。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **U-G1** | Admin 移除 Bot 成员级联清席+release；单测 | Go 后端 |
| **U-G2** | 脏绑（bound→已删 member）unbind 成功+人话；单测 | Go 后端 |
| **U-G3** | 他人存活占用：403 新人话（含显示名）；单测 | Go 后端 |
| **U-G4** | 无 bound 幂等 unbind 回归绿 | Go 后端 |
| **U-M1** | `ymbot seat leave`：U4/U5 当成功清本地；透传新人话 | Bot 记忆角色 |
| **U-R1** | 设置·移除成员成功后 roster/席卡即时无脏绑；错误 Toast 用人话 | React 前端 |
| **U-R2** | 席页 stop/unbind 文案与 U6 对齐（若有） | React 前端 |
| **U-Q1** | 全链路自测清单并自跑 PASS（禁甩用户） | QA 测试 |
| **U-P1** | 产品关 | 产品体验 |

## 产品勾选

- [ ] **U-G1** 移除级联
- [ ] **U-G2** 脏绑可清
- [ ] **U-G3** 占用中人话
- [ ] **U-G4** 幂等回归
- [ ] **U-M1** CLI 本地收尾
- [ ] **U-R1** FE 移除后席态
- [ ] **U-R2** FE 文案
- [ ] **U-Q1** QA 自测
- [ ] **U-P1** 关单

## QA 验收清单（U-Q1 必跑 · 禁甩用户）

1. Bot 入场绑席 → 本人 `seat leave` → 成功 → 席 `left` → 本地清。  
2. Bot 入场绑席 → Admin **停工** → Bot leave → 幂等成功。  
3. Bot 入场绑席 → Admin **离职解绑** → Bot leave → 幂等成功。  
4. Bot 入场绑席 → Admin **移除该成员** → 席已无脏绑；原钥/新身份再 leave → **不得**再出现「只能解绑自己…」；应为成功或「已失效已清理」。  
5. 席被**另一存活 Bot**占用 → 第三人 leave 该席 → 403 新人话含占用者显示名。  
6. Admin 清后再补位：仅 `stopped|left` + 合法补位路径；活跃禁顶替。  
7. 回归：SEAT_LEAVE Admin stop/unbind；SEAT_REFILL_BIND 补位码路径不坏。

## 派工

| 角色 | 首刀 |
| --- | --- |
| **Go 后端** | U-G1～U-G4（`MemberHandler.Remove` 级联 + `UnbindSeat` U5/U6） |
| **Bot 记忆角色** | U-M1（CLI leave 本地收尾+文案） |
| **React 前端** | U-R1～U-R2（移除后刷新席；Toast） |
| **QA 测试** | U-Q1 写清并自跑；出 spotcheck 纪要 |
| **产品体验** | 本裁定；U-P1 |
| **Issue 域** | 不动除非审计 comment 需顺手 |

## 进度注记

- 2026-10-25 08:30 · 产品拍板开刀。根因：移除成员不级联 → 脏绑 → 403「只能解绑自己绑定的席位」。LOGIN_REFRESH 并行。Soft 三件仍冻。
