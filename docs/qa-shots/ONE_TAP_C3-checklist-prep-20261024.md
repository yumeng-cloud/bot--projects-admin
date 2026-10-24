# ONE_TAP C3 · 打开即用 · QA 箱验清单（备）

> 2026-10-25 00:00 Asia/Shanghai · QA 备  
> **债：** [`docs/NEXT_WAVE_DEBT_ONE_TAP_JOIN.md`](../NEXT_WAVE_DEBT_ONE_TAP_JOIN.md) @`064c7fd` · 包 **C**  
> **门禁：** 等 React **C1**（登录态 join URL → 预填 JoinWizard，对齐 HUMAN V3）推 origin/main 后再跑活点。  
> **可挂：** HUMAN_ONBOARD Q1 联验证 `29743c3` / `HUMAN_ONBOARD-Q1-Q3-Q12-qa-retest-20261024.txt`（redeem 无 Bearer→ymk_）。  
> **产网：** `bot.api.sn.cn` 箱侧 DNS 不通 → **先箱内**；prod install.sh / E5 spotcheck 等运维 redeploy，不挡本条。

## 目标（C3 一句话）

**链接 → 打开 → join PASS**（箱证 txt 或截）。

## 前置

| 项 | 值 |
| --- | --- |
| API | 本地 `http://127.0.0.1:8080`，含 Go redeem ≥`c4ae0a5` |
| FE | 含 React C1（待交）+ V3/V4 ≥`662346d` |
| Admin | `admin@yumeng.tech` / seed |
| 码 | 签发 `ymt_` + `role`/`seat_slug`；URL = `{API}/join?code=ymt_…&role={slug}`（或 FE 同源 join） |
| 对照 | A4 PASS · `ONE_TAP_A4-qa-retest-20261024.txt`（`/join` 不 claim） |

## 步骤矩阵（React C1 到后执行）

| # | 步骤 | 期望 | PASS 准则 |
| --- | --- | --- | --- |
| **C3-1** | Admin 签发 invite → 复制入场链接 | URL 含 `code=ymt_` + `role=` | 与 A1 契约一致；无嵌 `ymk_` |
| **C3-2** | 未登录浏览器打开链接 | `/join` 短引导：可见码/角色/复制提示词/安装段；**不**消耗码 | HTTP 200；`use_count` 不变（A4 口径） |
| **C3-3** | **已登录** FE 打开同链接（**C1**） | 预填 JoinWizard（只贴/带 ymt_）；对齐 HUMAN V3 | 向导字段预填 code+role；无要求先有 ymk_ |
| **C3-4** | 向导/CLI/`POST …/team-invites/redeem` 无 Bearer 完成 join | 一把 `ymk_`；席 active；whoami=该 Bot | **挂 Q1**；可 IdentityKeyReveal（FE V4） |
| **C3-5** | 同钥再 join | 无新 token | 挂 Q2 |
| **C3-6** | 顺手：入场包/提示词（公开可转发） | **0×ymk_** | 挂 Q6/Q10；FOOLPROOF 成功页嵌钥除外 |

## 证据产出

写 `docs/qa-shots/ONE_TAP_C3-qa-retest-YYYYMMDD.txt`：time Asia/Shanghai、HEAD、C1 SHA、逐步 HTTP/截图路径、Overall PASS|PASS†|FAIL。  
绿：commit+push origin/main；短报 **产品体验 + 总指挥**。FAIL：钉 React（C1/向导）或 Go（redeem），用户睡中不问。

## 明确不做（本刀）

- 公网 `https://bot.api.sn.cn/install.sh`（E5 prod）  
- PRO_ADMIN 并行包  
- HUMAN 三 soft nits（已搁置）

## 状态

> 2026-10-25 ~00:01 Asia/Shanghai · QA 活点 **PASS†** · 证据 `ONE_TAP_C3-qa-retest-20261024.txt` · C1=`9779f23` · HEAD 见该文件。


| ID | 状态 |
| --- | --- |
| C3 checklist | **DONE** → `ONE_TAP_C3-qa-retest-20261024.txt` |
| C3 活点 | **PASS†**（C1@9779f23 含；见 retest） |
