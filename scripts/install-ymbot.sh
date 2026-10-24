#!/usr/bin/env bash
# ymbot one-shot installer — served at GET /install.sh (PUBLIC_API_BASE).
# Idempotent, non-interactive. Failures print bilingual human messages.
set -euo pipefail

REPO_GIT="${YMBOT_REPO_GIT:-https://github.com/yumeng-cloud/bot--projects-admin.git}"
GO_MODULE_INSTALL="${YMBOT_GO_INSTALL:-github.com/yumeng-tech/ai-collab-board/backend/cmd/ymbot@latest}"
BIN_NAME="ymbot"

die() {
  echo "" >&2
  echo "ERROR / 错误: $*" >&2
  echo "Fallback / 回退: install Go (>=1.22), then:" >&2
  echo "  go install ${GO_MODULE_INSTALL}" >&2
  echo "Or clone and build:" >&2
  echo "  git clone --depth 1 ${REPO_GIT} && cd bot--projects-admin/backend && go install ./cmd/ymbot" >&2
  exit 1
}

info() { echo "→ $*"; }

os="$(uname -s 2>/dev/null || echo unknown)"
arch="$(uname -m 2>/dev/null || echo unknown)"
case "$os" in
  Linux|Darwin) ;;
  MINGW*|MSYS*|CYGWIN*)
    die "Windows shell detected. Use WSL/macOS/Linux, or: go install ${GO_MODULE_INSTALL}"
    ;;
  *)
    die "Unsupported OS '${os}'. Use Linux/macOS, or: go install ${GO_MODULE_INSTALL}"
    ;;
esac
case "$arch" in
  x86_64|amd64|arm64|aarch64) ;;
  *)
    info "Unusual arch '${arch}'; continuing (go install will decide)."
    ;;
esac

# Prefer user-writable PATH dirs (no sudo).
if [[ -n "${YMBOT_INSTALL_DIR:-}" ]]; then
  DEST_DIR="${YMBOT_INSTALL_DIR}"
elif [[ -d "${HOME}/.local/bin" ]] || mkdir -p "${HOME}/.local/bin" 2>/dev/null; then
  DEST_DIR="${HOME}/.local/bin"
elif [[ -w /usr/local/bin ]]; then
  DEST_DIR="/usr/local/bin"
else
  DEST_DIR="${HOME}/.local/bin"
  mkdir -p "${DEST_DIR}" || die "Cannot create ${DEST_DIR}"
fi

DEST="${DEST_DIR}/${BIN_NAME}"
info "OS=${os} arch=${arch}"
info "Install target: ${DEST}"

need_path_hint=0
case ":${PATH}:" in
  *":${DEST_DIR}:"*) ;;
  *) need_path_hint=1 ;;
esac

have_go=0
if command -v go >/dev/null 2>&1; then
  have_go=1
  info "Found $(go version 2>/dev/null || echo go)"
fi

build_via_go_install() {
  GOBIN="${DEST_DIR}" go install "${GO_MODULE_INSTALL}"
}

build_via_clone() {
  local tmp
  tmp="$(mktemp -d 2>/dev/null || mktemp -d -t ymbot-install)"
  # shellcheck disable=SC2064
  trap "rm -rf '${tmp}'" EXIT
  info "Cloning ${REPO_GIT} (shallow)…"
  git clone --depth 1 "${REPO_GIT}" "${tmp}/repo" || die "git clone failed (need network + git)"
  (
    cd "${tmp}/repo/backend" || die "repo missing backend/"
    GOBIN="${DEST_DIR}" go install ./cmd/ymbot
  ) || die "go install ./cmd/ymbot failed"
  trap - EXIT
  rm -rf "${tmp}"
}

if [[ "${have_go}" -eq 1 ]]; then
  info "Trying: go install ${GO_MODULE_INSTALL}"
  if ! build_via_go_install 2>/tmp/ymbot-go-install.err; then
    info "Module path install failed; falling back to clone + go install ./cmd/ymbot"
    command -v git >/dev/null 2>&1 || die "git required for fallback (see /tmp/ymbot-go-install.err)"
    build_via_clone
  fi
else
  die "Go toolchain not found in PATH. Install Go, then re-run, or: go install ${GO_MODULE_INSTALL}"
fi

if [[ ! -x "${DEST}" ]]; then
  # go install may place binary under GOPATH/bin when GOBIN ignored
  gopath_bin="$(go env GOPATH 2>/dev/null)/bin/${BIN_NAME}"
  if [[ -x "${gopath_bin}" ]]; then
    mkdir -p "${DEST_DIR}"
    cp -f "${gopath_bin}" "${DEST}" || die "Cannot copy ${gopath_bin} → ${DEST}"
    chmod +x "${DEST}" || true
  else
    die "ymbot binary not found at ${DEST} after install"
  fi
fi

echo ""
echo "OK / 完成: installed ${BIN_NAME} → ${DEST}"
if [[ "${need_path_hint}" -eq 1 ]]; then
  echo ""
  echo "Add to PATH / 请加入 PATH:"
  echo "  export PATH=\"${DEST_DIR}:\$PATH\""
  echo "  # e.g. append to ~/.bashrc or ~/.zshrc"
fi
echo ""
echo "Next / 下一步:"
echo "  ymbot -h"
echo "  # then: ymbot team join --code ymt_… [--role …]"
echo ""
