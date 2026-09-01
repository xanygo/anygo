package db

import (
	"context"
	"fmt"
	"time"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/xor"
	"github.com/xanygo/anygo/store/xkv"
)

type ZSetModel struct {
	ID         int64    `db:"id,pk,auto_inc"`
	KeyHash    [32]byte `db:"k,unique_index=k_m[1],index=k_s[1]"`
	MemberHash [32]byte `db:"m,unique_index=k_m[2]"`

	KeyRaw    string `db:"k_raw"`
	MemberRaw string `db:"m_raw"`

	Score   float64 `db:"s,index=k_s[2]"`
	Created int64   `db:"c"`
	Updated int64   `db:"u"`
}

var _ xkv.ZSet[string] = (*ZSet)(nil)

type ZSet struct {
	Table string
	Meta  *Meta
}

func (z *ZSet) GetTable() string {
	if z.Table == "" {
		return "xkv_zset"
	}
	return z.Table
}

func (z *ZSet) orm(tx xdb.DBCore) *xor.Model[ZSetModel] {
	orm := xor.New[ZSetModel](tx)
	return orm.Table(z.GetTable())
}

func (z *ZSet) deleteWithKey(ctx context.Context, tx xdb.DBCore) error {
	orm := z.orm(tx)
	_, err := orm.Delete(ctx, xor.Where("k=?", z.Meta.KeyHash[:]))
	return err
}

func (z *ZSet) ZAdd(ctx context.Context, score float64, member string) error {
	memberHash := KeyHash(member)
	now := time.Now().UnixNano()
	return z.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		orm := z.orm(tx)
		data := ZSetModel{
			KeyHash:    z.Meta.KeyHash,
			KeyRaw:     z.Meta.KeyRaw,
			MemberHash: memberHash,
			MemberRaw:  member,
			Score:      score,
			Created:    now,
			Updated:    now,
		}
		_, err := orm.Upsert(ctx, []string{"k", "m"}, []string{"s", "u"}, data)
		return err
	})
}

func (z *ZSet) ZIncrBy(ctx context.Context, inc float64, member string) (num float64, err error) {
	now := time.Now().UnixNano()
	memberHash := KeyHash(member)
	err = z.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		orm := z.orm(tx)

		item, ok, err1 := orm.First(ctx, xor.Where("k=? and m=?", z.Meta.KeyHash[:], memberHash[:]), xor.Columns("s"))
		if err1 != nil {
			return err1
		}
		var old float64
		if ok {
			old = item.Score
		}
		num = old + inc
		data := ZSetModel{
			KeyHash:    z.Meta.KeyHash,
			KeyRaw:     z.Meta.KeyRaw,
			MemberHash: memberHash,
			MemberRaw:  member,
			Score:      num,
			Created:    now,
			Updated:    now,
		}
		_, err = orm.Upsert(ctx, []string{"k", "m"}, []string{"s", "u"}, data)
		return err
	})

	return num, err
}

func (z *ZSet) minMaxCond(min, max string) (where string, args []any, err error) {
	minBound := &xcmp.Bound[float64]{}
	if err1 := minBound.ParserMin(min); err1 != nil {
		return "", nil, err1
	}

	maxBound := &xcmp.Bound[float64]{}
	if err1 := maxBound.ParserMax(max); err1 != nil {
		return "", nil, err1
	}

	cond := &xdb.Condition{}
	cond.And("k=?", z.Meta.KeyHash[:])
	if !minBound.Inf {
		op := anygo.Ternary(minBound.Exclude, ">", ">=")
		cond.And(fmt.Sprintf("s %s ?", op), minBound.Value)
	}
	if !maxBound.Inf {
		op := anygo.Ternary(minBound.Exclude, "<", "<=")
		cond.And(fmt.Sprintf("s %s ?", op), maxBound.Value)
	}
	return cond.Build()
}

func (z *ZSet) ZCount(ctx context.Context, min, max string) (num int64, err error) {
	where, args, err1 := z.minMaxCond(min, max)
	if err1 != nil {
		return 0, err1
	}
	err = z.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := z.orm(tx)
		num, err1 = orm.Count(ctx, "*", xor.Where(where, args...))
		return err1
	})
	return num, err
}

func (z *ZSet) ZLen(ctx context.Context) (num int64, err error) {
	return z.ZCount(ctx, "-inf", "+inf")
}

func (z *ZSet) ZScore(ctx context.Context, member string) (score float64, found bool, err error) {
	memberHash := keyHashBytes(member)
	err = z.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := z.orm(tx)

		item, ok, err1 := orm.First(ctx, xor.Where("k=? and m=?", z.Meta.KeyHash[:], memberHash), xor.Columns("s"))
		if err1 != nil || !ok {
			return err1
		}
		found = true
		score = item.Score
		return nil
	})
	return score, found, err
}

func (z *ZSet) ZRange(ctx context.Context, fn func(member string, score float64) bool) error {
	return z.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := z.orm(tx)

		for item, err := range orm.ListIter(ctx, xor.Where("k=?", z.Meta.KeyHash[:]), xor.OrderBy("s asc"), xor.Columns("m_raw", "s")) {
			if err != nil {
				return err
			}
			if !fn(item.MemberRaw, item.Score) {
				return nil
			}
		}
		return nil
	})
}

func (z *ZSet) ZRangeByScore(ctx context.Context, min string, max string, fn func(member string, score float64) bool) error {
	where, args, err1 := z.minMaxCond(min, max)
	if err1 != nil {
		return err1
	}

	return z.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := z.orm(tx)

		for item, err := range orm.ListIter(ctx, xor.Where(where, args...), xor.OrderBy("s asc"), xor.Columns("m_raw", "s")) {
			if err != nil {
				return err
			}
			if !fn(item.MemberRaw, item.Score) {
				return nil
			}
		}
		return nil
	})
}

func (z *ZSet) ZRem(ctx context.Context, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	hashMembers := make([]any, 0, len(members))
	for _, member := range members {
		hashMembers = append(hashMembers, keyHashBytes(member))
	}
	return z.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		_, err := z.Meta.load(ctx, tx)
		if err != nil {
			return err
		}
		cond := &xdb.Condition{}
		cond.And("k=?", z.Meta.KeyHash[:])
		cond.AndInFmt("m in(%s)", hashMembers)

		orm := z.orm(tx)
		_, err = orm.Delete(ctx, xor.WhereByCond(cond))
		if err != nil {
			return err
		}
		return z.checkExists(ctx, orm)
	})
}

func (z *ZSet) ZRemRangeByScore(ctx context.Context, min, max string) (num int64, err error) {
	where, args, err := z.minMaxCond(min, max)
	if err != nil {
		return 0, err
	}
	err = z.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		_, err1 := z.Meta.load(ctx, tx)
		if err1 != nil {
			return err1
		}
		orm := z.orm(tx)
		num, err1 = orm.Delete(ctx, xor.Where(where, args...))
		if err1 != nil {
			return err1
		}
		return z.checkExists(ctx, orm)
	})
	return num, err
}

// checkExists 检查 key 是否还存在，若不存在，则删除 meta
func (z *ZSet) checkExists(ctx context.Context, orm *xor.Model[ZSetModel]) error {
	_, found, err := orm.First(ctx, xor.Where("k=?", z.Meta.KeyHash[:]), xor.Columns("c"))
	if err != nil || found {
		return err
	}
	return z.Meta.delete(ctx, orm.DB())
}

func (z *ZSet) ZRank(ctx context.Context, member string) (index int64, score float64, err error) {
	index = -1
	memberHash := keyHashBytes(member)
	err = z.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		orm := z.orm(tx)

		one, found, err1 := orm.First(ctx, xor.Where("k=? and m=?", z.Meta.KeyHash[:], memberHash), xor.Columns("s"))
		if err1 != nil || !found {
			return err1
		}
		score = one.Score

		cond := &xdb.Condition{}
		cond.And("k=?", z.Meta.KeyHash[:])
		cond.And("( s<? or ( s=? and m<? ) )", score, score, memberHash)

		var err2 error
		index, err2 = orm.Count(ctx, "*", xor.WhereByCond(cond))
		return err2
	})
	return index, score, err
}

func (z *ZSet) popXX(ctx context.Context, count int, orderBy string) (members []string, scores []float64, err error) {
	if count < 1 {
		return nil, nil, nil
	}
	err = z.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		orm := z.orm(tx)
		values, err1 := orm.List(ctx, xor.Where("k=?", z.Meta.KeyHash[:]), xor.OrderBy("s "+orderBy), xor.Limit(count), xor.Columns("s", "m", "m_raw"))
		if err1 != nil {
			return err1
		}
		hashMembers := make([]any, 0, len(values))
		for _, item := range values {
			members = append(members, item.MemberRaw)
			scores = append(scores, item.Score)
			hashMembers = append(hashMembers, item.MemberHash[:])
		}
		cond := &xdb.Condition{}
		cond.And("k=?", z.Meta.KeyHash[:])
		cond.AndInFmt("m in (%s)", hashMembers)

		_, err3 := orm.Delete(ctx, xor.WhereByCond(cond))
		if err3 != nil {
			return err3
		}
		return z.checkExists(ctx, orm)
	})
	if err != nil {
		return nil, nil, err
	}
	return members, scores, err
}

func (z *ZSet) ZPopMax(ctx context.Context, count int) (members []string, scores []float64, err error) {
	return z.popXX(ctx, count, "desc")
}

func (z *ZSet) ZPopMin(ctx context.Context, count int) (members []string, scores []float64, err error) {
	return z.popXX(ctx, count, "asc")
}
