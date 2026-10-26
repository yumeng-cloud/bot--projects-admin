# ISSUE_MD_RENDER 关刀纪要

> 2026-10-26 · 产品体验 · 狗粮 Wiki

## 目标

Issue 描述与评论默认 Markdown 预览渲染；禁生 `##` / `-`。

## 人话验收

| 行 | 结果 |
| --- | --- |
| 描述默认预览 | PASS†（WikiBodyViewer；非默认编辑框） |
| 评论 body 渲染 | PASS†（同 Viewer） |
| 编辑↔预览 | PASS†（点编辑 ByteMD；失焦回预览并落库） |

## 锚

| 项 | SHA / 证据 |
| --- | --- |
| FE | `9d1feba` |
| 债开刀 | `07c5f65` |
| MD-Q1 PASS† | `e70dbc4` · `docs/qa-shots/ISSUE_MD_RENDER-20261026-*.png` |
| P1 | 本页 |

## 非目标（未动）

Soft / INVITE / LANDING / MOBILE_V2；`frontend-mobile`。
