package db

import (
	"context"
	"strconv"
	"time"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv"
)

type StringModel struct {
	Key     string `db:"k,pk"`
	Value   string `db:"v"`
	Created int64  `db:"c"`
	Updated int64  `db:"u"`
}

var _ xkv.String[string] = (*String)(nil)

type String struct {
	Key   string
	Table string
	Meta  *Meta
}

func (d *String) GetTable() string {
	if d.Table == "" {
		return "xkv_string"
	}
	return d.Table
}

func (d *String) deleteWithKey(ctx context.Context, tx xdb.TxCore) error {
	orm := xdb.NewMode[StringModel](tx)
	orm.Table(d.GetTable())
	_, err := orm.Delete(ctx, "k=?", d.Key)
	return err
}

func (d *String) Set(ctx context.Context, value string) error {
	now := time.Now().Unix()
	return d.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		data := StringModel{
			Key:     d.Key,
			Value:   value,
			Created: now,
			Updated: now,
		}
		orm := xdb.NewMode[StringModel](tx)
		orm.Table(d.GetTable())
		_, err := orm.Upsert(ctx, []string{"k"}, []string{"v", "u"}, data)
		return err
	})
}

func (d *String) Get(ctx context.Context) (val string, ok bool, err error) {
	err = d.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[StringModel](tx)
		orm.Table(d.GetTable())
		orm.OnlyFields("v")
		value, found, err1 := orm.First(ctx, "k=?", d.Key)
		if err1 != nil {
			return err1
		}
		if found {
			val = value.Value
			ok = true
			return nil
		}

		return nil
	})
	return val, ok, err
}

func (d *String) Incr(ctx context.Context) (num int64, err error) {
	now := time.Now().Unix()
	err = d.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[StringModel](tx)
		orm.Table(d.GetTable())
		orm.OnlyFields("v")
		val, found, err1 := orm.First(ctx, "k=?", d.Key)
		if err1 != nil {
			return err1
		}

		var strVal string
		if found {
			num, err1 = strconv.ParseInt(val.Value, 10, 64)
			if err1 != nil {
				return err1
			}
			num = num + 1
			strVal = strconv.FormatInt(num, 10)
		} else {
			num = 1
			strVal = "1"
		}
		data := StringModel{
			Key:     d.Key,
			Value:   strVal,
			Created: now,
			Updated: now,
		}
		orm.Reset()
		orm.Table(d.GetTable())
		_, err1 = orm.Upsert(ctx, []string{"k"}, []string{"v", "u"}, data)
		return err1
	})
	return num, err
}

func (d *String) Decr(ctx context.Context) (num int64, err error) {
	now := time.Now().Unix()
	err = d.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		orm := xdb.NewMode[StringModel](tx)
		orm.Table(d.GetTable())
		orm.OnlyFields("v")
		val, found, err1 := orm.First(ctx, "k=?", d.Key)
		if err1 != nil {
			return err1
		}

		var strVal string
		if found {
			num, err1 = strconv.ParseInt(val.Value, 10, 64)
			if err1 != nil {
				return err1
			}
			num = num - 1
			strVal = strconv.FormatInt(num, 10)
		} else {
			num = -1
			strVal = "-1"
		}
		data := StringModel{
			Key:     d.Key,
			Value:   strVal,
			Created: now,
			Updated: now,
		}
		orm.Reset()
		orm.Table(d.GetTable())
		_, err1 = orm.Upsert(ctx, []string{"k"}, []string{"v", "u"}, data)
		return err1
	})
	return num, err
}
