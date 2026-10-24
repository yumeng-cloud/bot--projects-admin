package apperr_test

import (
	"net/http"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

func TestVersionConflict(t *testing.T) {
	e := apperr.VersionConflict("乐观锁冲突")
	if e.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d want 409", e.StatusCode)
	}
	if e.Code != "version_conflict" {
		t.Fatalf("code: got %q want version_conflict", e.Code)
	}
	plain := apperr.Conflict("依赖成环")
	if plain.Code != "conflict" {
		t.Fatalf("Conflict code: got %q want conflict", plain.Code)
	}
}
