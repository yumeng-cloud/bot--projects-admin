# 附件 §C · 关单清单 F1–F10

> 依据：`ACCEPTANCE_UX_BOT.md` §C  
> 证据目录：`/workspace/qa-shots/attach-c-close-25797ce/`（先验 `attach-c-close-3a3e294` Overall PASS 可对照）  
> 环境：API :8080 · vite :5174 · WS `ba1481c49148aba9553ba3811406b69f` · project `b1e996fb8ba9c350e249f7a862425a67` · admin@yumeng.tech  
> 联调 Issue：`8095f40326477e0607259be145285bb0`（YM-15）· Comment：`bfae3f2a1492344a80446845dcd5f289`  
> 关单时刻：2026-10-24 18:00 Asia/Shanghai · HEAD `f0c2642`（产品勾关对齐 `f0c2642`；API 硬刷证据基于 `25797ce` 联调栈）  

## 结果总览

| 项 | 结果 |
| --- | --- |
| `go test ./internal/attachments` | **PASS** |
| JWT API F1–F2,F4–F8,F10 + F3_API（live :8080 重跑） | **PASS** |
| FE F9（紧凑列表截图） | **PASS** |
| FE F3（图片预览 lightbox 截图） | **PASS**（`c749195` 箱截复验；I3 产品已勾） |
| **Overall** | **PASS**（产品 `f0c2642` 勾关对齐） |

## F1–F10

| ID | 标准 | 结果 | 证据 |
| --- | --- | --- | --- |
| **F1** | Issue 级上传 png/pdf | **PASS** | `f1-upload-png.json` `f1-upload-pdf.json` |
| **F2** | 评论挂附件 | **PASS** | `f2-comment-upload.json`（comment_id 挂载） |
| **F3** | 图片预览 | **PASS** | API：`f3-headers.txt` inline；FE：`docs/qa-shots/I3-image-preview.png` + `F3-fe-lightbox.png` |
| **F4** | 文档下载 | **PASS** | `f4-headers.txt`（`attachment; filename="sample.pdf"`）+ `f4-content.pdf` |
| **F5** | 类型限制中文 | **PASS** | `f5-reject.json` → `不支持的文件类型`（产品抽核） |
| **F6** | 大小限制中文 | **PASS** | `f6-reject.json` → `文件过大，上限 10MB`（产品抽核） |
| **F7** | 鉴权同路径 | **PASS** | `f7-noauth.json` 401 `请先登录`；`f7-auth.json` 200 列表 |
| **F8** | 删除权限 | **PASS** | `f8-upload.json` + `f8-delete.json` `{"ok":true}`（uploader JWT） |
| **F9** | 紧凑列表 UI | **PASS** | `F9-compact-list-*.png` + `I3-issue-image|doc|comment-attach|reject-toast.png` |
| **F10** | 失败不脏 / 删后不可下 | **PASS** | `f10-delete.json` + `f10-content-after.json` `附件不存在`；F5/F6 拒后无脏行 |

## Go

```
ok  	github.com/yumeng-tech/ai-collab-board/backend/internal/attachments	0.398s
```

## 关单

- 产品体验已勾关 §C Overall **PASS** at HEAD `f0c2642`（抽核 F5/F6）。
- QA live JWT F1–F10 已写入 `attach-c-close-25797ce/`；FE F3/F9 箱截复用 `c749195` 时代 `docs/qa-shots/`（I3 产品已勾）。
- Overall **PASS** · 对齐产品勾关。

## 产品勾选（对齐）

- [x] §C F1–F10 **关单** · 产品 `f0c2642` + QA 硬刷 `attach-c-close-25797ce/`（对照 `3a3e294`）
- 抽核 F5/F6 中文仍成立；I3 见 `ATTACH_I3_BOX_CHECKLIST`
