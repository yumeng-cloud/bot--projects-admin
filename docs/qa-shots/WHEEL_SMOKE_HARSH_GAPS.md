# 车轮修复 · 主航线苛刻冒烟缺口（QA 预列）

> 2026-10-24 21:11 Asia/Shanghai · QA  
> 总指挥令：产品裁债中；先列主航线苛刻冒烟缺口，**等债文立刻验**。  
> 范围：登录 / 队切换 / Issue / 归档 / Bot 席。不做全仓回归。

## 证据
- 主：`docs/qa-shots/WHEEL_SMOKE-qa-retest-YYYYMMDD.txt`
- 箱截按债文命名；无债名前缀用 `wheel-*.png`

---

## 1 · 登录 / 会话
| ID | 缺口 / 验点 | 近史 | 债到后 |
|----|-------------|------|--------|
| L1 | 账密登录 → `/auth/me` 200；错误密码中文 4xx | 基线 | |
| L2 | JWT 过期/伪造 → 401；刷新后可再登 | 基线 | |
| L3 | 无 WS 用户 → 开通引导（非 seed 死空） | WS_ADMIN A4 | |
| L4 | 平台超管旗标：`is_platform_admin` 仅超管见「平台管理」 | C11 PASS `198d019` | |

## 2 · 队切换（顶栏）
| ID | 缺口 / 验点 | 近史 | 债到后 |
|----|-------------|------|--------|
| T1 | 顶栏可见当前队；选项来自 `listTeams` | TEAM_SWITCH S1 | |
| T2 | 切队后 **list/board 池**跟 `acb.team.<ws>` | S2 曾回滚再修 | |
| T3 | 切队后 **Bot 席**只显示当前队 roster（非全量 / 无 slug 双份） | S2 PASS `71cc49c`（14↔5） | |
| T4 | 周期/排期/Wiki/工时同读当前 teamId | S3 路径勾 | |
| T5 | pins 墙不串队 | P4 PASS | |

## 3 · Issue（池 / 详情 / 串队）
| ID | 缺口 / 验点 | 近史 | 债到后 |
|----|-------------|------|--------|
| I1 | Create 必 `team_id`；错队 → 404 | T6-I PASS | |
| I2 | 默认 list 排除 `team_id IS NULL`（未分队不混入） | T6-I | |
| I3 | 详情主内容可读写；空描述 hug（ISSUE_DETAIL 已关） | 产品关 | |
| I4 | 乐观锁 / 看板拖拽不丢（冒烟级） | 回归债 | |

## 4 · 归档
| ID | 缺口 / 验点 | 近史 | 债到后 |
|----|-------------|------|--------|
| A1 | `POST …/issues/archive` 仅 done；非 done → 409 中文 | D2 PASS `aec8632` | |
| A2 | 默认 list/board/search **不见**归档 | D2 | |
| A3 | `include_archived=1` / `archived=only` 可拉 | D2 | |
| A4 | pack/周期 done/total **剔**归档 | D2（2/4→0/2） | |
| A5 | **软债**：ListRelations 边仍可指归档 id（产品已记） | 软注 | 不挡 |

## 5 · Bot 席
| ID | 缺口 / 验点 | 近史 | 债到后 |
|----|-------------|------|--------|
| B1 | roster 按当前队过滤（同 T3） | TEAM_SWITCH | |
| B2 | 双队 claim/bind **不串**（team-scoped `bot_role`） | T6-R PASS | |
| B3 | 从模板创建空席；未知模板 400 | T7-R PASS | |
| B4 | 人类 `ymw_` 邀请 ≠ 队席 `ymt_` | WS_ADMIN INV | |
| B5 | **软债**：WS 级 bot_roles 扁平拼历史席（非记忆墙红线） | T7-R 非挡 | 不挡 |

---

## 冒烟顺序（债文到后）
1. 热更 API + FE → healthz  
2. L1 → T1–T3 → I1–I2 → A1–A4 → B1–B2（红线优先）  
3. 其余轻抽；有红开缺陷给对应同事  

## 关报
短报总指挥+产品：Overall · 红线表 · 证据 · HEAD。测绿可 push main。
