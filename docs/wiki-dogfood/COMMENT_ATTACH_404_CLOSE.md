# COMMENT_ATTACH_404 关刀纪要

> 2026-10-26 · 产品体验 · 狗粮 Wiki

## 目标

打开 Issue 详情拉评论附件时，缺评/孤儿 ID 不得弹「评论不存在」Toast（含双弹）。

## 人话验收

| 行 | 结果 |
| --- | --- |
| FE skipToast | PASS†（listComment / listIssue） |
| Go ListByComment 软空 | PASS†（缺评 200[]；Upload 仍 404） |
| 有评论详情零 Toast | PASS†（YM-27；重开/F5 仍 0） |

## 锚

| 项 | SHA / 证据 |
| --- | --- |
| FE | `1a303c4` |
| Go | `47822fe` |
| 债开刀 | `5252e7a` |
| 开 CA-Q1 | `a54dd1e` |
| CA-Q1 PASS† | `49bb1b8` · `docs/qa-shots/COMMENT_ATTACH_404-*` |
| P1 | 本页 |

## Soft†（不挡关）

中间评论字面 `\n`（未渲染 MD）；另票。

## 非目标（未动）

Soft / INVITE / LANDING / MOBILE_V2 / ISSUE_MD_RENDER。
