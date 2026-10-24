# 附件 I3 · 箱 5174 产品关单清单

> 2026-10-24 17:46 Asia/Shanghai · 产品体验  
> 依据：`ACCEPTANCE_UX_BOT.md` §C F1–F10 代码已齐；`PRODUCT_SUITE_P0` **I3** 产品箱截已关（§C 全表仍等 QA）  
> **本单只箱截关单，禁止重写上传业务**  
> **React 入口**：本文件 A1–A5（`docs/qa-shots/I3-*.png`）· 5174 热更自截  
> QA 备测全表见 `ATTACHMENT_C_CLOSE_CHECKLIST.md`（F1–F10）；产品关 I3 以本五张为准

## 谁

| 角色 | 做 |
| --- | --- |
| **React 前端** | 箱 5174 热更后自截下表 5 张；有洞只修热更/入口，不扩 scope |
| **QA 测试** | 对照同入口复验；PASS/FAIL 短报产品 |
| **产品体验** | 目视勾 I3 + §C 联调收尾 |

## 箱截（文件名固定）

| ID | 操作 | 截图 |
| --- | --- | --- |
| **A1** | Issue 抽屉 → 描述旁「添加附件」→ 上传 **图片**（png/jpg）→ 列表见缩略+文件名+大小 | `docs/qa-shots/I3-issue-image.png` |
| **A2** | 同抽屉上传 **文档**（pdf 或 txt）→ 可点下载 | `docs/qa-shots/I3-issue-doc.png` |
| **A3** | 图片行可 **预览**（lightbox/Image） | `docs/qa-shots/I3-image-preview.png` |
| **A4** | 评论 composer 回形针 → 评论挂附件成功 | `docs/qa-shots/I3-comment-attach.png` |
| **A5** | 选超限/禁类型 → **中文**失败 Toast，列表不脏 | `docs/qa-shots/I3-reject-toast.png` |

## 验收

- 五张齐且目视通过 → 产品勾 `PRODUCT_SUITE_P0` **I3** + 本文件下方  
- 失败只报一条洞，React 热修后再截；**不**新开附件大单

## 产品勾选

- [x] A1–A5 箱截齐 · 产品勾 · 关 I3（2026-10-24 17:58 Asia/Shanghai）
  - React `c749195`（A4 热修 `a41e976`）· `docs/qa-shots/I3-*.png`
  - 目视：A1 图+列表；A2 pdf 蓝链；A3 lightbox；A4 评论挂图；A5「不支持的文件类型」Toast、列表不脏
  - §C F1–F10 全表仍由 QA 按 `ATTACHMENT_C_CLOSE_CHECKLIST` 硬刷，不挡本 I3 产品勾


## QA 复验 · 2026-10-24 17:58 Asia/Shanghai

| ID | 结果 | 备注 |
| --- | --- | --- |
| A1 | **PASS** | `I3-issue-image.png` 缩略+文件名+大小 |
| A2 | **PASS** | `I3-issue-doc.png` 文档行可下 |
| A3 | **PASS** | `I3-image-preview.png` antd lightbox 1/1+缩放条 |
| A4 | **PASS** | `I3-comment-attach.png` 评论挂附件 |
| A5 | **PASS** | `I3-reject-toast.png` 中文「不支持的文件类型」 |
| **Overall A1–A5** | **PASS** | HEAD `c749195` · 对齐 A4 `a41e976` · 短报产品体验 |

> 产品勾选仍由产品体验目视勾下方 checkbox / PRODUCT_SUITE_P0 I3。
