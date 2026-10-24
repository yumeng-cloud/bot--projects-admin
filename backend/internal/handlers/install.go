package handlers

import (
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/installscript"
)

// InstallHandler serves GET /install.sh — one-shot ymbot CLI installer (API root).
type InstallHandler struct{}

// Script handles GET /install.sh (also HEAD).
func (h *InstallHandler) Script(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "方法不允许 / Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	body := installscript.Script
	if body == "" {
		http.Error(w, "安装脚本未嵌入 / install script missing from binary", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", installscript.ContentType)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write([]byte(body))
}
