package mem

import (
	"slices"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/ds/xslice"
)

type ZSetValue struct {
	Members []string           `json:"m"` // 按照 score 升序排序的 members 集合
	Scores  map[string]float64 `json:"s"`
}

type memberScore struct {
	member string
	score  float64
}

func (mz *ZSetValue) Add(score float64, member string) {
	if mz.Scores == nil {
		mz.Scores = make(map[string]float64)
	}
	mz.Scores[member] = score
	mz.sort()
}

// 排序：先按照分数升序，若分数相同，则按照 member 升序
var memberScoreSortFn = xcmp.Chain(
	xcmp.OrderAsc(func(t memberScore) float64 {
		return t.score
	}),
	xcmp.OrderAsc(func(t memberScore) string {
		return t.member
	}),
)

func (mz *ZSetValue) sort() {
	list := make([]memberScore, 0, len(mz.Scores))
	for k, v := range mz.Scores {
		list = append(list, memberScore{
			member: k,
			score:  v,
		})
	}
	slices.SortFunc(list, memberScoreSortFn)
	mz.Members = make([]string, 0, len(mz.Scores))
	for _, item := range list {
		mz.Members = append(mz.Members, item.member)
	}
}

func (mz *ZSetValue) IncrBy(member string, add float64) float64 {
	if mz.Scores == nil {
		mz.Scores = make(map[string]float64)
	}
	mz.Scores[member] += add
	mz.sort()
	return mz.Scores[member]
}

func (mz *ZSetValue) Score(member string) (float64, bool) {
	if len(mz.Scores) == 0 {
		return 0, false
	}
	score, found := mz.Scores[member]
	return score, found
}

func (mz *ZSetValue) Remove(member string) bool {
	if len(mz.Members) == 0 {
		return false
	}
	_, found := mz.Scores[member]
	if !found {
		return false
	}
	delete(mz.Scores, member)
	mz.Members = xslice.DeleteValue(mz.Members, member)
	return true
}

func (mz *ZSetValue) Count(min, max *xcmp.Bound[float64]) (num int64) {
	if len(mz.Scores) == 0 {
		return 0
	}
	for _, s := range mz.Scores {
		if min.MatchMin(s) && max.MatchMax(s) {
			num++
		}
	}
	return num
}

func (mz *ZSetValue) Rank(member string) (int64, float64) {
	if len(mz.Scores) == 0 {
		return -1, 0
	}
	score, found := mz.Scores[member]
	if !found {
		return -1, 0
	}
	index := slices.Index(mz.Members, member)
	return int64(index), score
}
