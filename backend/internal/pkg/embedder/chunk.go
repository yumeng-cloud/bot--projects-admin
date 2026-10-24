package embedder

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Chunk 将文本切成重叠块，便于长文嵌入。按 rune 计长。
func Chunk(text string, maxRunes, overlap int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if maxRunes <= 0 {
		maxRunes = 800
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= maxRunes {
		overlap = maxRunes / 4
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return []string{string(runes)}
	}
	var out []string
	step := maxRunes - overlap
	for start := 0; start < len(runes); start += step {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[start:end]))
		if end == len(runes) {
			break
		}
	}
	return out
}

// EncodeVector 编码为 pgvector 文本字面量，如 [0.1,0.2]。
func EncodeVector(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.Grow(len(v) * 10)
	b.WriteByte('[')
	for i, x := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(x), 'f', 6, 32))
	}
	b.WriteByte(']')
	return b.String()
}

// RuneLen 返回 UTF-8 rune 数。
func RuneLen(s string) int { return utf8.RuneCountInString(s) }
