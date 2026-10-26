# P0 · Issue 描述/评论 Markdown 渲染

> 2026-10-26 Asia/Shanghai · 产品体验 · 总指挥报障后拍板  
> **刀名：** ISSUE_MD_RENDER  
> Soft/INVITE/LANDING/MOBILE_V2 **勿动**。不问用户。

## 一句话

Issue 详情里描述与评论须 **默认预览渲染** Markdown；禁生 `##` / `-` 直吐。

## 根因（已认）

| 位 | 现况 | 修法 |
| --- | --- | --- |
| 描述 | ByteMD `Editor` `mode=tab` 默认落编辑态 | 默认预览；对齐 Wiki Viewer（可复用 `WikiBodyViewer` 或 ByteMD 预览态） |
| 评论 | `Typography.Paragraph` 直吐 `body` | 同 Wiki 预览组件渲染 MD |

## 验收（短 QA）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **MD-R1** | 打开有 MD 描述的 Issue：默认见渲染后标题/列表，**非**生 `##`/`-` 编辑框 | React → QA |
| **MD-R2** | 评论列表 body 渲染 MD（列表/标题可读） | React → QA |
| **MD-R3** | 描述可进编辑；保存/失焦后回预览 | React → QA |
| **MD-Q1** | 箱截：描述预览 + 评论渲染各 ≥1；禁甩主人 | QA |
| **MD-P1** | 产品关 | 产品体验 |

## 非目标

- 不换编辑器品牌；不重做 Issue 双栏骨架  
- 不动移动端 `frontend-mobile`（本刀 PC `frontend/`）  
- 不解冻 Soft；不改 LANDING/INVITE

## 产品勾选

- [x] **MD-R1** 描述默认预览
- [x] **MD-R2** 评论渲染
- [x] **MD-R3** 编辑↔预览
- [ ] **MD-Q1** QA 短测 · **已开**（基 `9d1feba`）
- [ ] **MD-P1** 关

**状态：已开 MD-Q1** · 基 `9d1feba` → 待 PASS† → MD-P1。

## 进度注记

- 2026-10-26 · 总指挥 P0：生 `##`/`-`；已钉 React 对齐 Wiki Viewer。产品锁上表验收。
- 2026-10-26 · **开 MD-Q1** 基 FE `9d1feba`（R1–R3 齐）。
- 2026-10-26 · React：描述默认 `WikiBodyViewer` 预览、点编辑进 ByteMD、blur 回预览；评论列表同 Viewer。待产品开 MD-Q1。
