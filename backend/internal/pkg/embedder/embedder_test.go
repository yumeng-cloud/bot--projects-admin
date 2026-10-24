package embedder_test

import (
	"context"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/embedder"
)

func TestHashEmbedDimsAndNorm(t *testing.T) {
	h := embedder.NewHash("test", 64)
	vecs, err := h.Embed(context.Background(), []string{"Issue 状态机", "状态流转规则"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 2 || len(vecs[0]) != 64 {
		t.Fatalf("dims: %d %d", len(vecs), len(vecs[0]))
	}
	// overlapping Chinese chars → positive cosine
	cos := cosine(vecs[0], vecs[1])
	if cos < 0.1 {
		t.Fatalf("expected some similarity, got %v", cos)
	}
	unrelated, _ := h.Embed(context.Background(), []string{"完全无关的香蕉食谱xyz"})
	cos2 := cosine(vecs[0], unrelated[0])
	if cos2 >= cos {
		t.Fatalf("unrelated should be less similar: related=%v unrelated=%v", cos, cos2)
	}
}

func TestChunk(t *testing.T) {
	parts := embedder.Chunk("abcdefghij", 4, 1)
	if len(parts) < 2 {
		t.Fatalf("chunks: %v", parts)
	}
}

func TestEncodeVector(t *testing.T) {
	s := embedder.EncodeVector([]float32{1, 0.5})
	if s != "[1.000000,0.500000]" {
		t.Fatalf("got %s", s)
	}
}

func cosine(a, b []float32) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (sqrt(na) * sqrt(nb))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 16; i++ {
		z = 0.5 * (z + x/z)
	}
	return z
}
