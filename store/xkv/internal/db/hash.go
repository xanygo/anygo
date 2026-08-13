package db

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv"
)

type HashModel struct {
	KeyHash   [32]byte `db:"k,pk"`
	FieldHash [32]byte `db:"f,pk"`

	KeyRaw   string `db:"k_raw"`
	FieldRaw string `db:"f_raw"`

	Value string `db:"v"`

	Created int64 `db:"c"`
	Updated int64 `db:"u"`
}

var _ xkv.Hash[string] = (*Hash)(nil)

type Hash struct {
	Table string
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
	_, err := orm.Delete(ctx, "k=?", h.Meta.KeyHash[:])
	return err
}

func (h *Hash) HSet(ctx context.Context, field string, value string) error {
	now := time.Now().UnixNano()
	fieldHash := KeyHash(field)
	return h.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())

		data := HashModel{
			KeyHash:   h.Meta.KeyHash,
			KeyRaw:    h.Meta.KeyRaw,
			FieldHash: fieldHash,
			FieldRaw:  field,
			Value:     value,
			Updated:   now,
			Created:   now,
		}
		_, err := orm.Upsert(ctx, []string{"k", "f"}, []string{"v", "u"}, data)
		return err
	})
}

func (h *Hash) HMSet(ctx context.Context, data map[string]string) error {
	if len(data) == 0 {
		return nil
	}
	now := time.Now().UnixNano()
	return h.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		var items []HashModel
		for field, value := range data {
			item := HashModel{
				KeyHash:   h.Meta.KeyHash,
				KeyRaw:    h.Meta.KeyRaw,
				FieldHash: KeyHash(field),
				FieldRaw:  field,
				Value:     value,
				Created:   now,
				Updated:   now,
			}
			items = append(items, item)
		}
		_, err := orm.Upsert(ctx, []string{"k", "f"}, []string{"v", "u"}, items...)
		return err
	})
}

func (h *Hash) HGet(ctx context.Context, field string) (value string, found bool, err error) {
	fieldHash := KeyHash(field)
	err = h.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		orm.SelectFields("v")

		v, ok, err1 := orm.First(ctx, "k=? and f=?", h.Meta.KeyHash[:], fieldHash[:])
		if err1 != nil || !ok {
			return err1
		}
		found = true
		value = v.Value
		return nil
	})
	return value, found, err
}

func (h *Hash) HMGet(ctx context.Context, fields ...string) (result map[string]string, err error) {
	if len(fields) == 0 {
		return nil, nil
	}
	fs := make([]any, 0, len(fields))
	for _, f := range fields {
		fs = append(fs, keyHashBytes(f))
	}

	cond := xdb.Condition{}
	cond.And("k=?", h.Meta.KeyHash[:])
	cond.AndInFmt("f in (%s)", fs)
	where, args, err := cond.Build()
	if err != nil {
		return nil, err
	}
	err = h.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		orm.SelectFields("f_raw", "v")

		items, err1 := orm.List(ctx, where, args...)
		if err1 != nil {
			return err1
		}
		result = make(map[string]string, len(items))
		for _, item := range items {
			result[item.FieldRaw] = item.Value
		}
		return nil
	})
	return result, err
}

// checkExists 检查 key 是否还存在，若不存在，则删除 meta
func (h *Hash) checkExists(ctx context.Context, orm *xdb.Model[HashModel]) error {
	orm = orm.Clone().Reset()
	orm.Table(h.GetTable())
	orm.SelectFields("c")
	_, found, err := orm.First(ctx, "k=?", h.Meta.KeyHash[:])
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return h.Meta.delete(ctx, orm.Client())
}

func (h *Hash) HDel(ctx context.Context, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}

	fs := make([]any, 0, len(fields))
	for _, f := range fields {
		fs = append(fs, keyHashBytes(f))
	}

	return h.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		_, err := h.Meta.load(ctx, tx)
		if err != nil {
			return err
		}
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		b := xdb.Condition{}
		b.And("k=?", h.Meta.KeyHash[:])
		b.AndInFmt("f in (%s)", fs)
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
		orm.SelectFields("f_raw", "v")
		for item, err1 := range orm.ListIter(ctx, "k=?", h.Meta.KeyHash[:]) {
			if err1 != nil {
				return err1
			}
			if !fn(item.FieldRaw, item.Value) {
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
	fieldHash := keyHashBytes(field)
	err = h.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		orm.SelectFields("c")
		_, ok, err1 := orm.First(ctx, "k=? and f=?", h.Meta.KeyHash[:], fieldHash)
		if ok {
			found = true
		}
		return err1
	})
	return found, err
}

func (h *Hash) HIncrBy(ctx context.Context, field string, increment int64) (num int64, err error) {
	fieldHash := KeyHash(field)
	now := time.Now().UnixNano()
	err = h.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		orm.SelectFields("v")

		old, found, err1 := orm.First(ctx, "k=? and f=?", h.Meta.KeyHash[:], fieldHash[:])
		if err1 != nil {
			return err1
		}
		num = increment
		if found {
			oldNum, err2 := strconv.ParseInt(old.Value, 10, 64)
			if err2 != nil {
				return err2
			}
			num += oldNum
		}
		data := HashModel{
			KeyHash:   h.Meta.KeyHash,
			KeyRaw:    h.Meta.KeyRaw,
			FieldHash: fieldHash,
			FieldRaw:  field,
			Value:     strconv.FormatInt(num, 10),
			Created:   now,
			Updated:   now,
		}
		_, err3 := orm.Upsert(ctx, []string{"k", "f"}, []string{"v", "u"}, data)
		return err3
	})
	return num, err
}

func (h *Hash) HLen(ctx context.Context) (num int64, err error) {
	err = h.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[HashModel](tx)
		orm.Table(h.GetTable())
		count, err1 := orm.Count(ctx, "*", "k=?", h.Meta.KeyHash[:])
		if err1 == nil {
			num = count
		}
		return err1
	})
	return num, err
}
