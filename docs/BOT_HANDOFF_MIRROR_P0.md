# Bot 交接旁路 · handoff 定时镜像（P0）

> 状态：独立节点/插件脚本已落地（2026-10-24 · Asia/Shanghai）  
> **定位：备份旁路**，不是续命主路径。  
> 权威续命：[`BOT_HANDOFF_CONTINUITY_P0.md`](./BOT_HANDOFF_CONTINUITY_P0.md) · 短原则 [`BOT_CONTINUITY_P0.md`](./BOT_CONTINUITY_P0.md)

---

## 0. 一句话

续命仍靠服务端 **`handoff get`**；本旁路 = **定时只读**把各 Bot Key 的 handoff 全量 JSON 冷备到磁盘，**进场才读**（新会话/救场），日常不替代 get，**不写服务端**、不狂刷 memory。

---

## 1. 与主路径关系

| | 主路径（续命） | 本旁路（镜像） |
|--|----------------|----------------|
| 权威 | 服务端实时组装的 handoff | 磁盘快照 `latest.json` / `handoff-<ts>.json` |
| 何时用 | 每次新会话 / 换机续工 | **仅**服务端不可达、或需冷备审计 / 救场对照 |
| 动作 | `ymbot handoff get` / `board_handoff_get` | 定时 **GET only**（`handoff export --out` 或 REST GET） |
| 写服务端 | 干活中 memory/todo/comment | **禁止**（无 set / comment / claim / import） |

旁路 ≠ 主路径。脚本失败只记日志，绝不 import 回写。

---

## 2. 如何跑

**脚本**：`scripts/handoff-mirror.sh`  
**密钥示例**：`scripts/handoff-sidecar/bot-keys.example` → 复制为 `bot-keys.local`（gitignore）  
**cron 示例**：`scripts/handoff-sidecar/cron.example`

```bash
# 用法一行
BOT_KEYS_FILE=./scripts/handoff-sidecar/bot-keys.local \
YMBOT_API_URL=http://127.0.0.1:8080 \
YMBOT_WORKSPACE_ID=<workspace_id> \
MIRROR_DIR=/var/ymbot/handoff-mirror \
./scripts/handoff-mirror.sh
```

| 变量 | 说明 |
|------|------|
| `BOT_KEYS_FILE` | 每行 `name=ymk_…`（推荐多 Bot） |
| `YMBOT_API_KEY` + `BOT_NAME` | 单 Key 模式 |
| `YMBOT_API_URL` | 默认 `http://127.0.0.1:8080` |
| `YMBOT_WORKSPACE_ID` | 必填 |
| `MIRROR_DIR` | 默认 `/var/ymbot/handoff-mirror`（可配仓库外路径） |
| `KEEP_N` | 每 Bot 保留最近 N 份，默认 24；`0`=不删 |
| `INTERVAL_SEC` | `--loop` 间隔，默认 900（15min）；建议 ≥5–15min |
| `YMBOT_BIN` | 可选；否则 PATH 或 `backend/ymbot` |
| `LOG_FILE` | 可选；空则 stderr |

标志：`--once`（默认）· `--loop` · `--dry-run`（校验配置，不调 API）· `--curl`（强制 REST）

输出布局：

```
$MIRROR_DIR/<bot_name>/
  handoff-20261024T121500+0800.json
  latest.json          # 软链（或拷贝）→ 最新一份
```

### 定时示例

```bash
# cron：每 15 分钟一轮（见 cron.example）
*/15 * * * * … ./scripts/handoff-mirror.sh --once

# 或常驻循环
INTERVAL_SEC=900 … ./scripts/handoff-mirror.sh --loop
```

### 缺配置

无 Key / 无 `YMBOT_WORKSPACE_ID` → 清晰报错，**exit ≠ 0**（`--dry-run` 同样校验）。

---

## 3. 进场才读

新会话 / 救场时，若服务端 `handoff get` 不可用，才打开：

`$MIRROR_DIR/<bot>/latest.json`

对照或人工决策是否 `handoff import`（import 仍是救场非常用，见续命文档）。**日常续工不要读镜像目录。**

---

## 4. 非目标

- 不替代服务端 `handoff get` 成为日常续命  
- 不在定时任务里 `memory set` / comment / claim / import  
- 不把 API Key 写入快照或 git（密钥文件权限建议 `600`）  
- 不打断主链路、不做 fan-out / 不改 Issue handler  

---

## 5. 验收速记

- `bash -n scripts/handoff-mirror.sh`  
- 无 Key：`--dry-run` 或缺配置 → exit ≠ 0 且文案清晰  
- 有 Key：实跑一轮，见 `$MIRROR_DIR/<name>/latest.json`；失败只打日志  
