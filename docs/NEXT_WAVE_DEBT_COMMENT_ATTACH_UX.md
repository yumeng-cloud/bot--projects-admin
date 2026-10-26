# P0 · 评论/Issue 附件 UX 现代化

> 2026-10-26 Asia/Shanghai · 产品体验主持席 · 主人 t903u 狗粮吐槽 · 总指挥开议  
> **刀名：** COMMENT_ATTACH_UX  
> Soft/INVITE/LANDING/MOBILE_V2/ISSUE_MD_RENDER/COMMENT_ATTACH_404 **勿动**。不问用户。

## 气质一句

评论挂文件像 Linear / 飞书 / GitHub：**轻图标 + 芯片预览**，能办事；**禁**老 OA「添加附件 / 挂附件」大字文案按钮。

## 拍板（主持席 · 无反对即通过 · 已死）

总指挥建议 1–5 **全收**，细化为下表。

| # | 允许 | 禁止 |
| --- | --- | --- |
| **CAX-1 Composer** | 评论框工具条仅 **回形针图标** + tooltip「添加附件（≤10MB）」；待发文件 **芯片**（缩略图或文件名）可删；发送时一并上传 | 文案按钮「添加附件 / 附件 / 挂附件」 |
| **CAX-2 粘贴/拖拽** | 评论框内 **粘贴图片**、**拖文件进评论区** → 落 pending 芯片，随发送上传 | 仅靠点按钮才能挂文件；拖进整页无关区域乱弹 |
| **CAX-3 已发评论** | 有附件：缩略图/紧凑列表；**补挂**仅 hover 显轻量回形针，或行内「…」菜单「添加附件」 | 每条评论底下 **常驻**「挂附件」文字按钮 |
| **CAX-4 Issue 级** | `IssueAttachmentSection` 同气质：空态轻图标/一句弱提示；有文件列表 + 同款回形针；与评论共用上传控件气质 | 另搞一套 Form；空态大字「添加附件…」链接味 |
| **CAX-5 契约** | 仍走现有 `uploadCommentAttachment` / Issue upload；≤10MB；skipToast 只读 list 保持 | 改后端契约、新 API、炫技动效/花哨库 |

## 验收

| ID | 标准 | 谁 |
| --- | --- | --- |
| **CAX-R1** | Composer：无「添加附件/挂附件」文案钮；仅回形针 + pending 芯片可删 | React → QA |
| **CAX-R2** | 粘贴图 / 拖文件进评论区 → pending；发送后出现在该评论附件 | React → QA |
| **CAX-R3** | 已发评论无常驻「挂附件」；hover/… 可补挂；有附件见缩略/列表 | React → QA |
| **CAX-R4** | Issue 级附件区同气质，无「添加附件…」老链接味 | React → QA |
| **CAX-V1** | 视觉自审（密度/tooltip/芯片）；必要时交 UI视觉·审图 | React → 审图 |
| **CAX-Q1** | 箱全链路：composer 挂图发送 + 已发补挂 + Issue 区；禁甩主人 | QA |
| **CAX-P1** | 产品关 | 产品体验 |

## 执刀

- **React 前端**（`frontend/`：`IssueAttachmentUpload` / `IssueCommentList` / `IssueAttachmentSection` 等）
- 测绿直推 `origin/main`；后端契约不动（Go 无刀除非发现硬挡）
- 视觉：先自审；拿不准再拉 UI视觉·审图

## 非目标

- 不重做评论编辑器 / ByteMD；不改 MD 渲染（ISSUE_MD_RENDER 已关）
- 不动 `frontend-mobile`；不解冻 Soft；不改 LANDING/INVITE
- 不做多图灯箱大重构（缩略可点打开/下载即可）

## 产品勾选

- [x] **CAX-R1** Composer 图标+芯片
- [x] **CAX-R2** 粘贴/拖拽
- [x] **CAX-R3** 已发评论无常驻文案钮
- [x] **CAX-R4** Issue 级同气质
- [x] **CAX-V1** 审图 PASS† @`00fc54d`（软非挡 pending gap/触控发现性）
- [ ] **CAX-Q1** PASS†
- [ ] **CAX-P1** 产品关

**状态：开刀中** · FE `00fc54d` · R1–R4/V1 齐 · **已开 CAX-Q1**。

## 进度注记

- 2026-10-26 · 主人 t903u：评论附件「添加附件」像老系统；要好用能办事、禁 Form 味。
- 2026-10-26 · 主持席拍死上表；执刀 React；关刀前 CAX-Q1 PASS†。
- 2026-10-26 · React：CAX-R1–R4 勾上（Composer 仅回形针+pending 芯片；粘贴/拖拽；已发 hover 补挂；Issue 区同气质）。tsc+vite 绿。
- 2026-10-26 · **勾 CAX-V1** UI视觉·审图 PASS† @`00fc54d`（软非挡：pending gap、触控 hover）。
- 2026-10-26 · **开 CAX-Q1** 基 FE `00fc54d`（R1–R4+V1 齐）。
