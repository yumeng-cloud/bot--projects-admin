# F1 / T3 / S + P1–P5 箱内验收短报

时间：2026-10-24（Asia/Shanghai）· 箱 `5174` / BE `8080` · ws `yumeng`

## 交付

| 项 | 状态 |
| --- | --- |
| F1 加入班底向导（ymk_/ymt_、选席、成功→交接） | 已接线；invite 真 404 →「入场码服务尚未开通」+ 本区班底演示 |
| Admin「复制入场码」 | 已接线；404 人话；明文仅签发瞬间 Toast，列表末四 |
| T3 占用专面（无顶替） | 占用席可点进冲突面；claim 409 → Drawer |
| S 密度 / CLI 降级 | `bots.css` token；换机 Alert→Collapse 默认收起；Detail 去重 |
| P1–P5 看板 | bar→pack 两行已有；board.css token 保留；列模板 Popover |

## 截图（`docs/qa-shots/`）

`F1-empty` · `F1-wizard` · `F1-roster-pick` · `T3-conflict` · `S-bots-roster` · `P-pack-strip` · `P-col-template`

## 约束遵守

- 未覆盖/stub `bots.css`（保留 UI section-gap12 / card-pad10×12 / grid-min260 / row36）
- 未回退 `board.css` densify token；BoardPage TSX 未改（已是 bar→pack）
- 前端 only；禁 clone/CloudAgent

## 已知

- `/team-invites/redeem` 与 invites 签发仍 404（BE G3）；FE client + 优雅降级可演示选席/T3
- 空态截图用 `?f1empty=1`；线上无绑定时同空态
