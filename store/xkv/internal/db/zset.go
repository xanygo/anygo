package db

import (
	"context"
	"fmt"
	"time"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv"
)

type ZSetModel struct {
	Key     string  `db:"k,unique_index:idx_k_m,index:idx_k_i"`
	Member  string  `db:"m,unique_index:idx_k_m"`
	Score   float64 `db:"s,index:idx_k_i"`
	Created int64   `db:"c"`
	Updated int64   `db:"u"`
}

var _ xkv.ZSet[string] = (*ZSet)(nil)

type ZSet struct {
	Table string
	Key   string
	Meta  *Meta
}

func (z *ZSet) GetTable() string {
	if z.Table == "" {
		return "xkv_zset"
	}
	return z.Table
}

func (z *ZSet) deleteWithKey(ctx context.Context, tx xdb.TxCore) error {
	orm := xdb.NewMode[ZSetModel](tx)
	orm.Table(z.GetTable())
	_, err := orm.Delete(ctx, "k=?", z.Key)
	return err
}

func (z *ZSet) ZAdd(ctx context.Context, score float64, member string) error {
	now := time.Now().Unix()
	return z.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[ZSetModel](tx)
		orm.Table(z.GetTable())
		data := ZSetModel{
			Key:     z.Key,
			Member:  member,
			Score:   score,
			Created: now,
			Updated: now,
		}
		_, err := orm.Upsert(ctx, []string{"k", "m"}, []string{"s", "u"}, data)
		return err
	})
}

func (z *ZSet) ZIncrBy(ctx context.Context, inc float64, member string) (num float64, err error) {
	now := time.Now().Unix()
	err = z.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[ZSetModel](tx)
		orm.Table(z.GetTable())
		orm.SelectFields("s")

		item, ok, err1 := orm.First(ctx, "k=? and m=?", z.Key, member)
		if err1 != nil {
			return err1
		}
		var old float64
		if ok {
			old = item.Score
		}
		num = old + inc
		data := ZSetModel{
			Key:     z.Key,
			Member:  member,
			Score:   num,
			Created: now,
			Updated: now,
		}
		_, err = orm.Upsert(ctx, []string{"k", "m"}, []string{"s", "u"}, data)
		return err
	})

	return num, err
}

func (z *ZSet) ZCount(ctx context.Context, min, max string) (num int64, err error) {
	err = z.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}

		minBound := &xcmp.Bound[float64]{}
		if err1 := minBound.ParserMin(min); err1 != nil {
			return err1
		}

		maxBound := &xcmp.Bound[float64]{}
		if err1 := maxBound.ParserMin(max); err1 != nil {
			return err1
		}

		cond := xdb.Condition{}
		cond.And("k=?", z.Key)
		if !minBound.Inf {
			op := anygo.Ternary(minBound.Exclude, ">", ">=")
			cond.And(fmt.Sprintf("s %s ?", op), minBound.Value)
		}
		if !maxBound.Inf {
			op := anygo.Ternary(minBound.Exclude, "<", "<=")
			cond.And(fmt.Sprintf("s %s ?", op), maxBound.Value)
		}
		where, args, err2 := cond.Build()
		if err2 != nil {
			return err2
		}

		orm := xdb.NewMode[ZSetModel](tx)
		orm.Table(z.GetTable())
		num, err2 = orm.Count(ctx, "*", where, args...)
		return err2
	})
	return num, err
}

func (z *ZSet) ZLen(ctx context.Context) (num int64, err error) {
	return z.ZCount(ctx, "-inf", "+inf")
}

func (z *ZSet) ZScore(ctx context.Context, member string) (score float64, found bool, err error) {
	err = z.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[ZSetModel](tx)
		orm.Table(z.GetTable())
		orm.SelectFields("s")

		item, ok, err1 := orm.First(ctx, "k=? and m=?", z.Key, member)
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
	return z.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[ZSetModel](tx)
		orm.Table(z.GetTable())
		orm.SelectFields("m", "s")

		for item, err := range orm.ListIter(ctx, "k=?", z.Key) {
			if err != nil {
				return err
			}
			if !fn(item.Member, item.Score) {
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
	return z.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		_, err := z.Meta.load(ctx, tx)
		if err != nil {
			return err
		}
		cond := xdb.Condition{}
		cond.And("k=?", z.Key)
		cond.And(fmt.Sprintf("m in(%s)", xdb.Placeholder(len(members))), xslice.ToAnys(members)...)
		where, args, err := cond.Build()
		if err != nil {
			return err
		}
		orm := xdb.NewMode[ZSetModel](tx)
		orm.Table(z.GetTable())
		_, err = orm.Delete(ctx, where, args...)
		if err != nil {
			return err
		}
		return z.checkExists(ctx, orm)
	})
}

// checkExists 检查 key 是否还存在，若不存在，则删除 meta
func (z *ZSet) checkExists(ctx context.Context, orm *xdb.Model[ZSetModel]) error {
	orm = orm.Clone().Reset()
	orm.Table(z.GetTable())
	orm.SelectFields("c")
	_, found, err := orm.First(ctx, "k=?", z.Key)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return z.Meta.delete(ctx, orm.Client())
}

func (z *ZSet) ZRank(ctx context.Context, member string) (index int64, score float64, err error) {
	index = -1
	err = z.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[ZSetModel](tx)
		orm.Table(z.GetTable())
		orm.SelectFields("s")

		one, found, err1 := orm.First(ctx, "k=? and m=?", z.Key, member)
		if err1 != nil || !found {
			return err1
		}
		score = one.Score

		cond := xdb.Condition{}
		cond.And("k=?", z.Key)
		cond.And("( s<? or ( s=? and m<? ) )", score, score, member)
		where, args, err2 := cond.Build()
		if err2 != nil {
			return err2
		}
		index, err2 = orm.Count(ctx, "*", where, args...)
		return err2
	})
	return index, score, err
}
