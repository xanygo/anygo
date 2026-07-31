package file

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"slices"
	"unsafe"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/ds/xcontainer"
	"github.com/xanygo/anygo/safely"
)

type ZSet struct {
	Compact func()
	Base    *Base
}

func (zs *ZSet) saveMember(member string, score float64) error {
	m := memberScore{
		Member: unsafe.Slice(unsafe.StringData(member), len(member)),
		Score:  score,
	}
	bf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return zs.Base.writeMemberFile(zs.Base.md5(member), string(bf))
}

func (zs *ZSet) deleteMember(member string) error {
	return zs.Base.deleteMemberFile(zs.Base.md5(member))
}

func (zs *ZSet) memberScore(member string) (float64, bool, error) {
	str, found, err := zs.Base.readMemberFile(zs.Base.md5(member))
	if err != nil || !found {
		return 0, false, err
	}
	m := &memberScore{}
	bf := unsafe.Slice(unsafe.StringData(str), len(str))
	err = json.Unmarshal(bf, m)
	return m.Score, err == nil, err
}

func (zs *ZSet) ZAdd(ctx context.Context, score float64, member string) error {
	return zs.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		return zs.saveMember(member, score)
	})
}

func (zs *ZSet) ZIncrBy(ctx context.Context, score float64, member string) (result float64, err error) {
	err = zs.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		value, hasOld, err1 := zs.memberScore(member)
		if err1 != nil {
			return err1
		}
		result = value + score
		err2 := zs.saveMember(member, result)
		if !hasOld && err2 != nil {
			zs.Base.deleteKeyWhenNoMember(ctx)
		}
		return err2
	})
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (zs *ZSet) ZScore(ctx context.Context, member string) (score float64, found bool, err error) {
	err = zs.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}

		value, ok, err1 := zs.memberScore(member)
		if err1 != nil || !ok {
			return err1
		}
		score = value
		found = ok
		return nil
	})
	return score, found, err
}

func (zs *ZSet) rangeMembers(ctx context.Context, fn func(item *memberScore) (bool, error)) error {
	return zs.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
		bf, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		m := &memberScore{}
		err = json.Unmarshal(bf, m)
		if err != nil {
			return err
		}
		ok, err1 := fn(m)
		if err1 != nil {
			return err1
		}
		if !ok {
			return fs.SkipAll
		}
		return nil
	})
}

var zsMemberSortFn = xcmp.OrderAsc(func(t *memberScore) float64 {
	return t.Score
})

func (zs *ZSet) ZRange(ctx context.Context, fn func(member string, score float64) bool) error {
	return zs.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		var list []*memberScore
		err := zs.rangeMembers(ctx, func(item *memberScore) (bool, error) {
			list = append(list, item)
			return true, nil
		})
		if err != nil {
			return err
		}
		slices.SortFunc(list, zsMemberSortFn)
		for _, m := range list {
			if !fn(m.MemberString(), m.Score) {
				return nil
			}
		}
		return nil
	})
}

func (zs *ZSet) ZRem(ctx context.Context, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	return zs.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		defer func() {
			zs.Base.deleteKeyWhenNoMember(ctx)
			go safely.RunVoid(zs.Compact)
		}()
		var errs []error
		for _, member := range members {
			if err := zs.deleteMember(member); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	})
}

func (zs *ZSet) ZCount(ctx context.Context, min, max string) (num int64, err error) {
	minBound := &xcmp.Bound[float64]{}
	if err = minBound.ParserMin(min); err != nil {
		return 0, err
	}

	maxBound := &xcmp.Bound[float64]{}
	if err = maxBound.ParserMax(max); err != nil {
		return 0, err
	}
	err = zs.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return zs.rangeMembers(ctx, func(item *memberScore) (bool, error) {
			match := minBound.MatchMin(item.Score) && maxBound.MatchMax(item.Score)
			if match {
				num++
			}
			return true, nil
		})
	})
	return num, err
}

func (zs *ZSet) ZLen(ctx context.Context) (num int64, err error) {
	err = zs.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return zs.rangeMembers(ctx, func(item *memberScore) (bool, error) {
			num++
			return true, nil
		})
	})
	return num, err
}

func (zs *ZSet) ZRank(ctx context.Context, member string) (index int64, score float64, err error) {
	index = -1
	var idx int64
	err = zs.ZRange(ctx, func(name string, sc float64) bool {
		if member == name {
			index = idx
			score = sc
			return false
		}
		idx++
		return true
	})
	return index, score, err
}

func (zs *ZSet) popXX(ctx context.Context, count int, sort func(a, b *memberScore) bool) (members []string, scores []float64, err error) {
	if count < 1 {
		return nil, nil, nil
	}

	err = zs.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		defer zs.Base.deleteKeyWhenNoMember(ctx)

		xh := xcontainer.NewTopNHeap[*memberScore](count, sort)
		err1 := zs.rangeMembers(ctx, func(item *memberScore) (bool, error) {
			xh.Add(item)
			return true, nil
		})
		if err1 != nil {
			return err1
		}
		for _, item := range xh.Sorted() {
			member := item.MemberString()
			members = append(members, member)
			scores = append(scores, item.Score)

			// 不能保证原子性
			if err2 := zs.deleteMember(member); err2 != nil {
				return err2
			}
		}
		return nil
	})

	return members, scores, err
}

func memberScoreSortDesc(a, b *memberScore) bool {
	return a.Score < b.Score
}

func (zs *ZSet) ZPopMax(ctx context.Context, count int) (members []string, scores []float64, err error) {
	return zs.popXX(ctx, count, memberScoreSortDesc)
}

func memberScoreSortAsc(a, b *memberScore) bool {
	return a.Score > b.Score
}

func (zs *ZSet) ZPopMin(ctx context.Context, count int) (members []string, scores []float64, err error) {
	return zs.popXX(ctx, count, memberScoreSortAsc)
}

type memberScore struct {
	Member []byte  `json:"m"`
	Score  float64 `json:"s"`
}

func (fm memberScore) MemberString() string {
	return unsafe.String(unsafe.SliceData(fm.Member), len(fm.Member))
}
