package db

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv"
)

type SetModel struct {
	Key     string `db:"k,unique_index:idx_k_m"`
	Member  string `db:"m,unique_index:idx_k_m"`
	Created int64  `db:"c"`
}

var _ xkv.Set[string] = (*Set)(nil)

type Set struct {
	Table string
	Key   string
	Meta  *Meta
}

func (s *Set) GetTable() string {
	if s.Table == "" {
		return "xkv_set"
	}
	return s.Table
}

func (s *Set) deleteWithKey(ctx context.Context, tx xdb.TxCore) error {
	orm := xdb.NewMode[SetModel](tx)
	orm.Table(s.GetTable())
	_, err := orm.Delete(ctx, "k=?", s.Key)
	return err
}

func (s *Set) SAdd(ctx context.Context, members ...string) (num int64, err error) {
	if len(members) == 0 {
		return 0, nil
	}
	now := time.Now().Unix()
	err = s.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[SetModel](tx)
		orm.Table(s.GetTable())
		items := make([]SetModel, 0, len(members))
		for _, member := range members {
			item := SetModel{
				Key:     s.Key,
				Member:  member,
				Created: now,
			}
			items = append(items, item)
		}
		num, err = orm.Upsert(ctx, []string{"k", "m"}, nil, items...)
		return err
	})
	return num, err
}

func (s *Set) SRem(ctx context.Context, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	return s.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		_, err := s.Meta.load(ctx, tx)
		if err != nil {
			return err
		}
		cond := xdb.Condition{}
		cond.And("k=?", s.Key)
		cond.And(fmt.Sprintf("m in(%s)", xdb.Placeholder(len(members))), xslice.ToAnys(members)...)
		where, args, err := cond.Build()
		if err != nil {
			return err
		}
		orm := xdb.NewMode[SetModel](tx)
		orm.Table(s.GetTable())
		_, err = orm.Delete(ctx, where, args...)
		if err != nil {
			return err
		}
		return s.checkExists(ctx, orm)
	})
}

// checkExists 检查 key 是否还存在，若不存在，则删除 meta
func (s *Set) checkExists(ctx context.Context, orm *xdb.Model[SetModel]) error {
	orm = orm.Clone().Reset()
	orm.Table(s.GetTable())
	orm.SelectFields("c")
	_, found, err := orm.First(ctx, "k=?", s.Key)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return s.Meta.delete(ctx, orm.Client())
}

func (s *Set) SRange(ctx context.Context, fn func(member string) bool) error {
	return s.Meta.WithReadTx(ctx, func(as context.Context, tx xdb.TxCore, hasMeta bool) error {
		orm := xdb.NewMode[SetModel](tx)
		orm.Table(s.GetTable())
		orm.SelectFields("m")
		for item, err1 := range orm.ListIter(ctx, "k=?", s.Key) {
			if err1 != nil {
				return err1
			}
			if !fn(item.Member) {
				return io.EOF
			}
		}
		return nil
	})
}

func (s *Set) SMembers(ctx context.Context) ([]string, error) {
	var result []string
	err := s.SRange(ctx, func(member string) bool {
		result = append(result, member)
		return true
	})
	return result, err
}

func (s *Set) SCard(ctx context.Context) (num int64, err error) {
	err = s.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[SetModel](tx)
		orm.Table(s.GetTable())
		var err1 error
		num, err1 = orm.Count(ctx, "*", "k=?", s.Key)
		return err1
	})
	return num, err
}

func (s *Set) SIsMember(ctx context.Context, member string) (ok bool, err error) {
	err = s.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[SetModel](tx)
		orm.Table(s.GetTable())
		orm.SelectFields("c")
		_, found, err1 := orm.First(ctx, "k=? and m=?", s.Key, member)
		if err1 == nil {
			ok = found
		}
		return err1
	})
	return ok, err
}

func (s *Set) SMIsMember(ctx context.Context, members []string) (oks []bool, err error) {
	oks = make([]bool, len(members))
	if len(members) == 0 {
		return oks, nil
	}
	err = s.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[SetModel](tx)
		orm.Table(s.GetTable())
		orm.SelectFields("m")

		cond := xdb.Condition{}
		cond.And("k=?", s.Key)
		cond.AndInFmt("m in (%s)", xslice.ToAnys(members))
		where, args, err0 := cond.Build()
		if err0 != nil {
			return err0
		}
		items, err1 := orm.List(ctx, where, args...)
		if err1 == nil {
			mp := make(map[string]bool, len(items))
			for _, item := range items {
				mp[item.Member] = true
			}
			for i, member := range members {
				oks[i] = mp[member]
			}
		}
		return err1
	})
	return oks, err
}
