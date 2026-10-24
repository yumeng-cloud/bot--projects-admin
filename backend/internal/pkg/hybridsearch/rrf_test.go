package hybridsearch_test

import (
	"reflect"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/hybridsearch"
)

func TestRRF(t *testing.T) {
	a := []string{"p1", "p2", "p3"}
	b := []string{"p2", "p4", "p1"}
	got := hybridsearch.RRF(a, b)
	if got[0] != "p1" && got[0] != "p2" {
		t.Fatalf("top should be p1 or p2: %v", got)
	}
	// p2 appears rank1 in b and rank2 in a → often wins
	if !contains(got, "p4") {
		t.Fatalf("missing p4: %v", got)
	}
}

func TestRRFStableEmpty(t *testing.T) {
	if len(hybridsearch.RRF()) != 0 {
		t.Fatal("expected empty")
	}
	got := hybridsearch.RRF([]string{"a"}, nil)
	if !reflect.DeepEqual(got, []string{"a"}) {
		t.Fatalf("%v", got)
	}
}

func contains(ss []string, x string) bool {
	for _, s := range ss {
		if s == x {
			return true
		}
	}
	return false
}
