# 下一小包 · T7-R · 从模板创建团队（最小）

> 2026-10-24 19:14 Asia/Shanghai · 产品体验 · T6 全文关后下一刀  
> **本包：选模板 → 空席 + playbook_ref。升版 diff（T7-4/5）后置。**

## 裁定

| 候选 | 结论 |
| --- | --- |
| **T7 最小（创建）** | **开本刀**（T7-1/2/3/6/7） |
| T7 升版 dry-run/apply | **后置**（T7-4/5） |
| 其它 UX 债 | 本波先收 T7-R |

对齐 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) T7-1～7；权威勾选仍可回写该表。

## 一句话

选一份 bot-team-template → 本 WS 新建 `team` + N **空席**（slug/显示名/`playbook_ref`，**未**绑 Key）；同模板开第二队形状同、记忆/pins **不串**（依赖已绿 T6）。

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| 模板清单可读（内置或配置）+ `POST` 创建挂 `playbook_ref` | 升版 dry-run diff / apply（T7-4/5） |
| `roster/ensure` 幂等补空席形状 | 自动绑 Key；镜像本机 Agent |
| 可选 FE「从模板创建」最小预览表 → 创建 | 漂亮多步向导大改 |
| 箱证：双队形状同、memory/pins 不串 | 模板写回用户 git；静默覆盖 memory |

## 验收（R1–R5）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **R1** | 选模板 → 新 team + 空席数/slug 与手册一致；无自动 Key | **Go** + 记忆 |
| **R2** | 每席含 `playbook_ref`；非法/未知模板 → 4xx 中文 | Go |
| **R3** | 同模板开队 B：形状同；私有 memory / pins 不串 A | 记忆/QA 箱验 |
| **R4** | 证据 `docs/qa-shots/T7R-template-create-YYYYMMDD.txt` | → **产品**勾 |
| **R5** | FE 最小预览+创建（可并行；无则 API 关半单） | **React** |

## 派工

**Go** 主 · **Bot 记忆** 箱验 · **种子** 手册索引 · **React** R5 并行 · **QA** 苛刻 · **产品**勾

## 产品勾选

- [ ] R1–R4 API/箱验 · 产品勾
- [ ] R5 FE（可后）
