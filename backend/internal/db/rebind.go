package db

import (
	"strings"
)

// RebindPostgres converts SQLite-style ? placeholders to Postgres $1,$2,...
// Leaves existing $n alone; ignores ? inside single-quoted strings.
func RebindPostgres(q string) string {
	if !strings.ContainsRune(q, '?') {
		return q
	}
	var b strings.Builder
	b.Grow(len(q) + 8)
	n := 0
	inSingle := false
	for i := 0; i < len(q); i++ {
		c := q[i]
		if c == '\'' {
			if inSingle && i+1 < len(q) && q[i+1] == '\'' {
				b.WriteByte(c)
				b.WriteByte(q[i+1])
				i++
				continue
			}
			inSingle = !inSingle
			b.WriteByte(c)
			continue
		}
		if c == '?' && !inSingle {
			n++
			b.WriteByte('$')
			b.WriteString(itoa(n))
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// IsUniqueViolation reports SQLite UNIQUE or Postgres 23505 conflicts.
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "UNIQUE") ||
		strings.Contains(s, "unique constraint") ||
		strings.Contains(s, "SQLSTATE 23505") ||
		strings.Contains(s, "23505")
}
