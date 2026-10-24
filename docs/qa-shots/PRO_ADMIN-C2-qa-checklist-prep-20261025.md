# PRO_ADMIN C2 苛刻抽验清单（备）

> 2026-10-25 01:03 Asia/Shanghai · QA  
> 债：[`NEXT_WAVE_DEBT_PRO_ADMIN.md`](../NEXT_WAVE_DEBT_PRO_ADMIN.md) §C2  
> 产品拍：**仅解冻 C2**；**C3–D / E2 / G 仍冻**（勿验）。  
> 总指挥：有 FE SHA 即验；对齐会可能另开 HUMAN 软债/Toast — 随时接。

## 范围

| 验 | 不验 |
| --- | --- |
| **C2** Issue 列表多选 + 批量改状态/归档（解档可选）；操作条不挡表头 | C3 列配置、C4 超管 WS、D/E2/G |

## 箱截（React 应交）

- `docs/qa-shots/pro-c-bulk.png`（债 C5 中与 C2 对应；无则 FAIL 或记 soft† 等补）

## 苛刻点（任一红线 → FAIL）

### C2-1 多选
- [ ] 列表行可选（checkbox / 行选）；选中数可见
- [ ] 全选本页（若有）行为明确；不误伤筛选外行
- [ ] 中文主 UI；禁 English-only 主操作

### C2-2 批量改状态
- [ ] ≥2 条选中后可批量改状态；成功后列表态刷新
- [ ] Soft†：API 批量或逐条等价；失败人话 Toast，不静默半成功无提示
- [ ] 与单行改状态不打架

### C2-3 批量归档（解档可选）
- [ ] 批量归档后默认列表不可见（对齐 C1 默认剔归档）
- [ ] 若提供解档：仅在「显示/仅已归档」等可见归档上下文可用，或明确入口
- [ ] Soft†：`include_archived` / `archived=only` 下可见性正确

### C2-4 操作条不挡表头
- [ ] 多选后出现的批量操作条：**不遮挡**表头/筛选栏；不挤死主列表首行
- [ ] 对标有序密度：禁半屏死白操作条；禁叠在表头上
- [ ] 取消选择 / 关闭操作条后布局恢复

### 回归（抽）
- [ ] C1 三态归档滤镜仍可用（隐藏/显示/仅已归档）
- [ ] 搜索/状态/优先/负责人/标签筛选保持
- [ ] 禁 Form 栅格糊满列表工具区

## Soft† 建议

- API `:8080` · `admin@yumeng.tech` / `Admin123!`
- 多选 → 批量 status → 批量 archive → 列表默认不见 → include/only 可见
- 若缺批量 API：记缺口给 Issue 域；FE 若仅串行 PATCH 仍可 PASS†（注明 Soft†）

## 出件

- 报告：`docs/qa-shots/PRO_ADMIN-C2-qa-spotcheck-YYYYMMDD.txt`
- Overall PASS† / FAIL → 短报 **产品 + 总指挥 + React**（API 缺口另报 Issue 域）
- 测绿：`git push origin/main`（站立令，不必再问）

## 待命

- 等 React 交 C2 FE SHA + `pro-c-bulk.png`（或等价箱截）
- HUMAN 软债 / Toast 抽验：对齐会开则另接，不堵 C2


## VOID

2026-10-25 01:04 Asia/Shanghai · **作废勿验**。总指挥重申 C2–D 仍冻；产品撤回 C2 开刀。下一刀见 `docs/NEXT_WAVE_DEBT_SEAT_REFILL_SHOT.md`。
