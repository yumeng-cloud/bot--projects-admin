# P0 · 未登录官网 LANDING / SITE_LANDING

> 2026-10-25 Asia/Shanghai · 产品体验主持拍板  
> **承接：** 主人怒——总指挥自写落地页难看；总指挥停笔；产品定气质+文案，UI视觉·审图执刀（必要时 React）。  
> **视觉尺子（锁）：** [`SITE_LANDING_VISUAL_BASELINE.md`](./SITE_LANDING_VISUAL_BASELINE.md) @`170085f`——废稿黑名单与浅色 token **跟专册**，勿另皮。  
> **废稿参考（禁抄）：** `/workspace/yumeng-landing/index.html`（暗黑赛博）**可弃**。  
> **刀名互通：** **LANDING** ↔ **SITE_LANDING**（同一刀）。  
> **可与** [`INVITE_ON_DEMAND`](./NEXT_WAVE_DEBT_INVITE_ON_DEMAND.md) **并存**；本刀不改 INVITE/SEAT 业务逻辑。Soft 仍冻。  
> **不问用户。**

## 气质一句

**安静、好用、像真在干活的班**——克制留白、中文产品感；禁营销口号堆砌、禁 AI 免责腔、禁暗黑赛博大光晕堆特效。细节跟专册 §1 / §5。

## 刀名 / 主名

| 项 | 锁死 |
| --- | --- |
| **刀名** | **LANDING** / **SITE_LANDING**（互通） |
| **对外主名** | **育蒙协作看板** |
| **禁对外** | Linear-like / Grok Bot / 暗黑霓虹废稿腔 / 路线图自打脸 |

## 主人硬约束（锁死）

1. **路由：** 未登录默认页 = 官网；登录后才进后台。勿打乱现有鉴权：`/join` 深链保留；authed 的 `/` Home、`/w/:wsId/*`、`/admin` **不变**。  
2. **叙事：** 人与 Agent 像一支团队；「换机」= **分布式**——不同电脑认领不同角色干活，干完统一调度指挥（**禁止**写成普通「换机续会话 / 同步进度 / 换机续聊」话术）。  
3. **继续体现：** 席位班底、按需入场 K（有 K 用 / 无则生 / 替换才顶）、决策与审查用强模型 · 写码用轻量模型（档位配置 **本刀不实现**，只叙事铺垫）。

## 建议路由（L-R1 · 锁死意图）

现状（`App.tsx` @开刀前）：未登录 `path="*"` → `LoginPage`；`/join` 独立。

| 路径 | 未登录 | 已登录 |
| --- | --- | --- |
| `/` | **Landing**（官网） | Home（**不动**） |
| `/login` | 现 `LoginPage` | 可回跳原目标或 `/` |
| `/join` | 现 Join 深链（**不破**） | 现行为 |
| `/w/:wsId/*`、`/admin` | 仍可去登录（受保护） | **不变** |

登录成功：回跳原目标或 `/`。未登录访问受保护路径 → 登录（勿丢深链意图）。主 CTA「登录」进 `LoginPage`。

## 产品锁文案（勿再改意思 · 总监已认）

| 位 | 锁句 |
| --- | --- |
| **眉题** | 人与 Agent，一支团队 |
| **主标题** | 育蒙协作看板 |
| **副标** | 按席位组班；不同电脑认领角色，干完统一调度。 |
| **主 CTA** | 登录（进 `LoginPage`） |
| **价值① 标题** | 一支队 |
| **价值① 正文** | 人与 Agent 同场协作，共享看板与上下文。 |
| **价值② 标题** | 多机分工 |
| **价值② 正文** | 不同电脑认领不同角色干活，干完统一调度。（**禁**任何续会话 / 换机续聊措辞） |
| **价值③ 标题** | 席位 · 按需入场 · 模型档 |
| **价值③ 正文** | 班底按席占用；入场钥匙按需发；决策与审查用强模型，写码用轻量。 |
| **底栏** | 育蒙协作看板 |

### 眉题钉（产品补 · 禁 pill）

眉题「人与 Agent，一支团队」= **纯文一行**（字号 **13–14**、**secondary** 色）；**禁** accent 胶囊 pill（废稿黑名单 · 专册 §6）。

次 CTA 可有「了解入场」链简短说明或 `/login`，勿抢主 CTA。页脚勿路线图自打脸。

## 视觉（跟专册 · 勿另开尺子）

- Token / 废稿黑名单 / IA 上限 / 箱截路径：**整跟** [`SITE_LANDING_VISUAL_BASELINE.md`](./SITE_LANDING_VISUAL_BASELINE.md) @`170085f`。  
- 画布浅色 `--acb-bg-layout`；主色 `#5E6AD2`；禁 `#0b0d10` 暗底、radial 光晕、外挂 Google Fonts、eyebrow 胶囊。  
- 实现可落 `frontend/src/pages/Landing`（或等价）；样式跟现 FE token / 登录门 quiet 表面。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **L-P0** | 本债拍板（含锁文案 + 眉题禁 pill） | 产品体验 |
| **L-R1** | 路由：未登录 `/`=Landing；`/login`=Login；authed 后台不动；`/join` 不破 | React |
| **L-R2** | Landing 按上表锁文案落地（可组件在 `pages/Landing`）；眉题纯文 13–14 secondary，禁 pill | React / 审图 |
| **SL-V1** | 桌面箱截 `docs/qa-shots/site-landing-desktop.png`（1280×800 首屏含顶栏+Hero+至少一块价值） | 审图 → 总监 |
| **SL-V2** | 若超一屏：`site-landing-scroll.png`；可选 `site-landing-375.png` | 审图 → 总监 |
| **SL-V3** | 审图自审 playbook + 专册 → **视觉总监**复核放行/打回 | UI视觉·审图 → 视觉总监 |
| **L-Q1** | 未登录打开=/官网；登录后进后台；`/join` 仍可用；箱截；禁甩主人 | QA |
| **L-P1** | 产品关 | 产品体验 |

## 产品勾选

- [x] **L-P0** 本债拍板 · 2026-10-25（文案锁 + 眉题禁 pill + 链专册 `170085f`）
- [ ] **L-R1** 路由
- [ ] **L-R2** Landing 锁文案落地
- [ ] **SL-V1** 桌面首屏箱截
- [ ] **SL-V2** 全页 / 375（若需）
- [ ] **SL-V3** 视觉总监放行
- [ ] **L-Q1** QA
- [ ] **L-P1** 关单

## 派工

| 角色 | 首刀 |
| --- | --- |
| **产品体验** | 本裁定（气质+锁文案+眉题钉）；L-P0 / L-P1 |
| **React 前端** | L-R1、L-R2（路由 + Landing 页） |
| **UI视觉·审图** | 按专册出页；SL-V1/V2 箱截；自审后交总监 |
| **视觉总监** | 专册 @`170085f`；SL-V3 放行/打回 |
| **QA** | L-Q1（React+视觉可跑后；箱截；禁甩主人） |
| **Go / INVITE / SEAT / Soft** | **不动** |
| **总指挥** | 废稿停手；勿再自写营销 HTML |

## 非目标

- 不做定价 / 注册复杂漏斗。  
- 不做模型档位配置 UI（叙事铺垫即可）。  
- **不改** INVITE_ON_DEMAND / SEAT_* 业务逻辑。  
- Soft 三件（见 [`SOFT_RULING`](./NEXT_WAVE_DEBT_SOFT_RULING_20261025.md)）**仍冻**。  
- 不照搬 `/workspace/yumeng-landing` 暗黑赛博视觉（专册 §6 黑名单）。  
- 不重做已关 [`LOGIN_REFRESH`](./NEXT_WAVE_DEBT_LOGIN_REFRESH.md)（登录门仍在 `/login`）。

## 看板 Issue

- 标题：**LANDING 未登录官网**  
- 描述链相对路径 `docs/NEXT_WAVE_DEBT_LANDING.md` + 开刀 commit（push 后填 SHA）+ 视觉专册 `docs/SITE_LANDING_VISUAL_BASELINE.md` @`170085f`  
- Issue 链接：_待填_（缺口：公网狗粮 Bot 禁无父游离 Issue；与 INVITE 同缺口——仓内原无父票可挂。已有项目「产品债」待人类建父票后补链）

## 进度注记

- 2026-10-25 · **开刀拍板**。产品锁文案句（总监认）；眉题纯文禁 pill；链视觉专册 `170085f`；勾选对齐 L-R1/R2 + SL-V1/V2/V3 + L-Q1/L-P1；可与 INVITE 并存；Soft 仍冻。

## 关单进度

**开刀** · L-P0 已勾；等 React L-R1/R2 → 审图 SL-V1/V2 → 总监 SL-V3 → QA L-Q1 → L-P1。
