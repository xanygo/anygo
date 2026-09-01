package db

import (
	"context"
	"strconv"
	"time"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/xor"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xkv/internal"
)

type StringModel struct {
	ID      int64    `db:"id,pk,auto_inc"`
	KeyHash [32]byte `db:"k,unique"`
	KeyRaw  string   `db:"k_raw"`
	Value   string   `db:"v"`
	Created int64    `db:"c"`
	Updated int64    `db:"u"`
}

var _ xkv.String[string] = (*String)(nil)

type String struct {
	Table string
	Meta  *Meta
}

func (d *String) GetTable() string {
	if d.Table == "" {
		return "xkv_string"
	}
	return d.Table
}

func (d *String) deleteWithKey(ctx context.Context, tx xdb.DBCore) error {
	orm := xor.New[StringModel](tx)
	orm.Table(d.GetTable())
	_, err := orm.Delete(ctx, xor.Where("k=?", d.Meta.KeyHash[:]))
	return err
}

func (d *String) Set(ctx context.Context, value string) error {
	now := time.Now().UnixNano()
	return d.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		data := StringModel{
			KeyHash: d.Meta.KeyHash,
			KeyRaw:  d.Meta.KeyRaw,
			Value:   value,
			Created: now,
			Updated: now,
		}
		orm := xor.New[StringModel](tx)
		orm.Table(d.GetTable())
		_, err := orm.Upsert(ctx, []string{"k"}, []string{"v", "u"}, data)
		return err
	})
}

func (d *String) SetNX(ctx context.Context, value string) (ok bool, err error) {
	now := time.Now().UnixNano()
	err = d.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		_, has, err1 := d.Meta.loadExists(ctx, tx)
		if err1 != nil || has {
			return err1
		}
		meta := MetaModel{
			KeyHash:  d.Meta.KeyHash,
			KeyRaw:   d.Meta.KeyRaw,
			DataType: internal.DataTypeString,
			Created:  now,
			Updated:  now,
		}
		if err1 = d.Meta.save(ctx, tx, meta); err1 != nil {
			return err1
		}
		data := StringModel{
			KeyHash: d.Meta.KeyHash,
			KeyRaw:  d.Meta.KeyRaw,
			Value:   value,
			Created: now,
			Updated: now,
		}
		orm := xor.New[StringModel](tx)
		orm.Table(d.GetTable())
		_, err2 := orm.Upsert(ctx, []string{"k"}, []string{"v", "u"}, data)
		ok = err2 == nil
		return err2
	})
	return ok, err
}

func (d *String) Get(ctx context.Context) (val string, ok bool, err error) {
	err = d.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xor.New[StringModel](tx)
		orm.Table(d.GetTable())
		value, found, err1 := orm.First(ctx, xor.Where("k=?", d.Meta.KeyHash[:]), xor.Columns("v"))
		if err1 != nil {
			return err1
		}
		if found {
			val = value.Value
			ok = true
		}
		return nil
	})
	return val, ok, err
}

func (d *String) GetDel(ctx context.Context) (val string, ok bool, err error) {
	err = d.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		_, has, err1 := d.Meta.loadExists(ctx, tx)
		if err1 != nil || !has {
			return err1
		}
		orm := xor.New[StringModel](tx)
		orm.Table(d.GetTable())
		value, found, err2 := orm.First(ctx, xor.Where("k=?", d.Meta.KeyHash[:]), xor.Columns("v"))
		if err2 != nil || !found {
			return err2
		}
		val = value.Value
		ok = true
		_, err = orm.Delete(ctx, xor.Where("k=?", d.Meta.KeyHash[:]))
		return err
	})
	return val, ok, err
}

func (d *String) GetSet(ctx context.Context, value string) (old string, ok bool, err error) {
	now := time.Now().UnixNano()
	err = d.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		orm := xor.New[StringModel](tx)
		orm.Table(d.GetTable())
		item, found, err2 := orm.First(ctx, xor.Where("k=?", d.Meta.KeyHash[:]), xor.Columns("v"))
		if err2 != nil {
			return err2
		}

		nv := StringModel{
			KeyHash: d.Meta.KeyHash,
			KeyRaw:  d.Meta.KeyRaw,
			Value:   value,
			Created: now,
			Updated: now,
		}
		_, err3 := orm.Upsert(ctx, []string{"k"}, []string{"v", "u"}, nv)
		if err3 != nil {
			return err3
		}
		if found {
			old = item.Value
			ok = true
		}
		return nil
	})
	return old, ok, err
}

func (d *String) Incr(ctx context.Context) (num int64, err error) {
	return d.IncrBy(ctx, 1)
}

func (d *String) IncrBy(ctx context.Context, incr int64) (num int64, err error) {
	now := time.Now().UnixNano()
	err = d.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		orm := xor.New[StringModel](tx)
		orm.Table(d.GetTable())
		val, found, err1 := orm.First(ctx, xor.Where("k=?", d.Meta.KeyHash[:]), xor.Columns("v"))
		if err1 != nil {
			return err1
		}

		var strVal string
		if found {
			num, err1 = strconv.ParseInt(val.Value, 10, 64)
			if err1 != nil {
				return err1
			}
			num = num + incr
		} else {
			num = incr
		}

		strVal = strconv.FormatInt(num, 10)

		data := StringModel{
			KeyHash: d.Meta.KeyHash,
			KeyRaw:  d.Meta.KeyRaw,
			Value:   strVal,
			Created: now,
			Updated: now,
		}
		_, err1 = orm.Upsert(ctx, []string{"k"}, []string{"v", "u"}, data)
		return err1
	})
	return num, err
}

func (d *String) IncrByFloat(ctx context.Context, incr float64) (num float64, err error) {
	now := time.Now().UnixNano()
	err = d.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		orm := xor.New[StringModel](tx)
		orm.Table(d.GetTable())
		val, found, err1 := orm.First(ctx, xor.Where("k=?", d.Meta.KeyHash[:]), xor.Columns("v"))
		if err1 != nil {
			return err1
		}

		var strVal string
		if found {
			num, err1 = strconv.ParseFloat(val.Value, 64)
			if err1 != nil {
				return err1
			}
			num = num + incr
		} else {
			num = incr
		}

		strVal = strconv.FormatFloat(num, 'g', -1, 64)

		data := StringModel{
			KeyHash: d.Meta.KeyHash,
			KeyRaw:  d.Meta.KeyRaw,
			Value:   strVal,
			Created: now,
			Updated: now,
		}
		_, err1 = orm.Upsert(ctx, []string{"k"}, []string{"v", "u"}, data)
		return err1
	})
	return num, err
}

func (d *String) Decr(ctx context.Context) (num int64, err error) {
	return d.IncrBy(ctx, -1)
}
