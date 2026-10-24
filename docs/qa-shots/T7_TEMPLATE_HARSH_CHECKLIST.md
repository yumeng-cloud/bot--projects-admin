# T7-R 从模板创建 · 苛刻复测清单（QA）

> 权威：`docs/NEXT_WAVE_DEBT_T7.md`（**R1–R5**）· 对齐 `BOT_TEAM_TEMPLATE_T6_T7.md` T7-1/2/3/6/7  
> 开工锚 `32752f3` · **升版 T7-4/5 本包不做** · 等 Go SHA 后跑

## 证据文件名（对齐总指挥）
- **主（R1–R4）**：`docs/qa-shots/T7R-template-create-YYYYMMDD.txt`
- **独立复测（R1–R4）**：`docs/qa-shots/T7R-qa-retest-YYYYMMDD.txt`
- **R5 FE**：`docs/qa-shots/T7R-R5-qa-retest-YYYYMMDD.txt` + `t7r-r5-wizard-{open,preview,created}.png`
- 旧名 `T7-template-create-*` 不再使用

## 范围
| 做 | 不做 |
|----|------|
| 选模板 → 新 team + 空席（slug/名/playbook_ref，无 Key） | 升版 dry-run/apply |
| 非法/未知模板 4xx 中文 | 自动绑 Key |
| 同模板双队形状同；memory/pins 不串 | 漂亮多步向导 |

## 步骤（R1–R3 → 证据 R4）
1. 列可用模板；选一份 POST 创建 team
2. roster：席数/slug/显示名与手册一致；**无** bound Key；每席有 `playbook_ref`
3. 未知模板 / 非法 playbook_ref → 4xx 人话
4. 同模板再开队 B：形状同；A 的 memory/pins 不在 B
5. 写 `T7R-template-create-*.txt`；单测绿

## FE（R5）
最小预览表 → 创建。箱截 `bebe847` · 代码 `8cf4e59`。

| 项 | 标准 | 结果 | 证据 |
|----|------|------|------|
| **R5 open** | Bot 席位入口可见；弹窗中文完整；模板列表/默认班底可载；无空白破窗 | **PASS** | `t7r-r5-wizard-open.png` · `T7R-R5-qa-retest-20261024.txt` |
| **R5 preview** | 创建前预览 seats：slug/显示名/playbook_ref 与模板一致；空席无绑；不发明席 | **PASS** | `t7r-r5-wizard-preview.png` |
| **R5 created** | 成功 Toast 或落到新队；空席数/形状=模板；无 ghost bound；无崩/错 Toast | **PASS** | `t7r-r5-wizard-created.png`（队 `T7R-R5-1024` · 5 空席） |
| **R5 交叉** | 文案/按钮/API（GET templates；POST teams+template_key≡from-template）对得上 Modal | **PASS** | vs `CreateTeamFromTemplateModal.tsx` @ `8cf4e59` |

**R5 Overall = PASS** · 独立苛刻 · 2026-10-24 19:47 Asia/Shanghai

观察（不挡）：席位卡 WS bot_roles 平铺，新队五空席与历史混排；核靠 Toast+五空席行。

## 关报
短报产品+总指挥：Overall · R1–R4 · R5 · 证据路径 · HEAD
