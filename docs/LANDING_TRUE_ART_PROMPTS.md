# LANDING 真插图 · 生成提示词（L-V-SPEC · 气质重开）

> 视觉总监 · 2026-10-25 ~21:12 Asia/Shanghai  
> **打回**：主人 21:10——首屏像「弱项目看板演示」、没兴趣用。`b200881` 六图 **不作放行证据**。  
> 专册：[`SITE_LANDING_VISUAL_BASELINE.md`](./SITE_LANDING_VISUAL_BASELINE.md)  
> 落盘：`frontend/src/pages/Landing/assets/` · 文件名仍锁下表（覆盖旧 PNG）

## 0. 英雄气质一句话

对标 **飞书 / Notion 中文官网 Hero**：第一眼是「敢用的产品品牌页」，不是后台截图贴海报。  
须一眼读出：**人 + Agent 一支队 · 多机分工 · 统一调度**。

## 1. 共用风格锁

| 要 | 禁 |
| --- | --- |
| 浅色 `#F2F4F8` / 白表面；主色 `#5E6AD2` | 暗黑、径向 glow、赛博霓虹 |
| 软扁平 / 轻等距 **营销插画 + 产品 UI 合成** | **平庸三列看板截图感**（整图只有待办/进行/完成） |
| 专业协作产品感（班底、席位、调度） | **教学作业皮**（老师/课件/作业/教辅/课堂） |
| 主名可写 **育蒙协作看板** | 飞书 / Lark / Notion **品牌 Logo** |
| 人与 Agent 视觉同场（人形 + 简洁 Agent 点/牌） | 恐怖机器人海报、emoji 墙、英文主 slogan |

## 2. 六张画面锁（重生覆盖）

| 文件名 | 画幅 | 构图硬锁（重生必读） |
| --- | --- | --- |
| **`landing-hero-board.png`** | **16:9 优先**（可 4:3） | **英雄主视觉**：分层合成——前景「人 + Agent 席位条/角色牌」；中景「多机工位或调度枢纽」；远景可带 **局部** 看板碎片（≠整屏三列截图）。气质像飞书官网大图：有纵深、有品牌感。**禁**单独一张瘦三列看板窗。 |
| **`landing-why-compare.png`** | 4:3 | 左：多聊窗杂乱（痛）；右：席位班底整齐（解）。对比清晰，非教学场景。 |
| **`landing-team-board.png`** | 4:3 | **一支队**：人与 Agent 围同一工作台/共享上下文；可含看板+Wiki 一角，但主体是「同场协作」不是列卡片墙。 |
| **`landing-multi-machines.png`** | 16:9 | 三机（审图/写码/主持）→ **统一调度** 枢纽。强调分布式工位，禁「换机续会话」隐喻。 |
| **`landing-seats-keys.png`** | 4:3 | 席位占用表 + **一把**按需钥；专业设置感；禁堆墙码、禁教室排座。 |
| **`landing-flow-steps.png`** | 16:9 | 四步横向：建班底→各机认领→看板推进→统一调度；营销流程图，非教辅时间轴。 |

## 3. GenerateImage 提示词骨架（可微调，勿漂气质）

**Hero（最重要）**

> Premium Chinese B2B SaaS landing hero illustration for product 育蒙协作看板, Feishu/Notion homepage quality. Layered composition: foreground human and agent role chips as one crew; midground three workstation nodes merging into a unified scheduling hub; soft depth; accent #5E6AD2; light gray #F2F4F8; polished marketing art not a raw screenshot. MUST show human+Agent one team and multi-machine scheduling. FORBIDDEN: plain three-column kanban screenshot only, teaching classroom, homework, school, Feishu logo, dark neon cyber.

**Why**

> Split comparison: left chaotic overlapping chat windows; right calm seat roster crew panel. Soft flat Feishu-Notion style, purple #5E6AD2, light background. No classroom, no dark neon.

**Team**

> Humans and simple agent avatars collaborating around one shared workspace (board fragment + context), product 育蒙协作看板, quiet premium marketing illustration, light theme #5E6AD2. Not a full kanban screenshot dump. No school theme, no Feishu logo.

**Multi**

> Three labeled workstations 审图 写码 主持 flowing into hub 统一调度. Soft flat premium SaaS marketing, light, #5E6AD2. Distributed crew not session handoff. No neon.

**Seats**

> Professional seat occupancy table with human/agent tags plus one on-demand key card. Clean Feishu-Notion settings aesthetic, light, #5E6AD2. No QR wall, no classroom seating chart.

**Flow**

> Horizontal four-step premium process: 建班底 → 各机认领 → 看板推进 → 统一调度. Soft line, numbered, #5E6AD2, light marketing infographic. No textbook timeline.

## 4. 验收（气质）

| 检 | 过 | 不过 |
| --- | --- | --- |
| Hero 第一眼 | 想点登录的产品官网 | 「又一个弱看板演示」 |
| 叙事 | 人+Agent / 多机 / 调度 可读 | 只有三列任务卡 |
| 皮 | 专业协作 SaaS | 教学/作业/教辅 |
| 真图 | 静态 PNG 在 `Landing/assets/` | CSS mock / 色块假图 |

双帧 SL-V1/V2 → 视觉总监 SL-V3 → 再 QA。前次 `b200881` 图 **作废重交**。
