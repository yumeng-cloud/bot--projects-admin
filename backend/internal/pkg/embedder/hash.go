package embedder

import (
	"context"
	"hash/fnv"
	"math"
	"unicode"
)

// Hash 本地确定性嵌入（字符/bigram 哈希投影），离线开发与单测用。
// 对中文共享字/词有一定召回；非生产语义模型。ModelID 带 hash: 前缀。
type Hash struct {
	model string
	dims  int
}

func NewHash(model string, dims int) *Hash {
	if dims <= 0 {
		dims = DefaultDims
	}
	if model == "" {
		model = "hash-local-zh"
	}
	if len(model) < 5 || model[:5] != "hash:" {
		model = "hash:" + model
	}
	return &Hash{model: model, dims: dims}
}

func (h *Hash) ModelID() string { return h.model }
func (h *Hash) Dims() int       { return h.dims }

func (h *Hash) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	_ = ctx
	out := make([][]float32, len(texts))
	for i, t := range texts {
		out[i] = hashEmbed(t, h.dims)
	}
	return out, nil
}

func hashEmbed(text string, dims int) []float32 {
	v := make([]float32, dims)
	runes := []rune(text)
	add := func(tok string, w float64) {
		fh := fnv.New64a()
		_, _ = fh.Write([]byte(tok))
		x := fh.Sum64()
		i := int(x % uint64(dims))
		sign := float32(1)
		if x&1 == 1 {
			sign = -1
		}
		v[i] += sign * float32(w)
		// 第二哈希减碰撞
		i2 := int((x >> 17) % uint64(dims))
		v[i2] += sign * float32(w) * 0.5
	}
	for _, r := range runes {
		if unicode.IsSpace(r) {
			continue
		}
		add(string(r), 1.0)
	}
	for i := 0; i+1 < len(runes); i++ {
		if unicode.IsSpace(runes[i]) || unicode.IsSpace(runes[i+1]) {
			continue
		}
		add(string(runes[i:i+2]), 1.5)
	}
	// L2 normalize
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if sum > 0 {
		inv := float32(1 / math.Sqrt(sum))
		for i := range v {
			v[i] *= inv
		}
	}
	return v
}
