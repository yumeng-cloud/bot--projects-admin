package auth

import (
	"testing"
	"time"
)

func TestIssueParse(t *testing.T) {
	tok, err := IssueToken("secret", "u1", "a@b.c", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ParseToken("secret", tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.UserID != "u1" || c.Email != "a@b.c" {
		t.Fatalf("claims: %+v", c)
	}
}
