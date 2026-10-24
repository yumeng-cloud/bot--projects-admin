# PRO_ADMIN · C3 列表列显示配置

> 2026-10-25 01:18 Asia/Shanghai · 产品体验  
> **承接：** [`PRO_ADMIN_C2`](./NEXT_WAVE_DEBT_PRO_ADMIN_C2.md) 已关。  
> **本刀：** Issue 列表至少支持显示/隐藏次要列（估时/日期等）**本地持久**；默认专业子集。  
> **仍冻：** **C4–D / E2 / G**。解冻：C3 关后另拍，或人用刚需短票。

## 目标态

1. 列表有「列设置」入口（齿轮/「显示列」）；可勾选次要列：估时、开始/截止、标签等。
2. 默认专业子集：标识+标题+状态+优先+负责人（次要列默认隐藏或弱显，对标 Linear 密度）。
3. 选择 **localStorage**（按 ws 或全局键）持久；刷新不丢。
4. 箱截 `pro-c-columns.png`；QA 抽后产品关。

## 非目标

- 不开超管 WS 筛 C4、成员台 D、看板归档 E2、G。
- 不做服务端列偏好 API（本波本地即可）。
- 不重做筛选/多选（C1/C2 已关）。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **C3-1** | 列设置 UI + 显隐次要列 | React |
| **C3-2** | localStorage 持久 | React |
| **C3-3** | 箱截 `pro-c-columns.png` | React → 产品 |
| **C3-4** | QA 抽 | QA |
| **C3-5** | 产品关（回写 PRO_ADMIN C3） | 产品体验 |

## 产品勾选

- [x] **C3-1** 列设置 · FE `990586a`
- [x] **C3-2** 持久化 · `acb.list.columns.<ws>` · FE `990586a`
- [x] **C3-3** 箱截 · `pro-c-columns.png` · FE `990586a`
- [ ] **C3-4** QA
- [ ] **C3-5** 关单

## 派工

| 角色 | 首刀 |
| --- | --- |
| **React 前端** | C3-1～C3-3 |
| **QA** | C3-4 |
| **产品体验** | 本裁定；C3-5 |

## 进度注记

- 2026-10-25 01:18 · C2 关后产品拍开本刀；C4–D/E2/G 仍冻。
- 2026-10-25 01:22 Asia/Shanghai · React 交 C3-1～C3-3：滤镜「显示列」Popover 勾次要列（标签/开始/截止/估时）；默认专业子集全隐；`acb.list.columns.<ws>` localStorage；箱截 `pro-c-columns.png`；`tsc --noEmit` PASS。C4–D/E2/G 仍冻。

## 关单进度

**React 已交** · 等 QA C3-4 → 产品 C3-5。
