# 巡检例程 P0.5 · 最小范围

> 2026-10-24 17:46 Asia/Shanghai · 产品体验  
> 对齐：`BOT_TEAM_JOIN_KEY` §4.2 辅路径；Webhook 主路径已 PASS

## 目标（一句话）

席位 Bot **定时**拉「指派给我且未完成」；**无变化安静**；补 Webhook 丢包；**禁止**抢席/误叫已 leave 的旧 UUID。

## 本波最小（只这些）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **P1** | `ymbot`（或 MCP）一条命令/子命令：按当前绑定席位列出 open+assignee=me（或等价 inbox） | **Bot 记忆角色**（通道） |
| **P2** | 文档：建议例程 cron **5–15min**；伪代码 = `patrol` → 若有新指派则 `handoff get` + 领叶子；**零新单则 exit 0 无刷屏** | **Bot 记忆角色** |
| **P3** | 箱内自测：造 1 条指派 Issue → 跑 patrol 有输出；再跑一次无变化 → 安静（证据 txt） | **Bot 记忆角色** + **QA** 可辅 |
| **P4** | （可选）Go：若缺「指派给我」列表查询参数则补薄接口；有则不改 | **Go** 仅缺时 |

## 明确不做

- 巡检产品设置页 / 开关 UI  
- 重试队列产品面  
- 席位私有 Webhook URL  
- 替代 Webhook 主路径  

## 成功

- 证据路径写入 `BOT_TEAM_JOIN_KEY` J7 脚注「巡检 P0.5」  
- 产品勾本文件下方  


## 怎么跑

| 步 | 打开哪 / 怎么做 | 结果 |
| --- | --- | --- |
| 1 环境 | `YMBOT_API_URL` + `YMBOT_API_KEY`（ymk_）+ `YMBOT_WORKSPACE_ID`；可选 `YMBOT_PATROL_STATE_FILE` | 同主航线鉴权 |
| 2 CLI | `ymbot patrol` 或 `ymbot issue patrol` | 列出 **open**（todo/in_progress/in_review）且 **assignee=me** |
| 3 安静 | `ymbot patrol --quiet-if-unchanged [--state-file PATH]` | fingerprint 无变化 → **无 stdout、exit 0**；有新指派才刷表 |
| 4 MCP | `board_patrol`（`quiet_if_unchanged` / `state_file`） | 同 CLI；安静时 content 空 |
| 5 cron | 建议 **5–15min**；可用 `scripts/seat-patrol.sh` 调同一二进制 | 辅路径；**不替代** Webhook |
| 6 有新单 | `handoff get` → 领叶子干活 | 与主航线一致 |
| 7 主路径 | Webhook / 席位「立即推送」`POST …/issues/{id}/notify`（N4） | 实时下发；巡检只补漏 |

### 例程伪代码

```text
every 5..15 min:
  out = ymbot patrol --quiet-if-unchanged
  if out empty: exit 0          # 零新单，无刷屏
  ymbot handoff get
  # 按 out 中叶子 claim / 开工
```

禁：巡检当抢席；对已 leave 旧 UUID `SendToAgent`。

## 箱证

- 证据：`docs/qa-shots/ymbot-patrol-p05-20261024.txt`
- J7 脚注：`docs/BOT_TEAM_JOIN_KEY_P0.md`「巡检 P0.5」

## 产品勾选

- [ ] P1–P3 交件 · 产品勾 · J7 巡检债清
