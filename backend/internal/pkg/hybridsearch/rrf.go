// Package hybridsearch 提供多路召回融合（RRF），供 Wiki / Bot 记忆共用。
package hybridsearch

import "sort"

// DefaultK RRF 常数，业界常用 60。
const DefaultK = 60

// RRF 融合多路有序 ID 列表（每路已按相关性从高到低）。
// 公式：score(d) = Σ 1/(k + rank_i(d))，rank 从 1 起。
func RRF(rankLists ...[]string) []string {
	return RRFK(DefaultK, rankLists...)
}

// RRFK 可调 k。
func RRFK(k int, rankLists ...[]string) []string {
	if k <= 0 {
		k = DefaultK
	}
	scores := map[string]float64{}
	for _, list := range rankLists {
		seen := map[string]bool{}
		for rank, id := range list {
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			scores[id] += 1.0 / float64(k+rank+1)
		}
	}
	type pair struct {
		id  string
		sc  float64
	}
	arr := make([]pair, 0, len(scores))
	for id, sc := range scores {
		arr = append(arr, pair{id, sc})
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].sc == arr[j].sc {
			return arr[i].id < arr[j].id
		}
		return arr[i].sc > arr[j].sc
	})
	out := make([]string, len(arr))
	for i, p := range arr {
		out[i] = p.id
	}
	return out
}

// RRFScored 同 RRF，但保留融合分。
func RRFScored(k int, rankLists ...[]string) []Scored {
	if k <= 0 {
		k = DefaultK
	}
	scores := map[string]float64{}
	for _, list := range rankLists {
		seen := map[string]bool{}
		for rank, id := range list {
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			scores[id] += 1.0 / float64(k+rank+1)
		}
	}
	out := make([]Scored, 0, len(scores))
	for id, sc := range scores {
		out = append(out, Scored{ID: id, Score: sc})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].ID < out[j].ID
		}
		return out[i].Score > out[j].Score
	})
	return out
}

// Scored 融合结果。
type Scored struct {
	ID    string
	Score float64
}
