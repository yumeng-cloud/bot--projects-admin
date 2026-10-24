# HUMAN_ONBOARD · QA 用例补严（Q1–Q10）

> 2026-10-24 23:42 Asia/Shanghai · QA（复验更新见下）  
> **复验：** 2026-10-24 23:52 Asia/Shanghai · QA · Overall **PASS†** · HEAD 含 `c4ae0a5`+`662346d` · 证 `HUMAN_ONBOARD-Q1-Q3-Q12-qa-retest-20261024.txt`  
> **权威骨架：** [`docs/NEXT_WAVE_DEBT_HUMAN_ONBOARD.md`](../NEXT_WAVE_DEBT_HUMAN_ONBOARD.md) @`6eaa4ff` **§6**（Q1–Q10 为规范 ID，勿另起平行编号）。  
> **产品 H3：** QA §6 补严文件名 = **本路径** `docs/qa-shots/HUMAN_ONBOARD-QA-CASES.md`。  
> **人用硬钉：** Admin 只发 `ymt_` 入场包；Agent 入队/入席自动身份（当次一把 `ymk_`）；活跃禁顶替；新人空席/新席；停工才补位；**否决全班吐 ymk_**。

## S ↔ Q 映射（简）

| 场景（§2） | 主用例（§6） | 注 |
| --- | --- | --- |
| **S1 新入队** | **Q1** · **Q10** | 无钥 join → 自动身份；公开材料禁嵌钥 |
| **S2 续队** | **Q2** | 同 ymk_ 再 join 不重发钥、不丢 memory |
| **S3 补位** | **Q5** | 仅 left/stopped；审计；禁 active force |
| **S4 分布式派发** | **Q8** | 指派只叫醒绑席 Bot；left 不喊 |
| **S5 换机** | **Q9** | 同钥 handoff get；不靠 Admin 再吐钥 |
| **S6 建队组班** | **Q6** · **Q7** | AUTO_CREW 入场包；无全班身份钥匙入口 |
| **S7 扩编新人新席** | **Q4** | 无 empty + allow_expand → 新席；不挤人 |
| **S8 引用任用** | **Q3** · **Q5** | active→409；规则 FE/通道一致 |

横切：**入场自动身份** = Q1/Q2/Q10；**禁全班吐 ymk_** = Q6/Q7。吊销/双钥心智详见附录（非 §6 主列）。

## 状态图例

| 状态 | 含义 |
| --- | --- |
| **COVERED†** | 已有箱证 PASS/PASS†，可对账；†=软项（Modal 点击/产网 DNS）不挡硬门槛 |
| **GAP** | 产品规则已写，缺独立箱证或未挂进本债验收 |
| **BLOCKED** | **目标态未实装**（如无身份 join 自动 mint）；H0+H1 未开干前不硬验绿 |

---

## Q1 新入队自动身份

| 项 | 内容 |
| --- | --- |
| **前置** | 默认班底有 `status=open` 空席；Agent **无**有效 `ymk_`（或 bootstrap 无 Bearer）；Admin 已签发有效 `ymt_`（带 `role`/`seat_slug`） |
| **步骤** | 1) 持码执行 `ymbot team join --code ymt_… --role {slug}`（或等价 redeem→ensure→bind 主链） 2) 读当次响应 / 通道 stdout 3) `GET …/roster` +（若已得钥）`ymbot whoami` |
| **期望** | 成功（200/201）；响应含 **一把** 明文 `token=ymk_…`（仅此一次）+ `member_id`；目标空席 → `active` 且 holder=新成员；memory 空可接受；**不**要求 Admin 预先建 Bot/填 Key |
| **PASS 准则** | ① 响应/stdout 含恰好 1 个可解析 `ymk_[A-Za-z0-9]{8,}` ② roster 该 slug `status=active` 且 `bound_member_id` 非空 ③ whoami（用当次钥）`actor_kind=bot` 且 member 与绑席一致 ④ Admin 入场包路径仍无该钥 |
| **FAIL 信号** | 要先有 ymk_ 才能 redeem；成功但无 token；造人未绑席；绑到非目标席；返回 N×token；挤 active |
| **关联证/债** | 目标：HUMAN_ONBOARD §3.2 · BATCH_CREW B1 改向 `1f96809`；现网对照：J2 `ymbot-j2-evidence-20261024.txt`（**仍须预置 ymk_**，≠自动身份）· ONE_TAP E5 `ONE_TAP_E5-qa-retest-20261024.txt`（install→whoami→join，预置钥） |
| **状态** | **PASS**（live redeem 无 Bearer → 一把 ymk_；证 `HUMAN_ONBOARD-Q1-Q3-Q12-qa-retest-20261024.txt` @c4ae0a5+662346d） |

## Q2 续队不重发钥

| 项 | 内容 |
| --- | --- |
| **前置** | Q1 成功后同 Bot 持同一 `ymk_`；已写至少一条 memory（如 `last_focus`）或有 open Issue 指派 |
| **步骤** | 1) 同钥再 `team join --code ymt_… [--role]` 同队 2) 比对两次响应是否含新 `token` 3) `handoff get` / memory list |
| **期望** | **无**新 `token` / `token_issued≠true`；仍在原席或合法席；memory/todos/指派 **不丢**；join 幂等（可 skipped_enter） |
| **PASS 准则** | ① 第二次响应体/stdout **不**出现新明文 ymk_ ② member_id 与第一次相同 ③ handoff 含 Q1 后写入的 focus/Issue 标记 ④ roster 不出现双绑同 slug |
| **FAIL 信号** | 每次 join 刷一把新钥；清 memory；换 member_id；把别人从席上挤走 |
| **关联证/债** | J2 幂等 rejoin（`BOT_TEAM_JOIN_KEY_P0` J5）· T2 `T2-qa-retest-20261024.txt`（同 member 换钥 migrate，非「重发」）· BATCH_CREW 幂等条款 |
| **状态** | **PASS**（同 ymk_ 再 redeem 无新 token + memory Intact；同证 Q1 文件） |

## Q3 活跃禁顶替

| 项 | 内容 |
| --- | --- |
| **前置** | 目标席 `status=active`，holder=BotA；BotB（另一身份，可无钥 bootstrap 或另一 ymk_）持有效 `ymt_` 且指定该 slug/seat |
| **步骤** | 1) BotB `team join --code … --role {occupied_slug}` 或 `seat bind`/`enter`（含尝试 `--force`） 2) 读 HTTP/CLI 3) 再 GET roster |
| **期望** | **409** `conflict` + 中文人话（占用中 / 禁顶替）；原演员不变；FE/CLI **无**「强行顶替」主按钮 |
| **PASS 准则** | ① HTTP 409（或 CLI 非 0 且文案含占用/冲突）② roster 该席 `bound_*` 仍为 BotA ③ 带 force 对 active 仍拒（force 仅 stopped\|left）④ FE 409 专面无顶替 CTA |
| **FAIL 信号** | 200 且 holder 变 BotB；force 顶 active 成功且无审计；UI 提供确认顶替 |
| **关联证/债** | JOIN_KEY J3 · API.md seat.force_bind 收窄 · ONE_TAP E5 soft†「席位占用 409」· `T3-conflict.png` · WHEEL/T2 leave→empty 后方可进 |
| **状态** | **PASS**（live：无身份撞 active→409；Bot force→409；holder 不变；同证 Q1 文件。Admin 代绑 403 仍拒） |

## Q4 新人新席

| 项 | 内容 |
| --- | --- |
| **前置** | 班底 **无** empty；invite `allow_expand_seat=true`（或等价扩编许可）；他席已有绑定 |
| **步骤** | 1) 新人 join（指定 role 或默认）2) GET roster 前后 diff |
| **期望** | 创建新席（新 slug 或兼岗 slug）并入座；**他席绑定不变**；不允许 expand 时人话失败（非 200 装成功） |
| **PASS 准则** | ① roster len 增加或出现新 slug 行 ② 新席 active=新人 ③ 预先 active 席 holder 字节级不变 ④ `allow_expand_seat=false` → 非 200 + 中文说明 |
| **FAIL 信号** | 挤掉 active；静默失败；expand 改坏 playbook_ref 致 400 却当成功；无许可仍开席 |
| **关联证/债** | JOIN_KEY J3/J5 expand · DISTRIBUTED_ONBOARD G2 · E5 soft†「expand 需 playbook_ref」观察项 · BATCH_CREW 任用「无空席→expand」 |
| **状态** | **GAP**（有 expand/禁顶替局部证；缺「满员+allow_expand→新席且他席不变」独立苛刻箱证） |

## Q5 补位

| 项 | 内容 |
| --- | --- |
| **前置** | 目标席 `left` 或 `stopped`（Admin 标停工/本人 leave）；新人持码或 ymk_；可选关联开放 Issue |
| **步骤** | 1) 新人 join/bind（通道可 `--force` **仅此态**）2) GET roster + activities/comments |
| **期望** | 入座成功；写人对人可见审计（`seat.force_bind` activity 人话 + 关联开放 Issue comment）；**对 active 仍禁 force** |
| **PASS 准则** | ① 席 active 且 holder=新人 ② 存在 force/补位审计记录（action 或 comment 可检索）③ 同步对另一 active 席 force → 仍 409 |
| **FAIL 信号** | 无审计静默换人；把 left 当 empty 无痕；active 可 force |
| **关联证/债** | JOIN_KEY J4 · N3 `3c698b5` force 审计 · API.md `activities.action=seat.force_bind` · seat-t4/t5 shots |
| **状态** | **COVERED†**（停工补位+审计：J4/N3；与「自动身份补位」组合待 Q1 后补） |

## Q6 Admin 入场包

| 项 | 内容 |
| --- | --- |
| **前置** | 从模板建队成功（默认班底 N=5）或席页存在 N 空席；Admin JWT |
| **步骤** | 1) 建队成功挡层或席页「复制全班入场包」2) 取剪贴板/挡层 markdown（或同源 `issueCrewPack` 产物）3) 正则抽检 |
| **期望** | ≥模板席数章节；每席含 `ymt_` + `role=` 链接/提示词（含安装段）；**正则无嵌 ymk_ 密钥**（说明性「勿贴 ymk_」文案可有） |
| **PASS 准则** | ① `ymt_` 出现次数 = 空席数 N（默认 5）② 五角色 slug 均出现（frontend/pm/qa/backend/wiki）③ `ymk_[A-Za-z0-9]{8,}` 命中 = **0** ④ CTA 存在于建队后与席页空席态 |
| **FAIL 信号** | 包内嵌真实 ymk_；缺席；逐席人手点才齐；只 Toast「N 空席」无包 |
| **关联证/债** | AUTO_CREW C7 PASS† `37b2228` · `AUTO_CREW-qa-retest-20261024.txt` · `auto-crew-pack-after-create.png` · FE `fef5914` · SEAT_VISIBILITY `3df4345`（空席可见才能发包） |
| **状态** | **COVERED†** |

## Q7 否决全班吐钥

| 项 | 内容 |
| --- | --- |
| **前置** | 建队流 + 席页 + 设置 Bot；对照 BATCH_CREW 改向后代码/路由 |
| **步骤** | 1) 巡 FE：建队成功 CTA、席页主按钮、Modal 清单 2) 巡 API：无 `POST …/crew/batch` 回 N×token；Admin 路径无「一键复制全班身份钥匙」 |
| **期望** | **无**全班 `ymk_` 包入口；无 batch 回 N×token；设置「新建 API Key」仍可 **单把** FOOLPROOF 露钥（旁路，≠建队主路径） |
| **PASS 准则** | ① UI 文案/按钮名不含「全班身份钥匙/一键造班吐钥」主 CTA ② 对可疑 batch 路由 → 404/未挂或非 N×token ③ 建队主 CTA = 入场包（Q6）④ 单把 `POST …/api-keys` 仍 201+token（旁路保留） |
| **FAIL 信号** | BatchCrewRevealModal 全班钥；crew/batch 201 含 tokens[]；建队优先吐 ymk_ |
| **关联证/债** | HUMAN_ONBOARD §1/§3.1 · BATCH_CREW `1f96809`/`037a6ff`（否决）· 废口径 `decabb0` · FOOLPROOF F6 `a5299f5`（单把旁路 PASS†） |
| **状态** | **GAP**（产品否决已落文档；需一次 UI/API 巡检箱证锁定「无入口」；建议文件名 `HUMAN_ONBOARD-Q7-no-batch-ymk-*.txt`） |

## Q8 分布式派发

| 项 | 内容 |
| --- | --- |
| **前置** | 席已绑 Bot + Webhook/巡检订阅；另有 left 旧演员成员；Admin 可建 Issue |
| **步骤** | 1) 指派 Issue `assignee_id`=该席 bound member（或席条「下发」）2) 观察 webhook 扇出 / 立即推送 / 巡检 3) 对 left 席试指派或误叫 |
| **期望** | **仅**该席 active member 被唤醒；left/旧 UUID **不**喊；入场与下发分离（派发不代替 join） |
| **PASS 准则** | ① 指派后目标 Bot 侧可见 Issue（list --assignee me / handoff open_issues）② webhook payload 含正确 assignee/seat 过滤 ③ left 成员未收到开工唤醒（无误叫日志）④ N2 下发 FE 预填 bound_member_id |
| **FAIL 信号** | 喊已 leave 旧 Bot；裸 team UUID 当主输入叫醒错误席；全员广播 |
| **关联证/债** | JOIN_KEY J7 · N2 `N2-dispatch-seat.png` · N4 `N4-dispatch-push.png` · `BOT_PATROL_P05` · DISTRIBUTED_ONBOARD |
| **状态** | **COVERED†**（下发/推送/巡检有证；与「自动身份后新演员」联动可再抽一刀） |

## Q9 换机续命

| 项 | 内容 |
| --- | --- |
| **前置** | 同 member 有效 `ymk_`；旧机已 `memory set` focus + 至少一开放 Issue；新机/新 HOME 仅有钥（无本地 handoff 文件） |
| **步骤** | 1) 新环境配 `YMBOT_API_KEY=ymk_…` 2) `ymbot whoami` → `handoff get`（及可选 `--full`）3) **不**找 Admin 再发钥、**不**强制 handoff import |
| **期望** | whoami=原 Bot；handoff 含 focus/todos/开放 Issue；席与任务池不变；换机≠重新走全班入场包主路径 |
| **PASS 准则** | ① whoami member_id 与旧机相同 ② handoff 含预设 focus 标记 ③ open_issues 含原指派 ④ 全程无 Admin mint 新钥、无 import 必经 |
| **FAIL 信号** | 必须 Admin 再吐钥才能续；handoff 空且记忆在旧文件；换机抢别人席 |
| **关联证/债** | BOT.md 续命 · T2 `T2-qa-retest-20261024.txt` PASS · WHEEL_FIX C3 `WHEEL_FIX-C3-seat-enter-20261024.txt` · seat-t5-migrate-guide · FOOLPROOF 旁路仅当丢钥 |
| **状态** | **COVERED†** |

## Q10 入场包 ≠ 身份

| 项 | 内容 |
| --- | --- |
| **前置** | 公开 `GET /join?code=ymt_…&role=` 页；席页/全班「复制 Bot 提示词」；对照设置 FOOLPROOF「复制启动提示词」 |
| **步骤** | 1) 抽 `/join` HTML + 可转发入场提示词 2) 抽 Admin 入场包 3) 抽设置单把启动提示词（私贴） |
| **期望** | URL / 可转发提示词 / 入场包：**无** ymk_ 明文；自动身份 **只**出现在入队方私密成功响应（或设置单把挡层）；文案两码分家 |
| **PASS 准则** | ① `/join` body 与 ONE_TAP 入场提示词：`ymk_[A-Za-z0-9]{8,}` = 0 ② AUTO_CREW 包同 Q6 ③ FOOLPROOF 启动提示词 **允许**嵌当次 ymk_（私贴）且含勿转发警告 ④ UI 不把入场码叫成「API Key」主名 |
| **FAIL 信号** | 公开 join/入场包嵌钥；创建成功无露钥却把钥写进可转发链接；ymk/ymt 混称导致贴错 |
| **关联证/债** | ONE_TAP A4 `ONE_TAP_A4-qa-retest-20261024.txt` · E4/E5 · FOOLPROOF F5/F6 `a5299f5` · `FOOLPROOF-F6-qa-retest-20261024.txt` · AUTO_CREW C7 · HUMAN_ONBOARD §1 |
| **状态** | **COVERED†**（公开禁钥 + 单把私贴；自动身份私密响应随 Q1 PASS；入场包 API 再抽 PASS† 同证） |

---


## Q11（V1）拒造第二人

| 项 | 内容 |
| --- | --- |
| **前置** | 同 Agent/同入场路径已通过 join 自动（或预置）获得身份 member M1；队内已有其席绑定 |
| **步骤** | 1) 再次无身份或「假装新人」走同一 ymt_ 入场/自动造人路径 2) 或对已绑定身份重复触发 auto-mint |
| **期望** | **拒绝造第二人**：不新建第二个 `is_bot` member；不第二把无关 ymk_；幂等回到 M1 / 人话已存在 |
| **PASS 准则** | ① member 计数不 +1（同 fingerprint/claim 键）② 若回 token 则仍属 M1 ③ 席绑定仍 M1 ④ 无静默双 Bot |
| **FAIL 信号** | 每次 join 新造 Bot；双 ymk_ 双 member；席上串人 |
| **关联证/债** | HUMAN_ONBOARD 可投表 V1 · 自动身份首刀 · Bot 记忆单活 claim |
| **状态** | **PASS**（同码无 Bearer 二次→409「请粘贴已发的身份钥匙续队」；同证） |

---

## Q12（F）吊销

| 项 | 内容 |
| --- | --- |
| **前置** | 有效 ymt_ 入场码；有效 ymk_ 身份钥；对应席可观测 |
| **步骤** | 1) Admin 吊销 ymt_ → 再 redeem/join 2) Admin 吊销 ymk_ → 再 whoami/业务 API |
| **期望** | ymt_ revoked 后入场失败（人话）；ymk_ revoked 后鉴权失败；吊销幂等；席/审计符合产品（不把 active 静默顶掉除非规则允许） |
| **PASS 准则** | ① revoke ymt_ → status=revoked；再 join 非 2xx/人话失败 ② revoke 幂等 200 ③ revoke ymk_ → whoami/API 401 或人话无效钥 ④ UI Popconfirm 吊销路径可用 |
| **FAIL 信号** | 吊销后码/钥仍可用；吊销半截；无错误提示 |
| **关联证/债** | N1 `ec1742d` · FE 吊销 Popconfirm · API.md api-keys revoke · 附录吊销行 |
| **状态** | **PASS**（ymt_ revoke→403 再兑；ymk_ `POST /api/v1/api-keys/{id}/revoke`→whoami 401；同证。注：ws 嵌套 revoke 路径 404） |

---

## Q13（K·J）救场双机

| 项 | 内容 |
| --- | --- |
| **前置** | Bot 已绑定；机 A 持 ymk_ 可 whoami；机 B 为新环境；可选：脏绑/卡住需救场 |
| **步骤** | 1) **J 多机**：机 B 仅配同一 ymk_ + handoff get/pull 2) **K 救场**：模拟脏绑/误占后走产品救场（rebind/清脏/重发 ymt_ 仅当规则允许） |
| **期望** | 换机/双机不强迫 Admin 再吐全班钥；handoff 权威续命；不非法双绑挤 active；救场后人话可恢复且留痕 |
| **PASS 准则** | ① 机 B whoami=同一 Bot ② handoff 得 focus/todos/开放 Issue ③ 无第二 member ④ 救场后可开干；禁「全班吐 ymk_」救场 |
| **FAIL 信号** | 换机必须 Admin 重发全班钥；双机双 Bot；救场静默顶 active |
| **关联证/债** | Q9 换机 · T2/WHEEL handoff · HUMAN_ONBOARD S5/K/J · 单活 claim |
| **状态** | **GAP**（Q9 局部 COVERED†；双机+救场联验待开干后抽） |

---

## 汇总

| ID | 场景锚 | 状态 | 主证据 / 债 |
| --- | --- | --- | --- |
| Q1 | S1 | **PASS** | live 无 Bearer redeem→ymk_；`HUMAN_ONBOARD-Q1-Q3-Q12-qa-retest-20261024.txt` |
| Q2 | S2 | **PASS** | 同钥再 redeem 无新 token + memory；同证 |
| Q3 | S8 | **PASS** | live 占用 409 + force 拒；同证 |
| Q4 | S7 | **GAP** | 缺满员 expand 独立箱证 |
| Q5 | S3 | **COVERED†** | J4 / N3 force 审计 |
| Q6 | S6 | **COVERED†** | AUTO_CREW C7 `37b2228` |
| Q7 | S6 | **GAP** | 否决已文档化；缺「无入口」巡检证 |
| Q8 | S4 | **COVERED†** | J7 / N2 / N4 / PATROL |
| Q9 | S5 | **COVERED†** | T2 / WHEEL C3 / BOT 续命 |
| Q10 | S1×横切 | **COVERED†** | ONE_TAP A4/E5 + F6 + C7；自动身份段随 Q1 |
| Q11 | V1 | **PASS** | 二次无 Bearer→409 人话；同证 |
| Q12 | F | **PASS** | ymt_+ymk_ 吊销 live；同证 |
| Q13 | K·J | **GAP** | 救场双机；Q9 局部可挂 |

**计数：** PASS **5**（Q1/Q2/Q3/Q11/Q12）· COVERED† **5**（Q5/Q6/Q8/Q9/Q10）· GAP **3**（Q4/Q7/Q13）· BLOCKED **0**  
（箱证 `docs/qa-shots/HUMAN_ONBOARD-Q1-Q3-Q12-qa-retest-20261024.txt`；Overall PASS†：FE Modal 点操/V6 签发侧 soft。）

## 附录 · 吊销 / 双钥（非 §6 主列，供债验收引用）

| 主题 | 期望摘要 | 证据 |
| --- | --- | --- |
| 吊销 ymt_ | Admin revoke → status=revoked；再 redeem/join 人话失败；幂等 200 | N1 `ec1742d` · `N1-revoke-list.png` |
| 吊销 ymk_ | revoke api-key 后 whoami/业务 API 失败（401/人话无效钥） | **PASS** 同证；路径=`POST /api/v1/api-keys/{id}/revoke` |
| 双钥心智 | ymk_=身份 Bearer；ymt_=入队；crew pack 只有 ymt_；F6 挡层才嵌 ymk_ | FOOLPROOF 表 · Q6/Q7/Q10 |

## 产品吸收（H3）

- 勾 [`NEXT_WAVE_DEBT_HUMAN_ONBOARD.md`](../NEXT_WAVE_DEBT_HUMAN_ONBOARD.md) **H3**：QA §6 补严文件 = **`docs/qa-shots/HUMAN_ONBOARD-QA-CASES.md`**。  
- 开干自动身份后：优先重跑 **Q1 → Q2 → Q10 → Q11**，并补 Q7 巡检、Q13 双机救场。  
- Q11–Q13 对齐票过可投表（V1 拒造第二人 / F 吊销 / K·J 救场双机）。  
- 勿把 `decabb0` 旧「全班 ymk_」当验收；以冻结 HUMAN_ONBOARD + `1f96809` 改向为准。
