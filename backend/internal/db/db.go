package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql schema_postgres.sql
var schemaFS embed.FS

//go:embed postgres/*.sql
var postgresMigrateFS embed.FS

var registerOnce sync.Once

// Open opens sqlite or postgres/pgx.
// SQLite: always applies embedded schema.sql.
// Postgres: Ping + pool; applies postgres/*.sql (fallback schema_postgres.sql) only when AUTO_MIGRATE=1.
func Open(driver, dsn string) (*sql.DB, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "sqlite", "":
		return openSQLite(dsn)
	case "postgres", "postgresql", "pgx":
		return openPostgres(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", driver)
	}
}

func openSQLite(dsn string) (*sql.DB, error) {
	if filepath.Dir(dsn) != "." && filepath.Dir(dsn) != "" {
		_ = os.MkdirAll(filepath.Dir(dsn), 0o755)
	} else {
		_ = os.MkdirAll("./data", 0o755)
	}
	db, err := sql.Open("sqlite", dsn+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := migrateEmbedded(db, "schema.sql"); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func openPostgres(dsn string) (*sql.DB, error) {
	registerOnce.Do(func() {
		sql.Register("postgres-rebind", rebindDriver{d: stdlib.GetDefaultDriver()})
	})
	db, err := sql.Open("postgres-rebind", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	if shouldAutoMigratePostgres() {
		// Prefer numbered postgres/*.sql if present; else schema_postgres.sql monolith.
		if err := migratePostgresDir(db); err != nil {
			// Fall back to monolith (contains extension + full schema + embeddings)
			if err2 := migrateEmbedded(db, "schema_postgres.sql"); err2 != nil {
				_ = db.Close()
				combined := fmt.Errorf("postgres migrate dir: %v; monolith: %w", err, err2)
				if hint := migrateHint(combined); hint != "" && !strings.Contains(combined.Error(), "pg-bootstrap.sh") {
					combined = fmt.Errorf("%w%s", combined, hint)
				}
				return nil, combined
			}
		}
	}
	return db, nil
}

func shouldAutoMigratePostgres() bool {
	v := strings.TrimSpace(os.Getenv("AUTO_MIGRATE"))
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}

func migrateEmbedded(db *sql.DB, name string) error {
	raw, err := schemaFS.ReadFile(name)
	if err != nil {
		return err
	}
	return execSQLStatements(db, string(raw))
}

func migratePostgresDir(db *sql.DB) error {
	entries, err := postgresMigrateFS.ReadDir("postgres")
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return fmt.Errorf("no postgres/*.sql")
	}
	sort.Strings(names)
	for _, name := range names {
		raw, err := postgresMigrateFS.ReadFile("postgres/" + name)
		if err != nil {
			return err
		}
		if err := execSQLStatements(db, string(raw)); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

func execSQLStatements(db *sql.DB, raw string) error {
	for _, s := range splitSQL(raw) {
		s = strings.TrimSpace(s)
		if s == "" || strings.HasPrefix(s, "--") {
			continue
		}
		if _, err := db.Exec(s); err != nil {
			if toleratesMigrateError(db, s, err) {
				continue
			}
			return fmt.Errorf("migrate: %w\nstmt: %s%s", err, truncate(s, 160), migrateHint(err))
		}
	}
	return nil
}

func toleratesMigrateError(db *sql.DB, stmt string, err error) bool {
	msg := strings.ToLower(err.Error())
	// SQLite: CREATE TABLE already has start_at/due_at; trailing ALTER ADD COLUMN is for old DBs only.
	if strings.Contains(msg, "duplicate column name") {
		return true
	}
	if !strings.Contains(msg, "42501") && !strings.Contains(msg, "permission denied") {
		return false
	}
	up := strings.ToUpper(stmt)
	if strings.Contains(up, "CREATE EXTENSION") {
		// Superuser already installed via scripts/pg-bootstrap.sh — app role cannot CREATE EXTENSION.
		name := extensionNameFromStmt(stmt)
		if name == "" {
			return false
		}
		var one int
		qerr := db.QueryRow(`SELECT 1 FROM pg_extension WHERE extname = $1`, name).Scan(&one)
		return qerr == nil
	}
	return false
}

func extensionNameFromStmt(stmt string) string {
	// CREATE EXTENSION IF NOT EXISTS vector;
	up := strings.ToUpper(stmt)
	i := strings.Index(up, "EXTENSION")
	if i < 0 {
		return ""
	}
	rest := strings.TrimSpace(stmt[i+len("EXTENSION"):])
	rest = strings.TrimPrefix(rest, "IF NOT EXISTS")
	rest = strings.TrimPrefix(rest, "if not exists")
	rest = strings.TrimSpace(rest)
	rest = strings.TrimSuffix(rest, ";")
	rest = strings.TrimSpace(rest)
	rest = strings.Trim(rest, `"'`)
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToLower(fields[0])
}

func migrateHint(err error) string {
	if needsPostgresBootstrap(err) {
		return "\n提示: 干净库请先以超级用户执行 ./scripts/pg-bootstrap.sh（GRANT public + CREATE EXTENSION vector），再 AUTO_MIGRATE=1"
	}
	return ""
}

// needsPostgresBootstrap detects empty-DB privilege/search_path failures that bootstrap fixes.
func needsPostgresBootstrap(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	needles := []string{
		"42501",
		"3f000",
		"permission denied",
		"no schema has been selected",
		"must be owner",
		"create extension",
	}
	for _, n := range needles {
		if strings.Contains(msg, n) {
			// Avoid false positives on unrelated "create extension" in statements alone:
			if n == "create extension" && !(strings.Contains(msg, "42501") || strings.Contains(msg, "3f000") || strings.Contains(msg, "permission") || strings.Contains(msg, "no schema")) {
				continue
			}
			return true
		}
	}
	return false
}

func splitSQL(s string) []string {
	var out []string
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "--") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
		if strings.HasSuffix(trim, ";") {
			out = append(out, b.String())
			b.Reset()
		}
	}
	if t := strings.TrimSpace(b.String()); t != "" {
		out = append(out, t)
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// IsPostgres reports whether driver is a Postgres variant.
func IsPostgres(driver string) bool {
	d := strings.ToLower(driver)
	return d == "postgres" || d == "postgresql" || d == "pgx"
}
