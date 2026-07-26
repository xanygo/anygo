package db

import (
	"context"
	"fmt"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv"
)

var _ xdb.HasTable = ZSetModel{}

type ZSetModel struct {
	Table   string
	Key     string  `db:"k,unique_index:idx_k_m,index:idx_k_i"`
	Member  string  `db:"m,unique_index:idx_k_m"`
	Score   float64 `db:"s,index:idx_k_i"`
	Created int64   `db:"c"`
	Updated int64   `db:"u"`
}

func (dt ZSetModel) TableName() string {
	if dt.Table == "" {
		return "xkv_zset"
	}
	return dt.Table
}

func (dt ZSetModel) deleteWithKey(ctx context.Context, tx xdb.TxCore) error {
	ms := xdb.NewMode[ZSetModel](tx)
	_, err := ms.Delete(ctx, "k=?", dt.Key)
	return err
}

var _ xkv.ZSet[string] = (*ZSet)(nil)

type ZSet struct {
	Table string
	Key   string
	Meta  MetaModel
}

func (z *ZSet) ZAdd(ctx context.Context, score float64, member string) error {
	now := time.Now().Unix()
	return z.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[ZSetModel](tx)
		data := ZSetModel{
			Table:   z.Table,
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

func (z *ZSet) ZScore(ctx context.Context, member string) (score float64, found bool, err error) {
	err = z.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[ZSetModel](tx)
		orm.OnlyFields("s")
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
		orm.OnlyFields("m", "s")
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
		_, err = orm.Delete(ctx, where, args...)
		if err != nil {
			return err
		}
		return z.checkExists(ctx, orm)
	})
}

// checkExists 检查 key 是否还存在，若不存在，则删除 meta
func (z *ZSet) checkExists(ctx context.Context, orm *xdb.Model[ZSetModel]) error {
	orm = orm.Clone().Reset().OnlyFields("c")
	_, found, err := orm.First(ctx, "k=?", z.Key)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return z.Meta.delete(ctx, orm.Client())
}
