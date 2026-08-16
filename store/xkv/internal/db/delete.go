package db

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv/internal"
)

type Delete struct {
	Items []DeleteItem
}

func (d Delete) Delete(ctx context.Context) error {
	if len(d.Items) == 0 {
		return nil
	}
	return d.Items[0].Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		for _, item := range d.Items {
			if err := item.deleteAll(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	})
}

type DeleteItem struct {
	Meta        *Meta
	StringTable string
	ListTable   string
	HashTable   string
	SetTable    string
	ZSetTable   string
}

func (d DeleteItem) deleteAll(ctx context.Context, tx xdb.TxCore) error {
	orm := d.Meta.orm(tx)
	orm.SetSelectFields("dt")
	value, found, err := orm.First(ctx, "k=?", d.Meta.KeyHash[:])
	if err != nil || !found {
		return err
	}

	switch value.DataType {
	case internal.DataTypeString:
		data := &String{
			Table: d.StringTable,
			Meta:  d.Meta,
		}
		err = data.deleteWithKey(ctx, tx)
	case internal.DataTypeList:
		data := &List{
			Table: d.ListTable,
			Meta:  d.Meta,
		}
		err = data.deleteWithKey(ctx, tx)
	case internal.DataTypeHash:
		data := &Hash{
			Table: d.HashTable,
			Meta:  d.Meta,
		}
		err = data.deleteWithKey(ctx, tx)
	case internal.DataTypeSet:
		data := &Set{
			Table: d.SetTable,
			Meta:  d.Meta,
		}
		err = data.deleteWithKey(ctx, tx)
	case internal.DataTypeZSet:
		data := ZSet{
			Table: d.ZSetTable,
			Meta:  d.Meta,
		}
		err = data.deleteWithKey(ctx, tx)
	default:
		return fmt.Errorf("unspported data type: %s", value.DataType)
	}
	if err != nil {
		return err
	}
	return d.Meta.delete(ctx, tx)
}
