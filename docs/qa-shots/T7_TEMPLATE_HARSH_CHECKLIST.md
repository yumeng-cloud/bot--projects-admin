# T7 从模板创建 · 苛刻复测清单（QA）

> 权威：`docs/NEXT_WAVE_DEBT_T7.md`（R1–R5）· 对齐 `BOT_TEAM_TEMPLATE_T6_T7.md` T7-1/2/3/6/7  
> 开工锚 `32752f3` · **升版 T7-4/5 本包不做** · 等实现 SHA 后跑

## 证据文件名
- 主：`docs/qa-shots/T7-template-create-YYYYMMDD.txt`（总指挥口径）
- 别名可认：`T7R-template-create-*.txt`（债务表 R4）

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
5. 写证据；单测绿

## FE（R5，可后）
最小预览表 → 创建；不挡 API 半单关

## 关报
短报产品+总指挥：Overall · R1–R5 · 证据路径 · HEAD
