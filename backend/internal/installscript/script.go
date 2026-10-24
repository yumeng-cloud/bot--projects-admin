// Package installscript embeds the ymbot one-shot installer served at GET /install.sh.
package installscript

import _ "embed"

// Script is the body of install.sh (POSIX bash). Kept in sync with scripts/install-ymbot.sh.
//
//go:embed install.sh
var Script string

// GoInstallCmd is the documented fallback (module path matches go.mod).
const GoInstallCmd = "go install github.com/yumeng-tech/ai-collab-board/backend/cmd/ymbot@latest"

// ContentType for GET /install.sh.
const ContentType = "text/plain; charset=utf-8"
