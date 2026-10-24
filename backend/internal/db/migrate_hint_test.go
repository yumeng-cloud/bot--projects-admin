package db

import (
	"errors"
	"strings"
	"testing"
)

func TestMigrateHintIncludesBootstrapPath(t *testing.T) {
	cases := []error{
		errors.New(`ERROR: no schema has been selected to create in (SQLSTATE 3F000)`),
		errors.New(`ERROR: permission denied for schema public (SQLSTATE 42501)`),
		errors.New(`ERROR: permission denied to create extension "vector" (SQLSTATE 42501)`),
	}
	for _, err := range cases {
		h := migrateHint(err)
		if !strings.Contains(h, "pg-bootstrap.sh") {
			t.Fatalf("hint missing for %v: %q", err, h)
		}
	}
}
