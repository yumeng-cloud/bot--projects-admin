package issueboard

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

type Service struct {
	DB       *sql.DB
	postgres bool // SELECT … FOR UPDATE when true
}

func New(sqlDB *sql.DB) *Service {
	s := &Service{DB: sqlDB}
	// Detect live dialect (ignore env leftovers): Postgres accepts SHOW server_version.
	var v string
	if err := sqlDB.QueryRow(`SHOW server_version`).Scan(&v); err == nil {
		s.postgres = true
	}
	return s
}

// querier is satisfied by *sql.DB and *sql.Tx so cycle checks can run inside a transaction.
type querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func nowUTC() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func nullStr(s *string) any {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}

func nullInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func strPtr(ns sql.NullString) *string {
	if ns.Valid {
		s := ns.String
		return &s
	}
	return nil
}

func intPtr(ni sql.NullInt64) *int {
	if ni.Valid {
		v := int(ni.Int64)
		return &v
	}
	return nil
}

var validStatuses = map[string]bool{
	"backlog": true, "todo": true, "in_progress": true,
	"in_review": true, "done": true, "canceled": true,
}

var validPriorities = map[string]bool{
	"urgent": true, "high": true, "medium": true, "low": true, "none": true,
}

var validCycleStatuses = map[string]bool{
	"planned": true, "active": true, "completed": true,
}

var validRelationTypes = map[string]bool{
	"blocks": true, "blocked_by": true, "related": true, "parent": true,
}

func appendActivity(tx *sql.Tx, workspaceID, issueID, actorID, actorKind, action string, payload any) error {
	b, _ := json.Marshal(payload)
	if b == nil {
		b = []byte("{}")
	}
	var issue any
	if issueID != "" {
		issue = issueID
	}
	var actor any
	if actorID != "" {
		actor = actorID
	}
	if actorKind == "" {
		actorKind = "human"
	}
	_, err := tx.Exec(`INSERT INTO activities
		(id, workspace_id, issue_id, actor_id, actor_kind, action, payload, created_at)
		VALUES (?,?,?,?,?,?,?,?)`,
		idgen.New(), workspaceID, issue, actor, actorKind, action, string(b), nowUTC())
	return err
}

func normalizePrefix(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
