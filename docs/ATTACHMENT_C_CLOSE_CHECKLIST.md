# 附件 §C · 关单清单 F1–F10

> 依据：`ACCEPTANCE_UX_BOT.md` §C  
> 证据目录：`/workspace/qa-shots/attach-c-close-3a3e294/`  
> 环境：API :8080 · vite :5174 · WS `ba1481c49148aba9553ba3811406b69f` · project `b1e996fb8ba9c350e249f7a862425a67` · admin@yumeng.tech  
> 联调 Issue：`8095f40326477e0607259be145285bb0`（YM-15）· Comment：`bfae3f2a1492344a80446845dcd5f289`  
> 关单时刻：2026-10-24 17:58 Asia/Shanghai · HEAD `c749195`（A4 `a41e976`）  
> **禁止 push / 禁止 agent 消息**

## 结果总览

| 项 | 结果 |
| --- | --- |
| `go test ./internal/attachments` | **PASS** |
| JWT API F1–F2,F4–F8,F10 + F3_API | **PASS** |
| FE F9（紧凑列表截图） | **PASS** |
| FE F3（图片预览 lightbox 截图） | **PASS**（`c749195` / A4 `a41e976` · antd Image 1/1+缩放条） |
| **Overall** | **PASS** |

## F1–F10

| ID | 标准 | 结果 | 证据 |
| --- | --- | --- | --- |
| **F1** | Issue 级上传 png/pdf | **PASS** | `f1-upload-png.json` `f1-upload-pdf.json` |
| **F2** | 评论挂附件 | **PASS** | `f2-comment-upload.json`（comment_id 挂载） |
| **F3** | 图片预览 | **PASS** | API：`f3-headers.txt` inline；FE：`docs/qa-shots/I3-image-preview.png` + `F3-fe-lightbox.png`（antd Image lightbox 1/1+缩放） |
| **F4** | 文档下载 | **PASS** | `f4-headers.txt`（`attachment; filename="sample.pdf"`）+ `f4-content.pdf` |
| **F5** | 类型限制中文 | **PASS** | `f5-reject.json` → `不支持的文件类型` |
| **F6** | 大小限制中文 | **PASS** | `f6-reject.json` → `文件过大，上限 10MB` |
| **F7** | 鉴权同路径 | **PASS** | `f7-noauth.json` 401 `请先登录`；`f7-auth.json` 200 列表 |
| **F8** | 删除权限 | **PASS** | `f8-upload.json` + `f8-delete.json` `{"ok":true}`（uploader JWT） |
| **F9** | 紧凑列表 UI | **PASS** | `F9-compact-list-image.png` `F9-compact-list-doc.png` `F9-compact-list-ym9.png`（缩略/图标·文件名·大小；YM-9 png+pdf）；同源 `docs/qa-shots/I3-issue-*.png` |
| **F10** | 失败不脏 / 删后不可下 | **PASS** | `f10-delete.json` + `f10-content-after.json` `附件不存在`；F5/F6 拒后无脏行 |

## Go

```
ok  	github.com/yumeng-tech/ai-collab-board/backend/internal/attachments	0.329s
```

## 关单

- F3 FE 已由 React `c749195` 交 `I3-image-preview.png` / `F3-fe-lightbox.png`；QA 复验 **PASS** → Overall **PASS**。

## 未做

- 无 git push
- 无 agent / bot 消息

## 产品勾选（2026-10-24 18:00 Asia/Shanghai）

- [x] §C F1–F10 **关单**（QA Overall PASS · 证据 `/workspace/qa-shots/attach-c-close-3a3e294/` + FE `I3-*.png`/`F3-fe-lightbox.png`）
- 抽核：F5 `不支持的文件类型` · F6 `文件过大，上限 10MB` · F3 lightbox · I3 A1–A5 双 PASS
- I3 产品勾见 [`ATTACH_I3_BOX_CHECKLIST.md`](./ATTACH_I3_BOX_CHECKLIST.md) · `25797ce`
