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

type HashModel struct {
	Key   string `db:"k,unique_index:idx_k_f"`
	Field string `db:"f,unique_index:idx_k_f"`
	Value string `db:"v"`

	Created int64 `db:"c"`
	Updated int64 `db:"u"`
}

var _ xkv.Hash[string] = (*Hash)(nil)

type Hash struct {
	Table string
	Key   string
	Meta  *Meta
}

func (h *Hash) GetTable() string {
	if h.Table == "" {
		return "xkv_hash"
	}
	return h.Table
}

func (h *Hash) deleteWithKey(ctx context.Context, tx xdb.TxCore) error {
	orm := xdb.NewMode[HashModel](tx)
	orm.Table(h.GetTable())
	_, err := orm.Delete(ctx, "k=?", h.Key)
	return err
}

func (h *Hash) HSet(ctx context.Context, field string, value string) error {
	now := time.Now().Unix()
	return h.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())

		data := HashModel{
			Key:     h.Key,
			Field:   field,
			Value:   value,
			Updated: now,
			Created: now,
		}
		_, err := orm.Upsert(ctx, []string{"k", "f"}, []string{"v", "u"}, data)
		return err
	})
}

func (h *Hash) HMSet(ctx context.Context, data map[string]string) error {
	if len(data) == 0 {
		return nil
	}
	now := time.Now().Unix()
	return h.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		var items []HashModel
		for field, value := range data {
			item := HashModel{
				Key:     h.Key,
				Field:   field,
				Value:   value,
				Created: now,
				Updated: now,
			}
			items = append(items, item)
		}
		_, err := orm.Upsert(ctx, []string{"k", "f"}, []string{"v", "u"}, items...)
		return err
	})
}

func (h *Hash) HGet(ctx context.Context, field string) (value string, found bool, err error) {
	err = h.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())

		orm.OnlyFields("v")
		v, ok, err1 := orm.First(ctx, "k=? and f=?", h.Key, field)
		if err1 != nil || !ok {
			return err1
		}
		found = true
		value = v.Value
		return nil
	})
	return value, found, err
}

// checkExists 检查 key 是否还存在，若不存在，则删除 meta
func (h *Hash) checkExists(ctx context.Context, orm *xdb.Model[HashModel]) error {
	orm = orm.Clone().Reset().OnlyFields("c")
	orm.Table(h.GetTable())
	_, found, err := orm.First(ctx, "k=?", h.Key)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return h.Meta.delete(ctx, orm.Client())
}

func (h *Hash) HDel(ctx context.Context, fields ...string) error {
	return h.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		_, err := h.Meta.load(ctx, tx)
		if err != nil {
			return err
		}
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		b := xdb.Condition{}
		b.And("k=?", h.Key)
		b.And(fmt.Sprintf("f in(%s)", xdb.Placeholder(len(fields))), xslice.ToAnys(fields)...)
		where, args, err1 := b.Build()
		if err1 != nil {
			return err1
		}
		_, err = orm.Delete(ctx, where, args...)
		if err != nil {
			return err
		}
		return h.checkExists(ctx, orm)
	})
}

func (h *Hash) HRange(ctx context.Context, fn func(field string, value string) bool) error {
	return h.Meta.WithReadTx(ctx, func(as context.Context, tx xdb.TxCore, hasMeta bool) error {
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		for item, err1 := range orm.ListIter(ctx, "k=?", h.Key) {
			if err1 != nil {
				return err1
			}
			if !fn(item.Field, item.Value) {
				return io.EOF
			}
		}
		return nil
	})
}

func (h *Hash) HGetAll(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)
	err := h.HRange(ctx, func(field string, value string) bool {
		result[field] = value
		return true
	})
	return result, err
}

func (h *Hash) HExists(ctx context.Context, field string) (found bool, err error) {
	err = h.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		orm.OnlyFields("c")
		_, ok, err1 := orm.First(ctx, "k=? and f=?", h.Key, field)
		if ok {
			found = true
		}
		return err1
	})
	return found, err
}
