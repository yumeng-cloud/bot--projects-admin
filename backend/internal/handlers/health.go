package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
)

type HealthHandler struct {
	DB *sql.DB
}

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	dbOK := true
	if err := h.DB.Ping(); err != nil {
		status = "degraded"
		dbOK = false
	}
	code := http.StatusOK
	if !dbOK {
		code = http.StatusServiceUnavailable
	}
	middleware.JSON(w, code, map[string]any{
		"status":    status,
		"db":        dbOK,
		"time":      time.Now().UTC().Format(time.RFC3339Nano),
		"service":   "ai-collab-board",
		"timezone":  "Asia/Shanghai",
	})
}
