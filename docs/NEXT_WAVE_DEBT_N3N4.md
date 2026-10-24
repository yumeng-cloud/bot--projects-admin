# 下一小包 · N3/N4

> 2026-10-24 17:41 Asia/Shanghai · 产品体验  
> 前提：N1/N2 `ec1742d` 产品+QA PASS；Nits PASS

## N3 · J4 force 审计可见 comment（优先）

| | |
| --- | --- |
| **谁** | **Go 后端**（必要时 React 只展示已有 comment） |
| **债** | ~~J4 现仅 `seat.force_bind` activities；**Issue/席位正文 comment** 仍 P0.1~~ → **已清**（本波 Go） |
| **做** | 停工/离职席（及脏绑）`force` 成功 → 人话 `seat.force_bind` activity + 关联开放 Issue **comment**（谁、哪席、原状态→占用、force） |
| **不做** | 活跃席 force；改抢席语义 |
| **验收** | 箱：停工席 force 后，`GET …/activities` 见人话句；Issue 抽屉评论见同文案；单测 `TestBindSeatStoppedForceWritesVisibleAudit` / `…Left…` / `…ActiveForceNoVisibleAudit` |

## N4 · 下发后「立即推送」可选（J7 P0.5 一刀）

| | |
| --- | --- |
| **谁** | **React 前端** + **Go**（若缺「对席位 member 触发 webhook/通知」薄接口则补） |
| **做** | N2「下发任务」Modal 增勾选「立即推送」（默认关）；勾选则创建指派后走现有 Webhook 扇出（失败 Toast 人话，不挡建单） |
| **不做** | 巡检例程 UI；重试队列产品面 |
| **验收** | 箱：勾选推送建单成功；未勾选仅建单；截 `N4-dispatch-push` |

## 暂缓

| 项 | 原因 |
| --- | --- |
| 附件 §C 再开发 | F1–F10 代码已齐；另开「I3 箱截产品关单」不挡本包 |
| 巡检例程 | 通道域，非本波 FE |

## 派发

| # | 项 | 席位 |
| --- | --- | --- |
| 1 | N3 force 审计 comment | **Go 后端** |
| 2 | N4 立即推送可选 | **React**（接口缺再喊 **Go**） |


---

## 产品勾选（2026-10-24 17:46 Asia/Shanghai · 17:48 补强）

| 项 | 结果 | 证据 |
| --- | --- | --- |
| **N3 / J4 P0.1** | [x] **关** | `3c698b5`：stop→`force` bind → `GET …/activities` `action=seat.force_bind` 人话 comment；开放 Issue `GET …/comments` 同人话（Go 箱证 + 用户放行推 main） |
| **N4** | [x] **关** | `c947a76` + `b0e4dc0` · `docs/qa-shots/N4-dispatch-push.png`（立即推送已勾 + 说明） |

**清债**：J4 P0.1 + J7「立即推送」+ **巡检 P0.5**（`e6e4402`/`ymbot-patrol-p05-20261024.txt`）已清。仍欠：附件 I3 箱截（`ATTACH_I3_BOX_CHECKLIST`）。
