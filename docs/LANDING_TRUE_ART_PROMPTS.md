# LANDING 真插图 · 生成提示词（L-V-SPEC 交件）

> 视觉总监 · 2026-10-25 ~21:00 Asia/Shanghai  
> 专册：[`SITE_LANDING_VISUAL_BASELINE.md`](./SITE_LANDING_VISUAL_BASELINE.md) @`4d6073e` §4.1  
> 落盘：`frontend/src/pages/Landing/assets/` · 文件名锁下表

## 共用风格锁

- 气质：飞书 / Notion 中文产品介绍页；浅色 `#F2F4F8` 底；主色点缀 `#5E6AD2`
- 软扁平 / 轻 UI 场景；圆角；轻阴影
- **禁**：暗黑、径向 glow、赛博机器人海报、飞书/Lark **品牌 Logo**、英文主 slogan、CSS 色块假图
- 主名可出现：**育蒙协作看板**

## 六张（GenerateImage 后另存为锁名）

| 文件名 | 画幅 | 提示词要点 |
| --- | --- | --- |
| `landing-hero-board.png` | 4:3 | Hero 看板三列 待办/进行/完成；中文卡；人与 Agent 同场；产品名育蒙协作看板；无他牌 Logo |
| `landing-why-compare.png` | 4:3 | 左多聊窗杂乱 vs 右席位班底整齐对比 |
| `landing-team-board.png` | 4:3 | 一支队共享看板+知识库安静场景；育蒙协作看板；禁飞书 Logo |
| `landing-multi-machines.png` | 16:9 | 机A审图 机B写码 机C主持 → 统一调度 |
| `landing-seats-keys.png` | 4:3 | 席位表 + 按需钥；勿堆墙码 |
| `landing-flow-steps.png` | 16:9 | 四步：建班底→各机认领→看板推进→统一调度 |

## 验收

审图落盘 + React `<img>` 接线 + 下线 mocks 主视觉 → 双帧 SL-V1/V2 → 视觉总监 SL-V3 → 再 QA。
