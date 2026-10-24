#!/usr/bin/env bash
# 席位巡检辅路径：调 ymbot patrol（P1 以二进制/MCP 为准；本脚本仅 cron 包装）。
# 用法：YMBOT_API_KEY=ymk_… YMBOT_WORKSPACE_ID=… ./scripts/seat-patrol.sh
# cron 建议 */10 * * * *  （5–15min）；有新指派再 handoff get + 领叶子。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
YMBOT="${YMBOT:-$ROOT/backend/ymbot}"
if [[ ! -x "$YMBOT" ]]; then
  YMBOT="$(command -v ymbot || true)"
fi
if [[ -z "${YMBOT}" || ! -x "$YMBOT" ]]; then
  echo "seat-patrol: 找不到 ymbot（设 YMBOT= 或先 go build -o backend/ymbot ./cmd/ymbot）" >&2
  exit 1
fi
STATE="${YMBOT_PATROL_STATE_FILE:-${HOME}/.ymbot/patrol-state}"
exec "$YMBOT" patrol --quiet-if-unchanged --state-file "$STATE" "$@"
