package db

import "testing"

func TestRebindPostgres(t *testing.T) {
	got := RebindPostgres(`SELECT a FROM t WHERE x = ? AND y = ?`)
	want := `SELECT a FROM t WHERE x = $1 AND y = $2`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	// already $n
	q := `SELECT a FROM t WHERE x = $1`
	if RebindPostgres(q) != q {
		t.Fatalf("should leave $n alone: %q", RebindPostgres(q))
	}
	// question mark inside string
	got = RebindPostgres(`SELECT '?' FROM t WHERE x = ?`)
	want = `SELECT '?' FROM t WHERE x = $1`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestIsUniqueViolation(t *testing.T) {
	if !IsUniqueViolation(errStr("UNIQUE constraint failed: users.email")) {
		t.Fatal("sqlite")
	}
	if !IsUniqueViolation(errStr(`ERROR: duplicate key value violates unique constraint "users_email_lower_uidx" (SQLSTATE 23505)`)) {
		t.Fatal("postgres")
	}
	if IsUniqueViolation(errStr("no such table")) {
		t.Fatal("false positive")
	}
}

type errStr string

func (e errStr) Error() string { return string(e) }
