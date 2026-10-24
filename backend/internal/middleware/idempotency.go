package middleware

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

// Idempotency stores successful write responses keyed by Idempotency-Key + workspace.
func Idempotency(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
			if key == "" || (r.Method != http.MethodPost && r.Method != http.MethodPatch && r.Method != http.MethodPut && r.Method != http.MethodDelete) {
				next.ServeHTTP(w, r)
				return
			}
			wsID := WorkspaceID(r.Context())
			if wsID == "" {
				next.ServeHTTP(w, r)
				return
			}
			var status int
			var body string
			err := db.QueryRow(`
				SELECT response_status, response_body FROM idempotency_keys
				WHERE workspace_id = ? AND key_value = ?`, wsID, key).Scan(&status, &body)
			if err == nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Header().Set("X-Idempotent-Replay", "true")
				w.WriteHeader(status)
				_, _ = w.Write([]byte(body))
				return
			}

			rec := &statusRecorder{ResponseWriter: w, code: 200, buf: &bytes.Buffer{}}
			next.ServeHTTP(rec, r)
			if rec.code >= 200 && rec.code < 300 {
				_, _ = db.Exec(`
					INSERT INTO idempotency_keys
					(id, workspace_id, key_value, method, path, response_status, response_body)
					VALUES (?,?,?,?,?,?,?)
					ON CONFLICT(workspace_id, key_value) DO NOTHING`,
					idgen.New(), wsID, key, r.Method, r.URL.Path, rec.code, rec.buf.String())
			}
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	code int
	buf  *bytes.Buffer
}

func (s *statusRecorder) WriteHeader(code int) {
	s.code = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	_, _ = s.buf.Write(b)
	return s.ResponseWriter.Write(b)
}

var _ io.Writer = (*statusRecorder)(nil)
