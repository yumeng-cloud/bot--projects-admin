#!/usr/bin/env bash
# handoff-mirror.sh — 交接旁路：定时只读镜像各 Bot handoff 快照（备份旁路，不写服务端）
#
# 权威：docs/BOT_HANDOFF_MIRROR_P0.md
# 续命主路径仍是服务端 handoff get；本脚本 = 冷备旁路，进场才读。
# 禁止：memory set / comment / claim / import 等任何写操作。
#
# 用法一行：
#   BOT_KEYS_FILE=./scripts/handoff-sidecar/bot-keys.local \
#   YMBOT_WORKSPACE_ID=<ws> MIRROR_DIR=/var/ymbot/handoff-mirror \
#   ./scripts/handoff-mirror.sh
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# ---------- defaults ----------
YMBOT_API_URL="${YMBOT_API_URL:-http://127.0.0.1:8080}"
YMBOT_WORKSPACE_ID="${YMBOT_WORKSPACE_ID:-}"
MIRROR_DIR="${MIRROR_DIR:-/var/ymbot/handoff-mirror}"
KEEP_N="${KEEP_N:-24}"
INTERVAL_SEC="${INTERVAL_SEC:-900}"   # ≥5–15min；默认 15min
YMBOT_BIN="${YMBOT_BIN:-}"
BOT_KEYS_FILE="${BOT_KEYS_FILE:-}"
LOG_FILE="${LOG_FILE:-}"              # 空 = stderr
MODE="once"                           # once | loop
DRY_RUN=0
USE_CURL=0

usage() {
  cat <<'EOF'
handoff-mirror.sh — 只读镜像 Bot handoff（备份旁路）

环境 / 配置:
  YMBOT_API_URL         默认 http://127.0.0.1:8080
  YMBOT_WORKSPACE_ID    必填
  BOT_KEYS_FILE         推荐：每行 name=ymk_…（# 注释 / 空行忽略）
  YMBOT_API_KEY         单 Key 模式（配合 BOT_NAME，默认 bot）
  BOT_NAME              单 Key 时的目录名（默认 bot）
  MIRROR_DIR            输出根目录（默认 /var/ymbot/handoff-mirror）
  KEEP_N                每个 Bot 保留最近 N 份（默认 24；0=不删）
  INTERVAL_SEC          --loop 间隔秒（默认 900）
  YMBOT_BIN             ymbot 路径；空则 PATH / 仓库 backend/ymbot
  LOG_FILE              日志文件；空则 stderr

标志:
  --once       跑一轮后退出（默认）
  --loop       while sleep INTERVAL_SEC 循环
  --dry-run    校验配置并打印将拉取的 Bot 名，不写盘、不调 API
  --curl       强制用 curl REST（不调用 ymbot）
  -h|--help    本帮助

退出码:
  0  全部成功（或 dry-run 配置齐全）
  1  缺配置 / 用法错误
  2  至少一只 Bot 拉取失败（其它仍尽量完成）

密钥勿写入快照与 git。示例见 scripts/handoff-sidecar/bot-keys.example
EOF
}

log() {
  local line="[$(date '+%Y-%m-%d %H:%M:%S %z')] $*"
  if [[ -n "${LOG_FILE}" ]]; then
    mkdir -p "$(dirname "${LOG_FILE}")"
    printf '%s\n' "${line}" >>"${LOG_FILE}"
  fi
  printf '%s\n' "${line}" >&2
}

die() {
  log "ERROR: $*"
  exit 1
}

sanitize_name() {
  # 目录名：字母数字 _ - . ；其它变 _
  local n="$1"
  n="${n//[^A-Za-z0-9._-]/_}"
  [[ -n "$n" ]] || n="bot"
  printf '%s' "$n"
}

resolve_ymbot() {
  if [[ -n "${YMBOT_BIN}" && -x "${YMBOT_BIN}" ]]; then
    printf '%s' "${YMBOT_BIN}"
    return 0
  fi
  if command -v ymbot >/dev/null 2>&1; then
    command -v ymbot
    return 0
  fi
  if [[ -x "${ROOT}/backend/ymbot" ]]; then
    printf '%s' "${ROOT}/backend/ymbot"
    return 0
  fi
  return 1
}

# 收集 name=key 到并行数组 BOT_NAMES / BOT_KEYS
BOT_NAMES=()
BOT_KEYS=()

load_keys() {
  BOT_NAMES=()
  BOT_KEYS=()
  if [[ -n "${BOT_KEYS_FILE}" ]]; then
    [[ -f "${BOT_KEYS_FILE}" ]] || die "BOT_KEYS_FILE 不存在: ${BOT_KEYS_FILE}"
    local line name key
    while IFS= read -r line || [[ -n "$line" ]]; do
      line="${line%%#*}"
      line="$(printf '%s' "$line" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
      [[ -z "$line" ]] && continue
      if [[ "$line" != *=* ]]; then
        die "BOT_KEYS_FILE 行格式应为 name=ymk_…，收到: ${line:0:40}"
      fi
      name="${line%%=*}"
      key="${line#*=}"
      name="$(printf '%s' "$name" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
      key="$(printf '%s' "$key" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
      [[ -n "$name" && -n "$key" ]] || die "BOT_KEYS_FILE 空 name/key"
      BOT_NAMES+=("$(sanitize_name "$name")")
      BOT_KEYS+=("$key")
    done <"${BOT_KEYS_FILE}"
    [[ ${#BOT_KEYS[@]} -gt 0 ]] || die "BOT_KEYS_FILE 无有效条目: ${BOT_KEYS_FILE}"
    return 0
  fi
  if [[ -n "${YMBOT_API_KEY:-}" ]]; then
    BOT_NAMES+=("$(sanitize_name "${BOT_NAME:-bot}")")
    BOT_KEYS+=("${YMBOT_API_KEY}")
    return 0
  fi
  die "缺配置：请设 BOT_KEYS_FILE（推荐）或 YMBOT_API_KEY；见 docs/BOT_HANDOFF_MIRROR_P0.md"
}

require_config() {
  [[ -n "${YMBOT_WORKSPACE_ID}" ]] || die "缺配置：YMBOT_WORKSPACE_ID"
  load_keys
}


extract_json_str() {
  local file="$1" key="$2"
  if command -v python3 >/dev/null 2>&1; then
    python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get(sys.argv[2],"") or "")' "$file" "$key" 2>/dev/null || true
    return 0
  fi
  if command -v jq >/dev/null 2>&1; then
    jq -r --arg k "$key" '.[$k] // empty' "$file" 2>/dev/null || true
    return 0
  fi
  # 极简兜底：匹配 "key":"value"
  sed -n 's/.*"'"$key"'"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$file" | head -n1
}

# 拉取一只 Bot：只 GET handoff，写入快照。失败返回非 0，不写服务端。
mirror_one() {
  local name="$1" key="$2"
  local dest_dir out_ts out_latest tmp
  dest_dir="${MIRROR_DIR}/${name}"
  mkdir -p "${dest_dir}"
  out_ts="${dest_dir}/handoff-$(date '+%Y%m%dT%H%M%S%z').json"
  out_latest="${dest_dir}/latest.json"
  tmp="$(mktemp "${dest_dir}/.handoff.tmp.XXXXXX")"

  local ok=0
  if [[ "${USE_CURL}" -eq 0 ]]; then
    local ymbot_path
    if ymbot_path="$(resolve_ymbot)"; then
      # env -u：避免继承宿主 YMBOT_MEMBER_ID 串号；仅 GET export，无写
      local errf="${tmp}.err"
      if env -u YMBOT_MEMBER_ID \
        YMBOT_API_URL="${YMBOT_API_URL}" \
        YMBOT_API_KEY="${key}" \
        YMBOT_WORKSPACE_ID="${YMBOT_WORKSPACE_ID}" \
        "${ymbot_path}" handoff export --out "${tmp}" >/dev/null 2>"${errf}"
      then
        ok=1
        rm -f "${errf}"
      else
        log "WARN ${name}: ymbot export 失败，尝试 curl REST"
        if [[ -s "${errf}" ]]; then
          log "ymbot[${name}]: $(tr '\n' ' ' <"${errf}" | head -c 200)"
        fi
        rm -f "${errf}" "${tmp}"
        : >"${tmp}"
      fi
    else
      log "WARN: 未找到 ymbot，改用 curl REST"
    fi
  fi

  if [[ "${ok}" -eq 0 ]]; then
    # REST：先 me 拿 member_id，再 GET handoff（纯读）
    local base me_json mid http_code
    base="${YMBOT_API_URL%/}/api/v1/workspaces/${YMBOT_WORKSPACE_ID}"
    me_json="$(mktemp)"
    http_code="$(curl -s -o "${me_json}" -w '%{http_code}' \
      -H "Authorization: Bearer ${key}" \
      -H "Accept: application/json" \
      "${base}/bots/me" 2>/dev/null || true)"
    [[ -n "${http_code}" ]] || http_code="000"
    if [[ "${http_code}" != "200" ]]; then
      log "FAIL ${name}: GET /bots/me HTTP ${http_code}"
      rm -f "${me_json}" "${tmp}"
      return 1
    fi
    mid="$(extract_json_str "${me_json}" member_id)"
    rm -f "${me_json}"
    if [[ -z "${mid}" ]]; then
      log "FAIL ${name}: /bots/me 无 member_id"
      rm -f "${tmp}"
      return 1
    fi
    http_code="$(curl -s -o "${tmp}" -w '%{http_code}' \
      -H "Authorization: Bearer ${key}" \
      -H "Accept: application/json" \
      "${base}/bots/${mid}/handoff" 2>/dev/null || true)"
    [[ -n "${http_code}" ]] || http_code="000"
    if [[ "${http_code}" != "200" ]]; then
      log "FAIL ${name}: GET /handoff HTTP ${http_code}"
      rm -f "${tmp}"
      return 1
    fi
    # 美化缩进（失败则保留原文）
    if command -v python3 >/dev/null 2>&1; then
      python3 -c '
import json,sys
p=sys.argv[1]
with open(p) as f: d=json.load(f)
with open(p,"w") as f: json.dump(d,f,ensure_ascii=False,indent=2); f.write("\n")
' "${tmp}" || true
    fi
    ok=1
  fi

  if [[ "${ok}" -ne 1 ]]; then
    rm -f "${tmp}"
    return 1
  fi

  # 拒绝把 Key 误写入快照（粗检）
  if grep -qF "${key}" "${tmp}" 2>/dev/null; then
    log "FAIL ${name}: 快照疑似含 API Key，丢弃不落盘"
    rm -f "${tmp}"
    return 1
  fi

  chmod 600 "${tmp}"
  mv -f "${tmp}" "${out_ts}"
  # latest：优先软链，跨文件系统则拷贝
  rm -f "${out_latest}"
  if ln -s "$(basename "${out_ts}")" "${out_latest}" 2>/dev/null; then
    :
  else
    cp -f "${out_ts}" "${out_latest}"
    chmod 600 "${out_latest}"
  fi

  prune_old "${dest_dir}"
  log "OK   ${name}: ${out_ts}"
  return 0
}

prune_old() {
  local dir="$1"
  [[ "${KEEP_N}" =~ ^[0-9]+$ ]] || return 0
  [[ "${KEEP_N}" -eq 0 ]] && return 0
  # 只删 handoff-*.json，保留 latest.json
  local files
  mapfile -t files < <(ls -1t "${dir}"/handoff-*.json 2>/dev/null || true)
  local i=0
  for f in "${files[@]:-}"; do
    [[ -z "$f" ]] && continue
    i=$((i + 1))
    if [[ "$i" -gt "${KEEP_N}" ]]; then
      rm -f "$f"
      log "prune $(basename "$f")"
    fi
  done
}

run_round() {
  local i fail=0
  for i in "${!BOT_NAMES[@]}"; do
    if ! mirror_one "${BOT_NAMES[$i]}" "${BOT_KEYS[$i]}"; then
      fail=1
    fi
  done
  return "${fail}"
}

# ---------- parse args ----------
while [[ $# -gt 0 ]]; do
  case "$1" in
    --once) MODE="once"; shift ;;
    --loop) MODE="loop"; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    --curl) USE_CURL=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "未知参数: $1（试 --help）" ;;
  esac
done

require_config

if [[ "${DRY_RUN}" -eq 1 ]]; then
  log "dry-run: API_URL=${YMBOT_API_URL} WORKSPACE=${YMBOT_WORKSPACE_ID} MIRROR_DIR=${MIRROR_DIR} KEEP_N=${KEEP_N} bots=${#BOT_NAMES[@]}"
  local_i=0
  for local_i in "${!BOT_NAMES[@]}"; do
    log "dry-run: would mirror bot=${BOT_NAMES[$local_i]} key=ymk_…(redacted)"
  done
  if resolve_ymbot >/dev/null 2>&1; then
    log "dry-run: ymbot=$(resolve_ymbot)"
  else
    log "dry-run: ymbot=missing → will use curl"
  fi
  exit 0
fi

mkdir -p "${MIRROR_DIR}"
chmod 700 "${MIRROR_DIR}" 2>/dev/null || true

round_fail=0
if [[ "${MODE}" == "loop" ]]; then
  log "loop start interval=${INTERVAL_SEC}s bots=${#BOT_NAMES[@]}"
  while true; do
    if ! run_round; then
      round_fail=2
    fi
    sleep "${INTERVAL_SEC}"
  done
else
  if ! run_round; then
    round_fail=2
  fi
fi

exit "${round_fail}"
