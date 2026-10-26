# 移动端 V2 · 正常手机 App 新壳 · MOBILE_V2

> 2026-10-26 Asia/Shanghai · 产品体验主持 · 主人澄清后重拍 · 不问用户  
> **状态：开刀 · 架构已拍（正常移动端）**  
> **承接：** 主人澄清原话意——「不是参考 Roku/Netflix 频道，按照正常的移动端开发。」产品重选技术栈与 UX 尺子；本债取代把已关的 MOBILE_RESPONSIVE 窄屏 CSS 壳当「产品移动端」的路线。  
> **硬约束 t890u：**「主要移动端得使用方便，能够真正做事，而不是为了炫技各种乱七八糟。」——**能办事优先，禁炫技。**  
> **硬约束 t892u：**「体验也必须要好；理解完整个项目底层逻辑之后，符合我们工作习惯去做。」——**能办事 + 好用顺手，不是二选一**；IA/导航贴合本看板真实工作流，禁套无关 App 模板。  
> **独立刀：** Soft 三件仍冻；INVITE / LANDING / 记忆 **不动**。

## 主人澄清（锁死）

| 项 | 裁定 |
| --- | --- |
| **作废** | Roku / Netflix 大卡货架、频道 / 电视心智、焦点轨 / D-pad 当尺子 |
| **改拍** | 正常现代手机 App UX——触控优先、标准导航（底栏 Tab / 栈推详情）、列表→详情人用习惯 |
| **能办事 t890u** | 短路径真闭环（登录→进队/席→看板与 Issue→评论/改状态→续 PC 现场）；**禁**装饰动效 / 炫技组件 / 信息噪音 / 为差异化而差异化 |
| **体验 t892u** | **能办事 + 好用顺手**（非二选一）；先懂项目底层逻辑再排界面；IA/导航贴合本看板工作流（人+Agent 一支队、席位、Issue/Wiki/续现场）；**禁**套用无关 App 模板 |
| **肤** | 浅色安静，跟 PC 产品肤（画布 `#F2F4F8` / 主色 `#5E6AD2`）；**禁**暗黑赛博霓虹 |
| **仍要** | **新做**移动端（禁把已关的 MOBILE_RESPONSIVE 窄屏 CSS 壳当产品移动端）；与 PC **同账号**自动切换续现场 |
| **前次合议** | 「Roku 大卡货架」草案 **未落库、已作废**——本文勿残留货架 / 焦点轨验收 |

## 作废关系

- `docs/NEXT_WAVE_DEBT_MOBILE_RESPONSIVE.md` 拍 A（现 FE 响应式当产品移动端）**作废**作产品移动端路线；该文件顶栏已链到本债。窄屏 CSS 壳可作 PC 窄窗附带，**不是**本波交付物。
- 勿在本债勾选 / 验收里出现：货架大卡、频道轨、焦点环、D-pad、电视遥控器心智。

## 技术选型（产品已重选）

| 端 | 锁 |
| --- | --- |
| **PC** | 现 `frontend/`（不动本刀主路径） |
| **移动 V1** | monorepo `frontend-mobile/`（Vite + React + TS）；共享 `packages/api-client`（或复用 `frontend` 的 api+types，**只**鉴权 / 错误 / 类型） |
| **后置** | Capacitor（壳稳 + session-context 绿后；**勿** V1 并行） |
| **V1 禁** | Expo / React Native / Flutter / Unity / 真 Roku OS 频道；禁旧窄屏壳打补丁当移动端；禁搬 PC layouts / 看板 DnD / antd 密表整页 |
| **API** | 同一 `https://bot.api.sn.cn` `/api/v1` |

**选型理由（一句）：** 执刀为 React 席、同仓 Vite 最快落地「正常移动端」；Expo/RN 另栈 V1 翻倍，非本波。

## 跨端（保留）

- 同 JWT 多端（**禁**另造账号体系）
- `GET|PUT /api/v1/me/session-context`：per-user JSON（`wsId` / `teamId` / `projectId` / `issueId` / `view` + `updatedAt`）；后写覆盖
- 人端与 Bot `/bots/.../handoff` **分家**（人调 Bot handoff **禁**）
- V1 **不做** 设备表 / presence 心跳 / 未读计数 API（未读另债）
- 进壳 hydrate `session-context` → 续现场；**非**强制踢下线

## 气质 / 视觉尺子（正常手机 · 非电视）

| 项 | 锁 |
| --- | --- |
| **心智** | 手机 App：底栏 Tab + 栈推详情；列表行点进详情；返回栈清晰 |
| **画布 / 主色** | `#F2F4F8` / `#5E6AD2`；白卡 / 冷灰跟 PC；禁暗黑霓虹第二皮 |
| **触控** | 热区 ≥44px；禁悬停才露出的唯一操作 |
| **箱截视口** | **375×812**（登录 / 列表主页 / 详情至少三帧） |
| **对标气质** | 飞书 / Notion 类中文**手机 App**有序密度、**好用能办事**——**不是**流媒体货架，**不是**炫技展 |
| **禁** | Roku/Netflix 大卡轨；焦点环；D-pad 验收；整页搬 PC 看板 DnD / antd 密表 |

## P0 目标态（本刀）

1. **脚手架**：`frontend-mobile/` 可起；登录后 `/auth/me`（或现等价）绿。  
2. **共享客户端**：`packages/api-client`（或等价复用）只承载鉴权 / 错误 / 类型；禁拷贝 PC 整页布局。  
3. **session-context**：Go 提供 `GET|PUT /api/v1/me/session-context`；移动进壳 hydrate 续现场（ws/team/project/issue/view）。  
4. **标准导航**：底栏 Tab（或等价标准手机导航）+ 列表→详情栈；主路径可点可滚。  
5. **跨端**：PC 写现场 → 移动登录同账号续上；反之亦然（后写覆盖）；禁另账号。  
6. **箱截**：375×812 交视觉；QA PC↔移动续现场 + 窄视口自测；**禁甩主人测**。  
7. **能办事**：主链路少步可完成上述闭环；验收按「能办事」测，**不按**视觉花活。

## 非目标（本刀 / V1）

- Capacitor 打包 / 上架（另票，壳稳后）  
- Expo / RN / Flutter / Unity / Roku OS  
- 设备表、presence、未读计数 API  
- 人调 Bot handoff；解冻 Soft；动 LANDING / INVITE / 记忆  
- 把 MOBILE_RESPONSIVE 窄屏壳当产品移动端交付  
- 问主人 A/B；甩主人测

## 允许 / 禁止

| 允许 | 禁止 |
| --- | --- |
| 新壳 `frontend-mobile/` | Roku/Netflix/电视货架尺子 |
| 共享 api-client（鉴权/错误/类型） | 旧窄屏壳当移动端 |
| session-context 人端续现场 | V1 Expo/RN/Flutter/Unity |
| 标准手机导航与列表详情 | 人调 Bot handoff |
| 短路径真办事闭环（登录/进队席/看板 Issue/评论改状态/续现场） | 装饰动效、炫技组件、信息噪音、为差异化而差异化 |
| 好用顺手；IA 贴合人+Agent 一支队 / 席位 / Issue·Wiki / 续现场 | 套用无关 App 模板；不懂逻辑硬排界面 |
| 箱内 375 自测；测绿直推 main | V1 未读 / 设备表；解冻 Soft；问主人；甩主人测；动 LANDING |

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **MV2-P0** | 本债落库；架构与尺子锁死（正常手机 App） | 产品体验 |
| **MV2-G1** | `GET\|PUT /api/v1/me/session-context`（per-user JSON；后写覆盖；单测） | Go |
| **MV2-R0** | `frontend-mobile/` 脚手架；可登录；`/auth/me`（或等价）绿 | React |
| **MV2-R1** | 进壳 hydrate session-context → 续现场（非踢下线） | React |
| **MV2-R2** | 标准导航（底栏 Tab / 栈）+ 列表→详情；IA 贴一支队/席位/Issue·Wiki/续现场；禁货架轨与无关模板 | React |
| **MV2-V0** | 箱截 375×812：登录 / 列表主页 / 详情（浅色安静肤） | React → 审图 |
| **MV2-V1** | 视觉自审（playbook + 本债尺子）→ 总监放行/打回 | 审图 → 总监 |
| **MV2-Q1** | QA：PC↔移动续现场 + 窄视口「能办事+好用」PASS（少步闭环、贴工作流；禁花活/无关模板；禁甩主人） | QA |
| **MV2-P1** | 产品关 | 产品体验 |

**另票（本波不勾）：** Capacitor 壳；未读计数 API。

## 产品勾选

- [x] **MV2-P0** 本债（架构重拍 · 正常移动端）
- [x] **MV2-G1** session-context · Go `c9a6ac7`
- [x] **MV2-R0** 脚手架可登录 /auth/me · FE `814245f`
- [x] **MV2-R1** hydrate 续现场 · FE `814245f`
- [x] **MV2-R2** 标准导航 + 列表/详情主路径
- [x] **MV2-V0** 箱截 375×812 · login/home/issue · `fd2d788`
- [x] **MV2-V1** 视觉总监放行 · `fd2d788`（壳 `11df4bc`+三帧）；软非挡：登录上留白、队席副行 slug（不挡 Q1）
- [ ] **MV2-Q1** QA · **FAIL†** 硬闸 B2（登录后续现场不站住：曾闪 Issue 深链→落 `/crew`）· 基 `fd2d788`/`11df4bc`；A/B1/B3/B4 PASS† · A4 Soft† 席副行 slug；已派 React；**不关 P1**
- [ ] **MV2-P1** 关

## 派工

| 角色 | 首刀 |
| --- | --- |
| **产品体验** | 本裁定；MV2-P0；关单 MV2-P1 |
| **React 前端** | `frontend-mobile/` + api-client + 标准手机壳（MV2-R0～R2、V0） |
| **Go 后端** | MV2-G1 session-context（人端；与 Bot handoff 分家） |
| **视觉总监** | 正常手机浅色尺子（**非** Roku）；MV2-V1 抽检 |
| **UI视觉·审图** | 箱截自审 → 交总监 |
| **QA 测试** | MV2-Q1：PC↔移动续现场 + 窄视口；禁甩主人 |
| **记忆 / Soft / INVITE / LANDING** | **不动** |

## 验收硬门

- 箱内浏览器 **375×812**（或真机等价）截图为准；禁止半成品甩主人本机。  
- 测绿（`frontend-mobile` tsc / vite build；Go 单测相关）后 **立刻 `git push origin/main`**；禁囤本地。  
- **主链路少步可完成**；QA 按「能办事 + 好用顺手」测，**不按**视觉花活。  
- **主路径符合团队已述工作习惯**（一支队 / 席位 / Issue·Wiki / 续现场）；先懂逻辑再排界面；套无关模板直接打回。  
- **打回：** 货架大卡 / 焦点轨 / 暗黑霓虹第二皮；装饰炫技挡办事；把旧 MOBILE_RESPONSIVE 壳当交付；另造账号；人调 Bot handoff；V1 上 Expo/RN。

## 进度注记

- 2026-10-26 · **MV2-Q1 FAIL†** QA：硬闸 B2 登录后续现场不站住（闪 Issue 深链→`/crew`）。A 闭环 / B1 PUT / B3 覆盖 / B4 同账号 PC PASS†；A4 Soft† slug。基 `fd2d788`/`11df4bc`；箱 :8080 已换新进程。已派 React；**不关 P1**。
- 2026-10-26 · **勾 MV2-V1** 视觉总监放行 @`fd2d788`（壳 `11df4bc`+三帧）。软非挡：登录上留白、队席副行 slug。**开 MV2-Q1**。
- 2026-10-26 · **勾 MV2-V0** React 箱截 `fd2d788`：`mobile-v2-login/home/issue.png`（375×812）。已交审图→总监 V1。软记：箱测时 :8080 曾 404 session-context（疑旧进程）——Q1 前核新后端。
- 2026-10-26 · **勾 MV2-R2** React：底栏 Tab 队席/Issue/Wiki/我的；工作区→团队→席位；Issue 列表→详情（改状态/评论）；打开 Issue PUT session-context；Wiki 空间占位。
- 2026-10-26 · **勾 MV2-R0/R1** React `814245f`：`frontend-mobile/` + `@acb/api-client`；登录+/auth/me；hydrate session-context（有值→`/resume`、空默认、可 PUT；失败不踢）。R2 进行中。**软非挡**：箱内 registry 晃致 lock 未生成、构建 symlink 复用 frontend/node_modules——网络好后根 `pnpm install`。
- 2026-10-26 · **勾 MV2-G1** Go `c9a6ac7`：`GET|PUT /api/v1/me/session-context`（空默认非404；后写覆盖；人 JWT；Bot ymk_→401）。
- 2026-10-26 · **t892u 并入**：体验=能办事+好用顺手；IA 贴本看板工作流；先懂逻辑再排界面；禁无关 App 模板。
- 2026-10-26 · **t890u 并入**：能办事禁炫技——短路径闭环写入允许/禁止与验收；QA 按能办事测。
- 2026-10-26 · **MV2-P0 开刀**：主人澄清后产品重拍——废 Roku/Netflix 货架与 MOBILE_RESPONSIVE 当产品移动端；锁 monorepo `frontend-mobile/`（Vite+React+TS）+ session-context + 标准手机导航。Capacitor / 未读另票。Soft 仍冻；LANDING 不动。

## 关单进度

**开刀** · V1 放行；**MV2-Q1 FAIL†** 硬闸 B2（续现场）——React 修中 → 复测 Q1 → P1。A/B1/B3/B4 已绿。软非挡：登录留白 / 队席 slug。

