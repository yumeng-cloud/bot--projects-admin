# FULL_SMOKE_20261025 · 全页真人 E2E 验收清单（苛刻）

> **Wave:** `FULL_SMOKE_20261025`  
> **写清单时 HEAD:** `8d34c77`（`git rev-parse --short HEAD` · 2026-10-25 14:52 Asia/Shanghai）  
> **任务基线提及:** main @ `bc215de`（remote `bot--projects-admin`）  
> **授权:** 主人 t268u「全部都给我全面测试，每一页点开按照真人的全流程都测试一下」→ QA 全页真人流  
> **环境（开跑前已核）:**  
> - Vite `http://127.0.0.1:5175` → **200**  
> - API `http://127.0.0.1:8080/healthz` → **200** `{"status":"ok","service":"ai-collab-board",…}`  
> - Seed（狗粮，**禁主人账号**）: `admin@yumeng.tech` / `Admin123!`  
> **本文件状态:** Wave A 部分 + Wave B（项目 C1–C7）已跑 · **Overall 仍 PENDING** · **未宣称 FULL_SMOKE PASS** · A5 bare `/join` **FAIL** · LANDING 未关  

## 红线（违反即 FAIL / 禁写 PASS†）

1. **禁甩主人作 tester**；仅用 seed / 箱测账号。  
2. **PASS† 仅当有证据**（箱截 png + 纪要路径 + HEAD_at_run）；「能打开 / 不白屏」≠ PASS。  
3. 可编辑页必须走完：**打开 → 编辑 → 保存（成功 Toast / 列表即时刷新）→ 取消/关闭**；适用处覆盖空态与错误态（人话，禁技术栈）。  
4. 裸 hash / raw id / 技术字段当主文案 → **FAIL**。  
5. 大片漂浮空白、无产品空态、按钮无响应、白屏/未捕获错 → **FAIL**。  
6. Overall 全绿前禁止写 PASS / PASS†；进度记入 `FULL_SMOKE-progress-20261025.txt`。

## 判定符号

| 符号 | 含义 |
|------|------|
| **PENDING** | 未跑 / 无证据 |
| **PASS†** | 真人全流程绿 + 箱截/纪要证据齐 |
| **FAIL** | 任一红线；钉 React/Go + 截图 |

证据命名建议：`docs/qa-shots/FULL_SMOKE-*-20261025.png` · 纪要并入 progress 或另 `FULL_SMOKE-qa-YYYYMMDD.txt`。

---

## A · Landing / Auth / Join（未登录 + 登录门）

| # | 页/路径 | 真人步骤 | 期望（苛刻） | 证据 | 状态 |
|---|--------|----------|--------------|------|------|
| **A1** | `/` 未登录 Landing | 清 token / 无痕打开 `/`；滚完整页；点主 CTA「登录」 | 官网 Landing（眉题/主标「育蒙协作看板」/副标/价值区），**不是**sole 登录墙；主 CTA 进 `/login` | docs/qa-shots/FULL_SMOKE-20261025-landing-unauth.png | **PASS†** |
| **A2** | `/login` 登录 | 邮箱+密码提交 seed；错误密码试一次 | Toast「登录成功」→ 进后台 Home；错密人话失败，无白屏 | login-page + login-bad + home-authed（FULL_SMOKE-20261025-*） | **PASS†** |
| **A3** | `/login` 注册 Tab | 切「注册」；填姓名/邮箱/密码（≥8）→ 提交或取消 | 表单完整；成功入会话或人话错；取消不脏状态 | | **PENDING** |
| **A4** | `/join` 未登录 | 打开 `/join` 与 `/join?code=ymt_dummy`（或无效码） | 短引导可见（复制 Bot 提示词 / 登录后打开向导）；无效码人话；**不消耗码**；「回首页」可用 | | **PENDING** |
| **A5** | `/join` 已登录 | 登录后带有效/无效 `code` 开 `/join` | 跳席页并预填「加入班底」向导或明确警告；不破深链 | FULL_SMOKE-20261025-join-page.png · bare `/join` 已登录 Spin>20s | **FAIL** |
| **A6** | 受保护深链未登录 | 无痕打开 `/w/{wsId}/board` 或 `/admin` | 重定向 `/login` 且保留 from；登录后回跳或合理落 Home | | **PENDING** |

---

## B · Home + 工作区切换

| # | 页/路径 | 真人步骤 | 期望（苛刻） | 证据 | 状态 |
|---|--------|----------|--------------|------|------|
| **B1** | `/` 已登录 Home | 登录后打开 `/` | 「我的工作区」列表；每卡「看板/列表/周期/Wiki/设置」可点；超管见「平台管理」；「退出登录」可用 | FULL_SMOKE-20261025-home-authed.png · 8 WS | **PASS†** |
| **B2** | 创建工作区 | Home「创建」→ Modal 填名称+slug → 保存 | Toast 成功；列表出现新 WS；可进 `/w/{id}/board` | | **PENDING** |
| **B3** | 兑换邀请码 | Home「兑换」→ 填 `ymw_…` → 确认；试无效码/空提交 | 成功「已加入工作空间」；无效人话；队席码 `ymt_` 指引去 Bot 席 | | **PENDING** |
| **B4** | 进仓 + 顶栏切仓 | 从 Home 进 WS；顶栏/抽屉切换另一 WS | 切仓后侧栏名、项目池、Bot 席跟新仓；**零旧仓残留**（对照 WS_ISOLATION） | ws-entered + project-switch（FULL_SMOKE-20261025-*）· SQ1A 无 bleed | **PASS†** |
| **B5** | 空仓空态 | 新/无项目 WS 开 `/w/{ws}/board` | 空态「还没有项目」+ 主 CTA「创建项目」本页 Modal；「去设置」仅次要 | | **PENDING** |

---

## C · 项目视图 · 看板/列表/周期/排期 + Issue CRUD

> 侧栏中文：看板 · 列表 · 周期 · 排期 · Wiki · 工时 · 设置 · Bot  
> 项目路径：`/w/:wsId/projects/:projectId/{board,issues,cycles,timeline}`；无 projectId 的 `/board|/list|/cycles|/timeline` 经 ProjectRedirect。

| # | 页/路径 | 真人步骤 | 期望（苛刻） | 证据 | 状态 |
|---|--------|----------|--------------|------|------|
| **C1** | 看板 `…/board` | 打开有项目看板；拖一张卡（若支持）或点开 Issue | 列+卡片可见；非空壳；Issue 标识人话（如 YM-n）非裸 UUID | FULL_SMOKE-20261025-board.png | **PASS†** |
| **C2** | 列表 `…/issues` 或 `/list` | 打开列表；排序/筛选轻抽（若有） | 行数据或像样空态；无崩 | FULL_SMOKE-20261025-list.png | **PASS†** |
| **C3** | 周期 `…/cycles` | 打开周期；有则开详情；无则看空态 | 周期列表/详情区正常；禁漂浮空白 | FULL_SMOKE-20261025-cycles.png · minor 无「新建周期」 | **PASS†** |
| **C4** | 排期 `…/timeline` | 打开排期视图 | 时间线/排期渲染；非空白崩 | FULL_SMOKE-20261025-timeline.png | **PASS†** |
| **C5** | Issue 新建 | 看板或列表「新建」→ 填标题（±描述）→ **保存**；再 **取消** 一次半填关闭 | 保存后列表/看板可见；Toast/即时刷新；取消不落脏数据 | board-created · YM-33 FULL-SMOKE-B-20261025-145528 | **PASS†** |
| **C6** | Issue 打开·编辑·保存 | 开 Issue 抽屉/页；改标题/描述 → 保存；属性轨改状态/负责人（若可） | 保存成功；关闭再开值仍在；属性轨 label\|value 中文 | issue-detail · desc+状态「进行中」· minor 叶/橙斜 | **PASS†** |
| **C7** | Issue 评论 | 评论区发一条 → 见列表；空评论提交应拦 | 评论可见；失败人话；无未捕获错 | issue-detail · smoke comment OK | **PASS†** |
| **C8** | Issue 关联 | 关联区加/看关联；空态「暂无…」类文案 | 关联可建或空态像样；禁裸 id | | **PENDING** |
| **C9** | Issue 归档/取消归档 | 属性轨「归档」→ 确认；再「取消归档」（若有权） | 已归档只读态清晰；取消归档恢复可编 | | **PENDING** |
| **C10** | 附件（若入口可见） | 上传合法图/拒非法类型 | 预览/列表；拒类型 Toast 人话（对照 I3） | | **PENDING** |

---

## D · Wiki

| # | 页/路径 | 真人步骤 | 期望（苛刻） | 证据 | 状态 |
|---|--------|----------|--------------|------|------|
| **D1** | `/w/:wsId/wiki` | 打开 Wiki；浏览目录 | 目录+舞台；禁完成紫胶囊/裸 Select 行/漂浮空白 | | **PENDING** |
| **D2** | Wiki 开篇 | 点一篇已有页（或空态引导） | 正文可读；时间中文/上海时区友好 | | **PENDING** |
| **D3** | Wiki 编辑·保存·取消 | 进入编辑 → 改标题/正文 → **保存**；另开编辑 → **取消** | 保存后正文回显；取消丢未存改动；无白屏 | | **PENDING** |
| **D4** | Wiki 空态/权限 | 无页或只读角色路径 | 产品空态或只读提示；禁技术口吻 | | **PENDING** |

---

## E · 工时

| # | 页/路径 | 真人步骤 | 期望（苛刻） | 证据 | 状态 |
|---|--------|----------|--------------|------|------|
| **E1** | `/w/:wsId/time` | 打开工时全页 | 表/汇总可见；Issue 列 **YM-n + 标题**；成员显示名（如「管理员」）禁 raw id | | **PENDING** |
| **E2** | 工时编辑（若可） | 新增或改一条工时 → 保存；取消一次 | 保存刷新；取消不脏；错误人话 | | **PENDING** |
| **E3** | 工时空态 | 无记录筛选或新仓 | 像样空态，非大白块 | | **PENDING** |

---

## F · 设置 · 全 Tab 编辑·保存

> 路径 `/w/:wsId/settings` · Tab：`workspace` \| `projects` \| `teams` \| `members` \| `labels` \| `integrations`（UI「接入」）  
> 页眉：「设置」·「工作区 · 项目 · 团队 · 成员 · 标签 · 接入」

| # | Tab | 真人步骤 | 期望（苛刻） | 证据 | 状态 |
|---|-----|----------|--------------|------|------|
| **F1** | 工作区 `?tab=workspace` | Admin 改显示名（±slug）→ **保存**；非 Admin 看只读 | Toast「工作区已更新」；**顶栏/侧栏即时新名**；非 Admin「仅管理员可修改…」 | | **PENDING** |
| **F2** | 项目 `?tab=projects` | 「编辑」→ Modal 改名 → 保存；取消一次 | Toast+列表名变；Modal 关；取消不改 | | **PENDING** |
| **F3** | 团队 `?tab=teams` | 「编辑」改名→保存；点「设为当前」「成员」「Bot 席」 | 改名即时；三钮路径正确不回归 | | **PENDING** |
| **F4** | 成员 `?tab=members` | 邀请已注册成员；签发 `ymw_` 链接；成员「更多」改角色；空态（若可） | Toast「已邀请成员」/「角色已更新」；唯一管理员降级拦截人话；空态「还没有成员」+「邀请成员」；Bot 脚注与人类分家 | | **PENDING** |
| **F5** | 标签 `?tab=labels` | 新建标签；行「编辑」改名/色→保存；删或空态 | Toast「已新建标签」/编辑成功；空「还没有标签」 | | **PENDING** |
| **F6** | 接入 `?tab=integrations`（或「接入」） | 打开；若可创建 API Key → 看明文一次挡层；列表仅末四 | 与席码 `ymt_` 分家说明清晰；明文仅一次；无崩 | | **PENDING** |

---

## G · Bot 席 · 名册 / 邀请查找 / 详情

| # | 页/路径 | 真人步骤 | 期望（苛刻） | 证据 | 状态 |
|---|--------|----------|--------------|------|------|
| **G1** | `/settings/bots` 名册 | 打开 Bot 席；切团队（顶栏） | Roster 跟当前队；切队零串；空态「尚未绑定/请先创建或加入班底」产品文案 | | **PENDING** |
| **G2** | 席位邀请·查找 | 打开「加入班底」向导；粘贴/查找 `ymt_`；签发/吊销列表轻抽 | 向导可开可关；无效码人话；已签发列表可吊销（或接口未开通人话警告） | | **PENDING** |
| **G3** | 占席 / 停工 / 离职 | 有席时：占席或停工/离职解绑（勿毁唯一生产席；优先狗粮队） | Toast 人话；冲突面板可读；绑定失效有提示 | | **PENDING** |
| **G4** | Bot 详情 `/settings/bots/:memberId` | 从名册进详情；看记忆 / Todo / handoff Tab | 「Bot 记忆与交接」；席位 Tag；空「暂无记忆」；换仓无旧数据残留 | | **PENDING** |
| **G5** | 详情编辑 | 记忆保存；新建待办→改状态；handoff 导出/取消导入 | Toast「已保存」；Todo 状态中文；导入校验人话 | | **PENDING** |

---

## H · 平台管理（超管）

| # | 页/路径 | 真人步骤 | 期望（苛刻） | 证据 | 状态 |
|---|--------|----------|--------------|------|------|
| **H1** | `/admin` 入口 | seed 超管从 Home/侧栏进「平台管理」；非超管应拒 | 顶栏「平台管理」+「全站工作区与用户」；非超管「需要平台超级管理员权限」 | | **PENDING** |
| **H2** | 工作区台 | 名称搜索 + 状态筛；新建工作区 → 保存；「返回工作区」 | Toast「已新建工作区」；筛生效；表 hug 不崩 | | **PENDING** |
| **H3** | 用户台 | 搜索用户；点所属 WS 芯片；超管授予/撤消（狗粮号，**禁动主人**） | 进 WS 或「已复制工作区名称」；Toast「已授予/已撤消平台超管」 | | **PENDING** |
| **H4** | `/admin/*` 深链 | 打开已知子路径或刷新 | 不 404 白屏；回到对等台或合理态 | | **PENDING** |

---

## 开跑纪律

1. 每行自点后改状态 PENDING → PASS† / FAIL，并填证据路径。  
2. FAIL 钉同事（React/Go）+ 截图；**不问主人**。  
3. 进度追加 `docs/qa-shots/FULL_SMOKE-progress-20261025.txt`。  
4. 全组 A–H 全绿后方可 Overall **PASS†**；本清单骨架阶段 **Overall = PENDING**。  
5. 进度见 `FULL_SMOKE-progress-20261025.txt` run1；**禁 Overall PASS**；**禁关 LANDING**。

## Overall

| 项 | 值 |
|----|-----|
| Wave | FULL_SMOKE_20261025 |
| Checklist | **PENDING**（Wave A 部分 + C1–C7 PASS† · A5 FAIL · 非 Overall PASS） |
| Progress log | `docs/qa-shots/FULL_SMOKE-progress-20261025.txt` |
| HEAD_at_checklist_write | `8d34c77` |
| Env probe | Vite:5175 **200** · `/healthz` **200** · 2026-10-25 14:52 Asia/Shanghai |
