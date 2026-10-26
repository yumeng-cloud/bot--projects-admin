# P0 · Issue 详情「评论不存在」双 Toast

> 2026-10-26 Asia/Shanghai · 产品体验 · 总指挥报障后拍板  
> **刀名：** COMMENT_ATTACH_404  
> Soft/INVITE/LANDING/MOBILE_V2/ISSUE_MD_RENDER **勿动**。不问用户。

## 一句话

打开 Issue 详情拉评论附件时，缺评/孤儿 ID **不得**弹「评论不存在」Toast（含双弹）。

## 根因（已认）

| 位 | 现况 | 修法 |
| --- | --- | --- |
| FE | `listCommentAttachments` 经全局 `api()`，404 先 Toast 再被 `.catch` 清 | 只读 list 烘焙 `skipToast: true` |
| Go | `ListByComment` → `ensureComment` 404 | 只读缺评回 **200 `[]`**；写路径仍 404 |

## 验收（短 QA）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **CA-R1** | `listCommentAttachments` / `listIssueAttachments` 带 `skipToast` | React → QA |
| **CA-G1** | 缺评 / 错 issue 的 ListByComment → 空列表不 404；UploadComment 缺评仍 404 | Go → QA |
| **CA-Q1** | 箱：打开有评论的 Issue 详情，**零**「评论不存在」Toast（含双弹）；禁甩主人 | QA |
| **CA-P1** | 产品关 | 产品体验 |

## 非目标

- 不改附件上传/删除语义；不重做评论 UI  
- 不动 `frontend-mobile`；不解冻 Soft；不改 LANDING/INVITE/ISSUE_MD

## 产品勾选

- [x] **CA-R1** FE `1a303c4`（skipToast）
- [x] **CA-G1** Go `47822fe`（ListByComment 软空）
- [ ] **CA-Q1** PASS†
- [ ] **CA-P1** 产品关

**状态：开刀中** · FE `1a303c4` + Go `47822fe` · **已开 CA-Q1**。

## 进度注记

- 2026-10-26 · **开 CA-Q1** 基 FE `1a303c4` + Go `47822fe`（CA-R1/CA-G1 齐）。
- 2026-10-26 · **勾 CA-G1** Go @`47822fe` ListByComment 缺评→200[]。
- 2026-10-26 · 总指挥 P0：双 Toast；已钉 React skipToast + Go ListByComment 缺评 200[]。产品锁上表。
- 2026-10-26 · FE 已 main：`1a303c4` · 勾 CA-R1；Go 工作区有软空+测，待测绿推 main。
