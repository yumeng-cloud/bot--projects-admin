# 专业后台管理台（PRO_ADMIN）

> 2026-10-24 21:45 Asia/Shanghai · 产品体验  
> **用户定调（总指挥转）：** 整产品「凑合、不专业」→ 丰满成**专业后台管理台**（非鸡肋）。  
> **对标：** 飞书 / Notion 中文管理台 + Linear 密度；**有序密度**（禁挤叠，也禁大片死白）。见 `docs/CN_VISUAL_QA_PLAYBOOK.md`。  
> **承接：** `WHEEL_FIX` 已关（`cfed2d5`）；本波不重开已关能力，只在其上做「像专业后台」。

## 一句话

管理壳（侧栏 / 设置 / 超管台）+ 成员与列表能力 + Issue 工作台观感，达到「敢给人演示的专业后台」，不是功能堆砌也不是空白海。

## 现网基线（可扩，勿重做）

| 已有 | 本波姿态 |
| --- | --- |
| WS 自助开通 / 成员邀请 `ymw_` / 超管台 | **扩体验与密度**，不重开路由大包 |
| 用户搜索 `?q=` + 所属 WS 列（WHEEL D1/D2） | **扩筛选/批量/列可读** |
| Issue 归档/解档 + Relations 剔边 | **扩滤镜/批量与属性轨观感**；归档专面仍后置 |
| 看板/列表/详情 Linear 风已多轮 PASS | **补专业后台缺口**，禁另起欧美大留白皮肤 |

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| 设置 / 超管 / 空态 / 列表的有序密度与分区 | T7 升版 dry-run、未分队大入口 |
| 列表搜索/筛选/批量/列配置（Issue + 管理表） | 跨队搬家大 UI、物理删边 |
| 成员台与超管用户台「像飞书后台」 | 细粒度 RBAC 大重构、SSO |
| Issue 属性轨/归档滤镜专业感补强 | 归档专面整页、canceled 归档（仍后置） |
| 中文人话文案、禁 slug/UUID 露脸 | 重写 Wiki IA、工时计费模型 |

## 并行小包

### A · 信息架构（壳）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **A1** | 设置页：顶部分区清晰（工作区 / 成员 / 标签 / 接入），禁单页 Form 栅格挤满；左锚点或 Tabs 二级与侧栏语义一致 | 视觉基线 → React |
| **A2** | 超管台：工作区 / 用户两区视觉权重对等；顶栏标题+副文+主操作一排，禁漂浮卡片悬空半屏 | React + 视觉 |
| **A3** | 侧栏：通栏到底；当前项明确；设置/Bot/平台入口层级不打架（平台仅超管） | React |
| **A4** | 箱截 `pro-a-settings.png` + `pro-a-platform.png` | React → 产品 |

### B · 空态与引导

| ID | 标准 | 谁 |
| --- | --- | --- |
| **B1** | **P0** 新 WS / 无项目（`ProjectRedirect` 等）：主文案「暂无项目」+ **主按钮「创建项目」在本页打开创建表单/弹层**（`POST /projects`），成功直进该项目看板。**禁止**主 CTA「去设置」把人踢进设置迷宫；「去设置」最多作文字链次要。无 seed/技术口吻 | React |
| **B2** | 无成员：成员 Tab 空态引导邀请 / 签发 `ymw_`，与 Bot 席文案隔离 | React |
| **B3** | 列表/看板无 Issue：短引导+新建，不占半屏死白 | React |
| **B4** | 箱截 `pro-b-empty-*.png`（≥3：无项目/无成员/无 Issue） | React → 产品 |

### C · 列表能力（工作台感）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **C1** | Issue 列表：搜索+状态/优先/负责人/标签筛选保持；补「已归档」显隐开关（默认关，接既有 `include_archived`） | React |
| **C2** | Issue 列表：多选 + 批量改状态/归档（解档可选）；操作条不挡表头 | React；批量 API 缺口找 Issue 域 |
| **C3** | 列表列：至少支持显示/隐藏次要列（估时/日期等）本地持久；默认专业子集 | React |
| **C4** | 超管工作区表：名称搜索 + 状态筛选（启用/停用） | React（API 已有则只 FE） |
| **C5** | 箱截 `pro-c-list-filter.png` + `pro-c-bulk.png` + `pro-c-admin-ws.png` | React → 产品 |

### D · 权限与成员台（扩 WHEEL D）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **D1** | 设置·成员：表密度专业（角色芯片、操作折叠）；邀请区与表区分块，禁两坨 Form 糊一起 | React |
| **D2** | 超管用户台：搜索防抖保留；所属 WS 芯片可点跳（只读进 WS 或复制名）；空搜索提示 | React |
| **D3** | 超管用户台：按所属 WS 筛选（`workspace_id` 或 FE 滤；BE 有则优先） | Go 可选 + React |
| **D4** | 箱截 `pro-d-members.png` + `pro-d-users.png` | React → 产品 |

### E · Issue / 看板专业感

| ID | 标准 | 谁 |
| --- | --- | --- |
| **E1** | Drawer 属性轨：归档态/解档按钮与状态轨节奏统一（已有 B2，补间距与中文） | React + 视觉抽 |
| **E2** | 看板完成列「归档本列」入口与进度条不抢戏；文案人话 | React |
| **E3** | 深链已归档：banner+取消归档已有；补列表「仅已归档」视图或滤镜入口与 C1 对齐 | React |
| **E4** | 箱截 `pro-e-drawer-rail.png` + `pro-e-board-archive.png` | React → 产品 |

### F · 设置页密度与文案

| ID | 标准 | 谁 |
| --- | --- | --- |
| **F1** | 工作区改名/slug：短说明+字段，禁大段灰字；slug 仅管理员可见且有「用于 URL」一句人话 | React |
| **F2** | 标签 / API Key 等：表+顶操作，空态不露英文 key 名当标题 | React |
| **F3** | 全文扫：禁 `admin`/`member`/`is_bot` 原文；统一「管理员 / 成员 / 机器人」 | React |
| **F4** | 箱截 `pro-f-settings-density.png` | React → 产品 |

### G · 验收门（QA + 产品）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **G1** | QA：按包抽苛刻（挤叠/死白/露技术文案任一即 FAIL） | QA + UI 视觉 |
| **G2** | 证据：`docs/qa-shots/pro-*.png` + 可选 `PRO_ADMIN-smoke-YYYYMMDD.txt` | QA / React |
| **G3** | 产品目视全勾关单 | 产品体验 |

## 派工

| 角色 | 职责 |
| --- | --- |
| **视觉总监** | **首刀已落**：[`PRO_ADMIN_VISUAL_BASELINE.md`](./PRO_ADMIN_VISUAL_BASELINE.md)；教审图；抽 A/F/E |
| **UI 视觉·审图** | 按 playbook 自审箱截，交总监复核 |
| **React 前端** | A–F 主落地 + 箱截 |
| **Issue 域后端** | C2 批量若缺 API |
| **Go 后端** | D3 筛选若 FE 不够 |
| **QA** | G1–G2 |
| **产品体验** | 本裁定；布局验收；G3 |

## 产品勾选

- [x] **A1** 设置分区 · `709c6a6` / `pro-a-settings.png`（工作区·成员·标签·接入）
- [x] **A2** 超管台分区 · `709c6a6` / `pro-a-platform.png`（顶条+工作区/用户对等）
- [x] **A3** 侧栏层级 · 同 `pro-a-settings.png`（设置高亮·平台管理底入口）
- [x] **A4** 壳箱截 · `9f6ced2`；视觉终审放行（基线 `d0216ff`）
- [x] **B1** 无项目引导 · **P0 关绿** · FE `72e0447` + 底图 `640fd6e` · QA `631af4a` / `PRO_ADMIN-B1-qa-retest-20261024.txt` Overall PASS（B1-a～g）
- [x] **B2** 无成员引导 · `95974f1` / `pro-b-empty-no-members.png` · QA PASS `76824d1`
- [x] **B3** 无 Issue 引导 · `95974f1` · QA PASS `76824d1`
- [x] **B4** 空态箱截 · `95974f1`（无成员/无 Issue；无项目复用 B1 `pro-b-empty-no-project.png`）
- [x] **C1** 归档显隐 · FE `8c0fe5a` + QA `0e76217` / `pro-c-archived-toggle.png`
- [x] **C2** 多选批量 · FE `a33651d` + Issue `3b8a0d6` + QA `99795f9` · 正式关
- [x] **C3** 列显示配置 · FE `260fd1f` + QA `f65edc1` · 正式关
- [x] **C4** 超管 WS 筛选 · FE `516226d` + QA `5365db1` · 正式关
- [x] **C5** 列表箱截 · `pro-c-bulk`/`pro-c-columns`/`pro-c-admin-ws` · 随 C2–C4 齐 · 正式关
- [x] **D1** 成员表密度 · FE `f57c5c4` + QA `e0e55c7` · 正式关
- [x] **D2** 超管用户台 WS 芯片 · FE `38e0b4c` + QA `85b9acf` · 正式关
- [ ] **D3** 用户按 WS 筛
- [ ] **D4** 成员/用户箱截
- [x] **E1** 属性轨节奏 · FE `75fb14a` + QA `5fd8775` · 正式关 / `pro-e-drawer-rail.png`
- [x] **E2** 看板归档入口 · FE `11ac139` + QA `d2ff642` · 正式关
- [x] **E3** 仅归档视图入口 · FE `41c6e88` + QA `a8e0635` · 正式关 / `pro-e-archived-only.png`
- [ ] **E4** Issue 箱截
- [x] **F1** 工作区表单文案 · 短说明 +「用于 URL」· `pro-a-settings.png`
- [x] **F2** 标签/Key 表空态 · 成员分块邀请+表密度见 `pro-f-settings-density.png`（标签/接入同壳）
- [x] **F3** 角色中文人话 · 管理员/成员/机器人 · `pro-f-settings-density.png`
- [x] **F4** 设置密度箱截 · `9f6ced2` / `pro-f-settings-density.png`
- [ ] **G1** QA 苛刻
- [ ] **G2** 证据齐
- [ ] **G3** 产品关单

## 进度注记

- 2026-10-25 02:13 · D3 FE Soft† `7def3a2` 交验；产品勾 D3-1～D3-3，开 QA D3-4。**D4 / G 仍冻**。
- 2026-10-25 02:08 · **E2 正式关**后产品拍：**仅解冻 D3**（超管用户台按所属 WS 筛）；**D4 / G 仍冻**。见 [`PRO_ADMIN_D3`](./NEXT_WAVE_DEBT_PRO_ADMIN_D3.md)。
- 2026-10-25 · React 交 **D3-1～D3-3 Soft† FE**（用户台所属 WS 筛；API 无 workspace_id；箱截 `pro-d-users.png`）；等 QA/产品。**D4/G 仍冻**。
- 2026-10-25 02:06 · E2 FE `11ac139` 交验；产品勾 E2-1～E2-3，开 QA E2-4。**D3–D4 / G 仍冻**。
- 2026-10-25 01:58 · **D2 正式关**后产品拍：**仅解冻 E2**（看板完成列「归档本列」入口）；**D3–D4 / G 仍冻**。见 [`PRO_ADMIN_E2`](./NEXT_WAVE_DEBT_PRO_ADMIN_E2.md)。好用优先：日用看板归档 > 超管 D3 筛。
- 2026-10-25 01:55 · D2 FE `38e0b4c` 交验；产品勾 D2-1～D2-3，开 QA D2-4。**D3–D4 / E2 / G 仍冻**。
- 2026-10-25 01:55 Asia/Shanghai · React 交 **D2-1～D2-3**（用户台 WS 芯片可点跳/复制；空搜人话；`pro-d-users.png`）；等 QA/产品关。**D3–D4 / E2 / G 仍冻**。
- 2026-10-25 01:49 · **D1 正式关**后产品拍：**仅解冻 D2**（超管用户台：搜索防抖保留；所属 WS 芯片可点跳；空搜索提示）；**D3–D4 / E2 / G 仍冻**。见 [`PRO_ADMIN_D2`](./NEXT_WAVE_DEBT_PRO_ADMIN_D2.md)。
- 2026-10-25 01:47 Asia/Shanghai · React 交 **D1-1～D1-3**（成员芯片+⋯折叠+邀请/表分块；`pro-d-members.png`）FE `f57c5c4`；等 QA/产品关。**D2–D4 / E2 / G 仍冻**。
- 2026-10-25 01:36 · **C4+C5 正式关**（C 列表能力齐）后产品拍：**仅解冻 D1**（设置·成员表密度：角色芯片、操作折叠；邀请区与表分块）；**D2–D4 / E2 / G 仍冻**。见 [`PRO_ADMIN_D1`](./NEXT_WAVE_DEBT_PRO_ADMIN_D1.md)。
- 2026-10-25 01:34 · **React 交 C4-1+C4-2** FE `516226d`：超管工作区名称搜索+状态筛（FE 滤）；箱截 `pro-c-admin-ws.png`（=C5 本刀份；C1/C2/C3 截已有）。等 QA/产品勾；**D / E2 / G 仍冻**。
- 2026-10-25 01:24 · **C3 正式关**后产品拍：**仅解冻 C4**（超管工作区表：名称搜索 + 启用/停用筛选；C5 箱截并入本刀）；**D / E2 / G 仍冻**。见 [`PRO_ADMIN_C4`](./NEXT_WAVE_DEBT_PRO_ADMIN_C4.md)。
- 2026-10-25 01:18 · **C2 正式关**后产品拍：**仅解冻 C3**（列表列显示/隐藏次要列，本地持久；默认专业子集）为下一 P0；**C4–D / E2 / G 仍冻**。见 [`PRO_ADMIN_C3`](./NEXT_WAVE_DEBT_PRO_ADMIN_C3.md)。
- 2026-10-25 01:06 · 软债一次拍死 [`SOFT_RULING`](./NEXT_WAVE_DEBT_SOFT_RULING_20261025.md)：仅 SHOT 与 C2 并行开；V2 orphan/V6/revoke/focus/LEAVE soft/Q5 **搁置**；禁通码静默补。
- 2026-10-25 01:05 · **产品统一短票（总指挥反对 `8a722ae`）**：用户令不停/不堵/能干跑完/好用优先。**SHOT ∥ C2 并行**；截图不挡多选批量。**C3–D / E2 / G 仍冻**；解冻条件：C2 关后另拍 / 或人用刚需反对短票。已重派 React（C2+SHOT）+ Issue 域（C2-2 待命）。
- 2026-10-25 01:04 · （已被 01:05 统一短票取代）曾收回 C2 只开 SHOT；总指挥反对后改 **SHOT∥C2**。
- 2026-10-25 01:03 · （已撤回）曾拍解冻仅 C2；总指挥重申仍冻 C2–D 后收回。
- 2026-10-25 00:54 · SEAT_REFILL 关后下一刀 **SEAT_REFILL_BIND**（冻项外）。**C2–D / E2 / G 仍冻**（明示）。
- 2026-10-25 00:47 · **E3 正式关已落**；下一刀跳出 PRO_ADMIN → [`NEXT_WAVE_DEBT_SEAT_REFILL.md`](./NEXT_WAVE_DEBT_SEAT_REFILL.md)（停工席补位引导）。**C2–D / E2 / G 仍冻**（明示）。
- 2026-10-25 00:46 · **E3 正式关** FE `41c6e88` + QA PASS† `a8e0635`。三态归档滤镜齐。**C2–D / E2 / G 仍冻**（明示）。PRO_ADMIN 本波可验项已尽；下一刀听拍（另开或解冻）。
- 2026-10-25 00:45 · **产品勾 E3** FE `41c6e88`：三态「隐藏/显示/仅已归档」；only→`archived=only`；箱截 `pro-e-archived-only.png` 过。等 QA。C2–D / E2 / G 仍冻。
- 2026-10-25 00:45 · **E3 关绿** React `41c6e88`：列表归档滤镜升三态 Select（隐藏已归档 / 显示已归档 / 仅已归档）；only→`archived=only`（FE `archived_only`）；与 C1 include 互斥；深链 banner 不动；箱截 `pro-e-archived-only.png`；`tsc --noEmit` PASS。**C2–D / E2 / G 仍冻**（明示）。
- 2026-10-25 00:36 · **下一刀·产品拍 仅 E3**：列表「仅已归档」视图/滤镜入口与 C1「显示已归档」对齐（深链归档 banner 已有；本刀补列表入口）。**C2–D / E2 / G 仍冻**（明示）。React 首刀；齐后 QA 抽。
- 2026-10-25 00:36 · **E1 正式关** FE `75fb14a` + 视觉放行 + QA PASS† `5fd8775` / `PRO_ADMIN-E1-qa-spotcheck-20261025.txt`。C2–D / E2–E3 / G 仍冻。
- 2026-10-25 00:35 · **视觉抽 E1 放行** `pro-e-drawer-rail.png` / FE `75fb14a`：属性轨同节奏、主栏留白、非 Form。等 QA 齐后正式关。
- 2026-10-25 00:34 · **产品勾 E1** FE `75fb14a`：属性轨「归档」单行与状态/优先同节奏（已归档 + 取消归档）；主栏留白；箱截 `pro-e-drawer-rail.png` 过。等视觉抽 + QA。C2–D / E2–E3 / G 仍冻。
- 2026-10-25 00:35 · **E1 关绿** React `75fb14a`：Drawer 属性轨归档态/解档改为与状态/优先同节奏单行（13px·中文·label|value）；主栏居中留白保持；箱截 `pro-e-drawer-rail.png`；`tsc --noEmit` PASS。C2–D / E2–E3 / G 仍冻。
- 2026-10-25 00:30 · **SEAT_LEAVE 整包关**（Go `c5f5596` + FE `4883426` + QA `43ecff8` / `a54535e`）。产品内部拍：**下一 P0 = 仅 E1**（Drawer 属性轨：归档态/解档与状态轨节奏统一，补间距与中文；对标 Linear Property 轨）。C2–D / E2–E3 / G 仍冻。React 首刀；齐后视觉抽 + QA 抽。
- 2026-10-25 00:21 · **C1 正式关** FE `8c0fe5a` + QA PASS† `0e76217` / `PRO_ADMIN-C1-qa-spotcheck-20261025.txt`。产品+QA 双过。C2–E/G 仍冻。
- 2026-10-25 00:21 · **QA C1 抽验 PASS** `0e76217` / `PRO_ADMIN-C1-qa-spotcheck-20261025.txt`（默认关/开归档/其它筛/截全过）。**C1 关绿维持**。
- 2026-10-25 00:19 · **产品勾 C1** `8c0fe5a`：滤镜「显示已归档」默认关；开→`include_archived`+弱标签灰行；其它筛保持；箱截 `pro-c-archived-toggle.png` 过。C2–E/G 仍冻。**下一刀 SEAT_LEAVE 开刀**。
- 2026-10-25 00:18 · **C1 关绿** React `8c0fe5a`：列表滤镜「显示已归档」默认关；开→`include_archived=1`；归档行弱标签+灰态；其它筛保持；箱截 `pro-c-archived-toggle.png`；`tsc --noEmit` PASS。C2–E/G 仍冻；下一刀 SEAT_LEAVE。
- 2026-10-25 00:53 · **对齐总指挥**：本波认 **C1 归档显隐**（@`2e328dd`）先关；SEAT_LEAVE 排 C1 关后下一刀（债已落 `71e9b4a`，不丢）。C2–E/G 仍冻。
- 2026-10-25 00:50 · **C1（曾短暂让路 SEAT_LEAVE，已对齐回先关）**：总指挥拍下一 P0=SEAT_LEAVE；C–E/G 仍冻。见 [`NEXT_WAVE_DEBT_SEAT_LEAVE.md`](./NEXT_WAVE_DEBT_SEAT_LEAVE.md)。
- 2026-10-25 00:45 · **B2/B3/B4 小包正式关** FE `95974f1` + QA `76824d1`。
- 2026-10-25 00:45 · **下一 P0 = C1 仅**：「已归档」显隐开关（默认关，`include_archived`）；搜索/状态等既有筛选保持。不并行 C2–C5/D/E/G。React 首刀；齐后 QA 抽 + 产品勾。
- 2026-10-25 00:40 · **QA 抽验 PASS** `76824d1` / `PRO_ADMIN-B2-B3-qa-spotcheck-20261025.txt`（B2/B3 三条苛刻全过）。**B2/B3/B4 关绿维持**。
- 2026-10-25 00:35 · **产品勾 B2/B3/B4** `95974f1`：目视箱截过（成员空态邀请/`ymw_` 与 Bot 分家；看板/列表无 Issue 短引导+新建，非半屏死白）。等 QA 抽验。
- 2026-10-25 00:15 · **B2/B3/B4 关绿** React `95974f1`：成员 Tab 空态邀请/`ymw_`（与 Bot 席隔离）；列表/看板无 Issue 短引导+新建；箱截 `pro-b-empty-no-members.png` / `pro-b-empty-no-issue.png` / `pro-b-empty-no-issue-list.png`。整包 C–E/G 仍冻。
- 2026-10-25 00:25 · **开刀仅 B2/B3（+B4 箱截）**：ONE_TAP 已关；总指挥令只开人用空态。B2=成员 Tab 无成员→邀请/`ymw_`（与 Bot 席文案隔离）；B3=列表/看板无 Issue→短引导+新建（不半屏死白）。整包 C–E/G **不并行**。测绿 push；交勾带 SHA+截。
- 2026-10-25 00:10 · **产品**：HUMAN 关后不并行整包；建议后置只开 **B2/B3 空态**。本波席位给 ONE_TAP 收口。
- 2026-10-24 22:20 · **B1 QA 关绿** `631af4a` Overall PASS。

- 2026-10-24 22:17 · **B1 复勾**：有效底图 `640fd6e`；等 QA F 复验最终关绿。

- 2026-10-24 22:15 · **B1 暂挂**：QA Overall FAIL 仅因主证空白（`PRO_ADMIN-B1-qa-retest` @`f5ea674`）；功能 B1-a～g 绿。等重拍再勾。

- 2026-10-24 22:12 · 产品勾 **B1 P0** `72e0447`（代码+弹层截 PASS；空态底图空白催补拍）。

- 2026-10-24 22:05 · **P0 空态**：B1 加硬——无项目主按钮=本页创建项目，禁踢设置；催 React 热修 `ProjectRedirect`。

- 2026-10-24 22:00 · 产品勾 A1–A4 / F1–F4（视觉终审 `9f6ced2` · 基线 `d0216ff` · FE `709c6a6`）。下一刀 **B 空态**。

- 2026-10-24 21:45 · 债文定稿；WHEEL 已关，本波接上。首刀见派工「视觉总监」。

## 关单进度

**B/C1–C5/D1–D2/E2/SEAT_LEAVE/E1/E3 + SEAT_REFILL/BIND/SHOT 正式关** · **下一刀仅 D3**；**D4 / G 仍冻**。
