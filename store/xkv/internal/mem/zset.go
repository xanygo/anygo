package mem

import (
	"context"
	"slices"
	"sync"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xkv/internal"
)

type zsetValue struct {
	Members []string           `json:"m"` // 按照 score 升序排序的 members 集合
	Scores  map[string]float64 `json:"s"`
	mux     sync.RWMutex
}

type memberScore struct {
	member string
	score  float64
}

func (mz *zsetValue) Add(score float64, member string) {
	mz.mux.Lock()
	defer mz.mux.Unlock()

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

func (mz *zsetValue) sort() {
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

func (mz *zsetValue) IncrBy(member string, add float64) float64 {
	mz.mux.Lock()
	defer mz.mux.Unlock()

	if mz.Scores == nil {
		mz.Scores = make(map[string]float64)
	}
	mz.Scores[member] += add
	mz.sort()
	return mz.Scores[member]
}

func (mz *zsetValue) Score(member string) (float64, bool) {
	mz.mux.RLock()
	defer mz.mux.RUnlock()

	if len(mz.Scores) == 0 {
		return 0, false
	}
	score, found := mz.Scores[member]
	return score, found
}

func (mz *zsetValue) Remove(member string) bool {
	mz.mux.Lock()
	defer mz.mux.Unlock()

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

func (mz *zsetValue) Count(min, max *xcmp.Bound[float64]) (num int64) {
	mz.mux.RLock()
	defer mz.mux.RUnlock()

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

func (mz *zsetValue) Rank(member string) (int64, float64) {
	mz.mux.RLock()
	defer mz.mux.RUnlock()

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

func (mz *zsetValue) Len() int64 {
	mz.mux.RLock()
	defer mz.mux.RUnlock()
	return int64(len(mz.Members))
}

func (mz *zsetValue) PopMax(count int) (members []string, scores []float64) {
	mz.mux.Lock()
	defer mz.mux.Unlock()
	mz.Members, members = xslice.PopTailN(mz.Members, count)
	for _, m := range members {
		scores = append(scores, mz.Scores[m])
		delete(mz.Scores, m)
	}
	return members, scores
}

func (mz *zsetValue) PopMin(count int) (members []string, scores []float64) {
	mz.mux.Lock()
	defer mz.mux.Unlock()
	mz.Members, members = xslice.PopHeadN(mz.Members, count)
	for _, m := range members {
		scores = append(scores, mz.Scores[m])
		delete(mz.Scores, m)
	}
	return members, scores
}

func zsetValueEmpty(mz *zsetValue) bool {
	return mz == nil || mz.Len() == 0
}

type ZSet struct {
	Base *Base
	Key  string
}

func (m *ZSet) withLocked(fn func(*zsetValue) (*zsetValue, operate, error)) error {
	return withLocked[*zsetValue](m.Base, m.Key, internal.DataTypeZSet, func(value *zsetValue) (*zsetValue, operate, error) {
		if value == nil {
			value = &zsetValue{}
		}
		return fn(value)
	}, zsetValueEmpty)
}

func (m *ZSet) ZAdd(ctx context.Context, score float64, member string) error {
	return m.withLocked(func(zv *zsetValue) (*zsetValue, operate, error) {
		zv.Add(score, member)
		return zv, opWrite, nil
	})
}

func (m *ZSet) ZScore(ctx context.Context, member string) (float64, bool, error) {
	var score float64
	var found bool
	err := m.withLocked(func(zv *zsetValue) (*zsetValue, operate, error) {
		score, found = zv.Score(member)
		return zv, opSkip, nil
	})
	return score, found, err
}

func (m *ZSet) ZIncrBy(ctx context.Context, incr float64, member string) (float64, error) {
	var score float64
	err := m.withLocked(func(zv *zsetValue) (*zsetValue, operate, error) {
		score = zv.IncrBy(member, incr)
		return zv, opWrite, nil
	})
	return score, err
}

func (m *ZSet) ZCount(ctx context.Context, min, max string) (num int64, err error) {
	minBound := &xcmp.Bound[float64]{}
	if err = minBound.ParserMin(min); err != nil {
		return 0, err
	}

	maxBound := &xcmp.Bound[float64]{}
	if err = maxBound.ParserMax(max); err != nil {
		return 0, err
	}
	err = m.withLocked(func(zv *zsetValue) (*zsetValue, operate, error) {
		num = zv.Count(minBound, maxBound)
		return zv, opSkip, nil
	})
	return num, err
}

func (m *ZSet) ZLen(ctx context.Context) (num int64, err error) {
	return m.ZCount(ctx, "-inf", "+inf")
}

func (m *ZSet) ZRange(ctx context.Context, fn func(member string, score float64) bool) error {
	var value *zsetValue
	err := m.withLocked(func(zv *zsetValue) (*zsetValue, operate, error) {
		value = zv
		return zv, opSkip, nil
	})
	if err != nil {
		return err
	}
	if value != nil {
		for _, member := range value.Members {
			if !fn(member, value.Scores[member]) {
				return nil
			}
		}
	}
	return nil
}

func (m *ZSet) ZRank(ctx context.Context, member string) (index int64, score float64, err error) {
	err = m.withLocked(func(zv *zsetValue) (*zsetValue, operate, error) {
		index, score = zv.Rank(member)
		return zv, opSkip, nil
	})
	return index, score, err
}

func (m *ZSet) ZRem(ctx context.Context, members ...string) error {
	return m.withLocked(func(value *zsetValue) (*zsetValue, operate, error) {
		var op operate
		for _, member := range members {
			if value.Remove(member) {
				op = opWrite
			}
		}
		return value, op, nil
	})
}

func (m *ZSet) ZPopMax(ctx context.Context, count int) (members []string, scores []float64, err error) {
	err = m.withLocked(func(value *zsetValue) (*zsetValue, operate, error) {
		members, scores = value.PopMax(count)
		return value, opWrite, nil
	})
	return members, scores, err
}

func (m *ZSet) ZPopMin(ctx context.Context, count int) (members []string, scores []float64, err error) {
	err = m.withLocked(func(value *zsetValue) (*zsetValue, operate, error) {
		members, scores = value.PopMin(count)
		return value, opWrite, nil
	})
	return members, scores, err
}
