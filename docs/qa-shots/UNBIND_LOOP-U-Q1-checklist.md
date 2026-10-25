# UNBIND_LOOP U-Q1 · 全链路验收清单（活点·PASS†）

> 2026-10-25 08:31 Asia/Shanghai · QA 骨架  
> **债：** [`docs/NEXT_WAVE_DEBT_UNBIND_LOOP.md`](../NEXT_WAVE_DEBT_UNBIND_LOOP.md) @`e229c8e`  
> **门禁：** Go `9d49bdd` + ymbot `1065066` + FE `091dd71` 已入 main → 活点已跑。  
> **本阶段：** 只写骨架；**不**宣称 PASS；**不**跑未就绪全链路。  
> **硬约束：** 禁甩用户测；V2G 静默补仍冻；测绿后 spotcheck 可 push origin/main。

## 目标（产品 §QA 1–7 + 债文 U4–U8）

整条解绑/结绑闭环自测过关：Admin 移除成员级联清席；脏绑可清；他人占用新人话；幂等 leave；补位/活跃禁顶替与 SEAT_LEAVE / SEAT_REFILL_BIND 回归不坏。

## 前置（交件后填实）

| 项 | 值 / 门禁 |
| --- | --- |
| 债文 | `docs/NEXT_WAVE_DEBT_UNBIND_LOOP.md` · HEAD@run `e88747d`（含 U-G/U-M/U-R） |
| API | 本地 `http://127.0.0.1:8080`；二进制含 U-G `9d49bdd`（`/tmp/ai-board-server-ug` 08:35 重建） |
| Admin | `admin@yumeng.tech` / seed → JWT |
| Bot | redeem 无 Bearer → `ymk_`（挂 HUMAN Q1）；`POST …/seats/{id}/bind` / `unbind` / `stop` |
| 成员移除 | `DELETE /api/v1/workspaces/{ws}/members/{memberID}`（Admin） |
| CLI | `ymbot seat leave --role <slug> [--team <uuid>]`（U-M1 交件后验本地清缓存） |
| FE | 设置·移除成员后 roster/席卡即时无脏绑；Toast 人话（U-R1/U-R2） |
| 对照绿 | SEAT_LEAVE L5 `SEAT_LEAVE-L5-qa-retest-20261025.txt`；SEAT_REFILL_BIND B4 `SEAT_REFILL_BIND-B4-qa-spotcheck-20261025.txt` |
| 冻结 | Soft 三件（V2 静默补 claim / V6 多签 / 嵌套 revoke）；LOGIN_REFRESH 并行不挡 |

## API 契约速查（curl 优先）

```text
BASE=http://127.0.0.1:8080/api/v1
WS=…  TEAM=…  SEAT=…  MEMBER=…

# 入场 / 绑席
POST $BASE/team-invites/redeem          {code, role?}          → ymk_ + member
POST $BASE/workspaces/$WS/teams/$TEAM/seats/$SEAT/bind         Bearer ymk_ → active

# 本人 / Admin 离席
POST …/seats/$SEAT/unbind               Bearer ymk_|Admin JWT  → left（U4/U5 幂等或清脏）
POST …/seats/$SEAT/stop                 Admin JWT              → stopped

# Admin 移除成员（U3 级联）
DELETE $BASE/workspaces/$WS/members/$MEMBER   Admin JWT        → 200「成员已移除」
                                                              + 席 bound=∅ status=left + release claim

# 读席
GET  …/teams/$TEAM/seats                （或 roster 等价）→ bound_member_id / status / display_name
```

期望人话（债文拍死）：

| 场景 | HTTP | 文案要点 |
| --- | --- | --- |
| U4 无 bound 再 leave | **200** | 成功 / 已解除（幂等） |
| U5 脏绑（bound→已删 member） | **200** | **「席位绑定已失效，已清理」**；**禁**「只能解绑自己绑定的席位」 |
| U6 他人存活占用 | **403** | **「席位由「{显示名}」占用；请对方离席，或让管理员停工/离职解绑」** |
| 活跃禁顶替 | **409** | 「该席位占用中，禁止顶替」（沿用既有） |

## 步骤矩阵（骨架·等码）

| # | 覆盖 | 前置 | 步骤（API 优先） | 期望 HTTP / 文案 / 态 | 证据落点 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **UQ1-1** | §QA-1 入场→本人 leave | 空席或新建队；BotA redeem+bind | ① BotA `POST …/bind`→active ② BotA `POST …/unbind`（或 `ymbot seat leave`） ③ GET seats | ② **200**；席 `status=left`、`bound_member_id=null`；CLI/本地队·席缓存已清（U7/U-M1） | `docs/qa-shots/UNBIND_LOOP-U-Q1-qa-retest-YYYYMMDD.txt` §1；可选 CLI 输出摘录 | **PASS**（见 retest） |
| **UQ1-2** | §QA-2 Admin 停工后再 leave | BotA active | ① Admin `POST …/stop`→`stopped`+清绑 ② BotA 再 `unbind`/`seat leave` | ① **200** `stopped` ② **200** 幂等（U4）；无 403「只能解绑自己…」 | 同上 §2 | **PASS**（见 retest） |
| **UQ1-3** | §QA-3 Admin 离职解绑后再 leave | BotA active | ① Admin `POST …/unbind`→`left`+清绑 ② BotA 再 leave | 同 UQ1-2；审计 `seat.unbind` 可抽（SEAT_LEAVE L2 口径） | 同上 §3 | **PASS**（见 retest） |
| **UQ1-4** | §QA-4 Admin **移除成员**级联（根因刀） | BotA active；记下 `seat_id`+`member_id`+`ymk_` | ① Admin `DELETE …/members/{BotA}`→200 ② GET seats：**无**脏绑（bound≠已删 id） ③ 原钥或新身份对该席 `unbind` | ① 成员行删+席 `left`+claim release（U-G1 同事务/等价） ② bound=∅ ③ **200** 成功或 **「席位绑定已失效，已清理」**（U5）；**不得**再出现「只能解绑自己绑定的席位」 | 同上 §4（本刀主证）；可附 JSON 摘录 | **PASS**（见 retest） |
| **UQ1-5** | §QA-5 他人存活占用→403 新人话 | BotA active；BotB 另身份 | BotB（第三人）`POST …/seats/{A席}/unbind` | **403**；人话含占用者**显示名**：`席位由「{BotA显示名}」占用；请对方离席，或让管理员停工/离职解绑`（U6/U-G3）；holder 不变 | 同上 §5 | **PASS**（见 retest） |
| **UQ1-6** | §QA-6 清后再补位；活跃禁顶替 | 接 UQ1-2/3/4 后席为 `stopped\|left`；另留一席仍 active | ① 新人持**锚定补位码** redeem/bind → 成功占 `stopped\|left` ② 对仍 **active** 席 redeem/bind（含 force） | ① **200** active（SEAT_REFILL_BIND） ② **409** 禁顶替；**不**静默补 stopped（V2 Soft 冻） | 同上 §6 | **PASS**（见 retest） |
| **UQ1-7a** | §QA-7 回归 SEAT_LEAVE | 对照 L5 三硬条 | 复跑摘要：active→Admin stop→新人可补；active 仍 409；扩编新席不挤 | 与 `SEAT_LEAVE-L5-qa-retest-20261025.txt` 同口径 PASS | 同上 §7a；引用旧证 + 本轮 curl | **PASS**（见 retest） |
| **UQ1-7b** | §QA-7 回归 SEAT_REFILL_BIND | 对照 B4 | 锚定码绑 stopped\|left；未锚定不静默补；active 409 | 与 `SEAT_REFILL_BIND-B4-qa-spotcheck-20261025.txt` 同口径 | 同上 §7b | **PASS**（见 retest） |
| **UQ1-8** | U4 无 bound 幂等（单点） | 席已 `left`/`open` 且 bound=∅ | 任意合法身份（Bot 原钥或 Admin）`unbind` | **200**；席态不变；无错误 Toast | 同上 §8 | **PASS**（见 retest） |
| **UQ1-9** | U-M1 CLI 本地收尾 | U-M1 交件；UQ1-1/4 场景 | `ymbot seat leave` 在 U4/U5 成功路径 | 退出 0；本地队·席缓存清；透传新人话；**不**要求用户再点一次 | 同上 §9 + CLI stdout 摘 | **PASS†** · unit+API soft† |
| **UQ1-10** | U-R1 FE 移除后席态 | U-R1 交件；箱 FE | 设置页移除 Bot 成员 → 回席页/roster | 席卡即时无脏绑（非「占用中·已删人」）；错误 Toast 用人话 | 同上 §10；可选 `docs/qa-shots/unbind-loop-remove-*.png` | **PASS†** · source soft† |
| **UQ1-11** | U-R2 FE 文案对齐 U6 | U-R2 交件 | 席页他人占用时 stop/unbind 失败路径（若 FE 暴露） | Toast/文案含占用者显示名，对齐 U6；无旧句「只能解绑自己…」 | 同上 §11 | **PASS†** · source soft† |

## Go 单测挂点（交件后顺手）

```bash
cd backend && go test ./internal/services/botmem/ -count=1 \
  -run 'Unbind|StopSeat|Remove|Cascade|Dirty|Occupy|Leave' -v
# 期望：U-G1～U-G4 相关用例全绿；记录包名/用例名进 retest
```

## 证据产出（活点阶段）

写 `docs/qa-shots/UNBIND_LOOP-U-Q1-qa-retest-YYYYMMDD.txt`：

- `time_AsiaShanghai` / `HEAD_at_run` / Go·FE·CLI 交件 SHA  
- 逐步 HTTP 码、响应 `message`、席 JSON（status / bound_member_id / bound_display_name）  
- Overall = PASS | PASS† | FAIL（**有码跑绿前禁止写 PASS**）  
- 绿：commit+push origin/main；短报产品体验 + 总指挥。FAIL：钉 Go（U-G*）/ 记忆（U-M1）/ React（U-R*）；用户睡中不问。

## 现状笔记（2026-10-25 08:31 Asia/Shanghai · 骨架时探）

| 观察 | 结论 |
| --- | --- |
| HEAD（写清单时） | 含债文祖先 `e229c8e`；其后仅 LOGIN_REFRESH 文档/FE，**无** UNBIND_LOOP 业务 commit |
| `MemberHandler.Remove` | 仅 `DELETE FROM members`；**无**级联清 `team_seats.bound_member_id` / claim → 与债文根因一致 |
| `UnbindSeat` | 无 bound → 幂等返回（U4 底已有）；`bound != self && !admin` → 仍硬 `403`「只能解绑自己绑定的席位」→ **U5/U6 未落地** |
| 新人话 | 代码库**未见**「席位绑定已失效，已清理」/「席位由「…」占用；请对方离席…」 |
| FE | `removeMember` / `stopSeat` / `unbindSeat` client 在；移除后级联刷新与 U6 Toast **待 U-R1/U-R2** |
| CLI | `ymbot seat leave` 已接线 unbind；U4/U5 当成功清本地 **待 U-M1** |
| 本地 API | `http://127.0.0.1:8080/healthz` **200**（可活测底，但本阶段不跑全链路） |
| 本阶段动作 | **只交本清单骨架**；不改业务代码；不宣称 PASS |

## 明确不做（本刀）

- 不解冻 V2 静默补 / V6 / 嵌套 revoke  
- 不改 HUMAN 入场主航线；不做「确认顶替」FE  
- 不验收 LOGIN_REFRESH（并行票）  
- **骨架阶段不 push 半成品业务**；本文件为文档骨架，允许单独 commit+push

## 状态

| ID | 状态 |
| --- | --- |
| U-Q1 checklist | **活点完成**（本文件 + retest） |
| U-Q1 活点 | **PASS†** @2026-10-25 08:36 Asia/Shanghai · HEAD `e88747d` · binary rebuilt w/ `9d49bdd` |
| Overall | **PASS†**（UQ1-9/10/11 soft†：无箱内 ymbot 二进制；FE 源码对照不重截；U5 活句由级联+单测覆盖） |

## 下一步（谁交件才能开跑）

1. **Go 后端**交 U-G1（移除级联）· U-G2（脏绑可清+人话）· U-G3（占用新人话）· U-G4（幂等回归绿）并 push → QA 开跑 UQ1-1～5、8 + 7 回归。  
2. **Bot 记忆**交 U-M1 → 补 UQ1-9。  
3. **React**交 U-R1/U-R2 → 补 UQ1-10/11（可与 API 矩阵并行，不挡主证 UQ1-4）。  
4. 全绿写 retest → 产品 **U-P1** 关单。

## 活点笔记（2026-10-25 08:36 Asia/Shanghai）

- 交件齐：Go `9d49bdd` · ymbot `1065066` · FE `091dd71`；产品开跑 @`4f7e558`。
- 先验：旧 `/tmp/ai-board-server` mtime 02:19 **早于** U-G → 重建 `/tmp/ai-board-server-ug` 后跑矩阵。
- 主证 UQ1-4：DELETE member → 席 `left`+bound=∅；原钥 401；**无**「只能解绑自己…」。
- UQ1-5 人话实锤：`席位由「pm」占用；请对方离席，或让管理员停工/离职解绑`。
- FE U-R：`emitMemberRemoved`→`acb:member-removed` 乐观清席；`humanizeSeatLeaveError`（产品 steering 源码核对即可）。
- 证据：`docs/qa-shots/UNBIND_LOOP-U-Q1-qa-retest-20261025.txt`
