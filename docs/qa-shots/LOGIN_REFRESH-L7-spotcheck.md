# LOGIN_REFRESH L7 · QA spotcheck

| 项 | 值 |
| --- | --- |
| **time** | 2026-10-25 08:32 Asia/Shanghai |
| **result** | **PASS†**（harsh Soft†：源码 + 箱截对照，未强制起 Vite） |
| **debt** | `docs/NEXT_WAVE_DEBT_LOGIN_REFRESH.md` |
| **FE commit** | `005a3c18567eb2c6e664e7c47cb26a30d84b8085` · `feat(fe): LOGIN_REFRESH L1–L5 登录门焕新` |
| **HEAD at run** | `0881a3aff8ecfaa0e7ad4e6e70758149b7faa623` · `docs(ux): LOGIN_REFRESH L6 视觉放行；开 QA L7` |
| **ancestor** | `005a3c1` ∈ ancestry of `0881a3a` · YES |
| **截图复用** | YES · `docs/qa-shots/login-refresh.png` + `login-refresh-register.png`（同 FE `005a3c1`，未重截） |
| **live Vite** | 未起（源码与箱截一致且 L6 已放行；Soft† 可接受） |
| **禁区** | 未改业务功能代码；未碰 UNBIND_LOOP / Soft† 静默补 |

## Method

1. `git cat-file -t 005a3c1` + `merge-base --is-ancestor` → FE 在 main 祖先链
2. `git show 005a3c1` → 仅 `LoginPage.tsx` / `login.css` / `index.html` + 两帧箱截 + debt 注记
3. Read `frontend/src/pages/LoginPage.tsx` + `login.css` + `frontend/index.html`
4. Read(PNG) 两帧箱截对照主名/副标/Tabs
5. `rg` 禁文案：`Linear` / `Grok Bot` / `laner` / 主名「AI 协作看板」→ Login 门零命中

## Checklist

- [x] **可进登录页** — `LoginPage` 路由仍挂 `App.tsx` `path="*"`；Tabs「登录」表单完整（邮箱+密码+提交）；箱截 `login-refresh.png` 无白屏、无错栈、无破碎布局 → Soft† PASS
- [x] **可进注册** — Tabs「注册」：姓名(可选)+邮箱+密码(≥8)+提交；箱截 `login-refresh-register.png` 同气质、无白屏/错栈 → Soft† PASS
- [x] **主名可见「育蒙协作看板」** — 源码 `LoginPage.tsx:40` + `index.html` `<title>`；箱截两帧 H1 逐字一致 → PASS†
- [x] **副标可见「人与 Bot 同场协作」** — 源码 `LoginPage.tsx:43`；箱截两帧副标逐字一致 → PASS†
- [x] **禁文案已砍** — Login 门无 Linear-like / Grok Bot / laner；主名非「AI 协作看板」→ PASS
- [x] **气质 Soft†** — 冷灰 `--acb-bg-layout` + 白卡宽 `400`（`login.css`）；箱截居中白卡、有序密度；L6 总监已放行，本刀不重审视觉 → 接受

## Evidence

| 路径 | 说明 |
| --- | --- |
| `docs/qa-shots/login-refresh.png` | 20390 B · mtime 2026-10-25 08:29 Asia/Shanghai · blob in `005a3c1` · 登录 Tab |
| `docs/qa-shots/login-refresh-register.png` | 20510 B · mtime 2026-10-25 08:29 Asia/Shanghai · blob in `005a3c1` · 注册 Tab |
| `frontend/src/pages/LoginPage.tsx` | 主名/副标/Tabs/表单 |
| `frontend/src/pages/login.css` | 冷灰画布 + 400 白卡 |
| `frontend/index.html` | `<title>育蒙协作看板</title>` |

## FAIL redlines hit?

NONE

- 白屏 / 错栈：NO（箱截完整卡面）
- 主名/副标缺失或错字：NO
- Linear / Grok Bot / laner 回潮：NO
- 本刀改业务功能 / UNBIND：NO（仅本纪要 + debt L7 勾）

## SUMMARY

| # | 项 | 结果 |
| --- | --- | --- |
| 1 | 登录可进 · 无白屏/错栈 | PASS† Soft† |
| 2 | 注册可进 · 无白屏/错栈 | PASS† Soft† |
| 3 | 主名「育蒙协作看板」 | PASS† |
| 4 | 副标「人与 Bot 同场协作」 | PASS† |

**Overall = PASS†**

给产品：L7 绿，请关 **L8**。
