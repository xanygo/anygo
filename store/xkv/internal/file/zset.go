package file

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"unsafe"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/ds/xcontainer"
	"github.com/xanygo/anygo/safely"
)

type ZSet struct {
	Compact func()
	Base    *Base
}

func (z *ZSet) saveMember(member string, score float64) error {
	m := memberScore{
		Member: unsafe.Slice(unsafe.StringData(member), len(member)),
		Score:  score,
	}
	bf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return z.Base.writeMemberFile(z.Base.md5(member), string(bf))
}

func (z *ZSet) deleteMember(member string) error {
	return z.Base.deleteMemberFile(z.Base.md5(member))
}

func (z *ZSet) memberScore(member string) (float64, bool, error) {
	str, found, err := z.Base.readMemberFile(z.Base.md5(member))
	if err != nil || !found {
		return 0, false, err
	}
	m := &memberScore{}
	bf := unsafe.Slice(unsafe.StringData(str), len(str))
	err = json.Unmarshal(bf, m)
	return m.Score, err == nil, err
}

func (z *ZSet) ZAdd(ctx context.Context, score float64, member string) error {
	return z.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		return z.saveMember(member, score)
	})
}

func (z *ZSet) ZIncrBy(ctx context.Context, score float64, member string) (result float64, err error) {
	err = z.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		value, hasOld, err1 := z.memberScore(member)
		if err1 != nil {
			return err1
		}
		result = value + score
		err2 := z.saveMember(member, result)
		if !hasOld && err2 != nil {
			z.Base.deleteKeyWhenNoMember(ctx)
		}
		return err2
	})
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (z *ZSet) ZScore(ctx context.Context, member string) (score float64, found bool, err error) {
	err = z.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}

		value, ok, err1 := z.memberScore(member)
		if err1 != nil || !ok {
			return err1
		}
		score = value
		found = ok
		return nil
	})
	return score, found, err
}

func (z *ZSet) rangeMembers(ctx context.Context, fn func(item *memberScore) (bool, error)) error {
	return z.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
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

func (z *ZSet) ZRange(ctx context.Context, fn func(member string, score float64) bool) error {
	return z.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return z.rangeMembers(ctx, func(item *memberScore) (bool, error) {
			if !fn(item.MemberString(), item.Score) {
				return false, nil
			}
			return true, nil
		})
	})
}

func (z *ZSet) ZRangeByScore(ctx context.Context, min string, max string, fn func(member string, score float64) bool) error {
	minBound := &xcmp.Bound[float64]{}
	if err := minBound.ParserMin(min); err != nil {
		return err
	}

	maxBound := &xcmp.Bound[float64]{}
	if err := maxBound.ParserMax(max); err != nil {
		return err
	}
	return z.ZRange(ctx, func(member string, score float64) bool {
		if minBound.MatchMin(score) && maxBound.MatchMax(score) {
			return fn(member, score)
		}
		return true
	})
}

func (z *ZSet) ZRem(ctx context.Context, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	return z.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		defer func() {
			z.Base.deleteKeyWhenNoMember(ctx)
			go safely.RunVoid(z.Compact)
		}()
		var errs []error
		for _, member := range members {
			if err := z.deleteMember(member); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	})
}

func (z *ZSet) ZCount(ctx context.Context, min, max string) (num int64, err error) {
	minBound := &xcmp.Bound[float64]{}
	if err = minBound.ParserMin(min); err != nil {
		return 0, err
	}

	maxBound := &xcmp.Bound[float64]{}
	if err = maxBound.ParserMax(max); err != nil {
		return 0, err
	}
	err = z.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return z.rangeMembers(ctx, func(item *memberScore) (bool, error) {
			match := minBound.MatchMin(item.Score) && maxBound.MatchMax(item.Score)
			if match {
				num++
			}
			return true, nil
		})
	})
	return num, err
}

func (z *ZSet) ZLen(ctx context.Context) (num int64, err error) {
	err = z.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return z.rangeMembers(ctx, func(item *memberScore) (bool, error) {
			num++
			return true, nil
		})
	})
	return num, err
}

func (z *ZSet) ZRank(ctx context.Context, member string) (index int64, score float64, err error) {
	index = -1
	err = z.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		scoreValue, found, err1 := z.memberScore(member)
		if err1 != nil || !found {
			return err1
		}
		var valueIndex int
		err2 := z.rangeMembers(ctx, func(item *memberScore) (bool, error) {
			name := item.MemberString()
			if name != member {
				// 相同分数时，若 member 的字典顺序在前面，也排在前面
				if item.Score < scoreValue || (item.Score == scoreValue && name < member) {
					valueIndex++
				}
			}
			return true, nil
		})
		if err2 != nil {
			return err2
		}
		index = int64(valueIndex)
		score = scoreValue
		return nil
	})
	return index, score, err
}

func (z *ZSet) popXX(ctx context.Context, count int, sort func(a, b *memberScore) bool) (members []string, scores []float64, err error) {
	if count < 1 {
		return nil, nil, nil
	}

	err = z.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		defer z.Base.deleteKeyWhenNoMember(ctx)

		xh := xcontainer.NewTopNHeap[*memberScore](count, sort)
		err1 := z.rangeMembers(ctx, func(item *memberScore) (bool, error) {
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
			if err2 := z.deleteMember(member); err2 != nil {
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

func (z *ZSet) ZPopMax(ctx context.Context, count int) (members []string, scores []float64, err error) {
	return z.popXX(ctx, count, memberScoreSortDesc)
}

func memberScoreSortAsc(a, b *memberScore) bool {
	return a.Score > b.Score
}

func (z *ZSet) ZPopMin(ctx context.Context, count int) (members []string, scores []float64, err error) {
	return z.popXX(ctx, count, memberScoreSortAsc)
}

type memberScore struct {
	Member []byte  `json:"m"`
	Score  float64 `json:"s"`
}

func (fm memberScore) MemberString() string {
	return unsafe.String(unsafe.SliceData(fm.Member), len(fm.Member))
}
