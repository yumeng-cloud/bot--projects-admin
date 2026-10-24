package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInstallScript_OK(t *testing.T) {
	h := &InstallHandler{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/install.sh", nil)
	h.Script(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Fatalf("content-type %s", ct)
	}
	body := rr.Body.String()
	if strings.TrimSpace(body) == "" {
		t.Fatal("empty body")
	}
	for _, need := range []string{"ymbot", "go install", "#!/usr/bin/env bash"} {
		if !strings.Contains(body, need) {
			t.Fatalf("missing %q", need)
		}
	}
}
