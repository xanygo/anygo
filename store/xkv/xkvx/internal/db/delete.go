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
	Meta        MetaModel
	StringTable string
	ListTable   string
}

func (d DeleteItem) deleteAll(ctx context.Context, tx xdb.TxCore) error {
	key := d.Meta.Key
	mod := xdb.NewMode[MetaModel](tx)
	value, found, err := mod.First(ctx, "k=?", key)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	switch value.DataType {
	case internal.DataTypeString:
		data := StringModel{
			Table: d.StringTable,
			Key:   key,
		}
		err = data.deleteWithKey(ctx, tx)
	case internal.DataTypeList:
		data := ListModel{
			Table: d.StringTable,
			Key:   key,
		}
		err = data.deleteWithKey(ctx, tx)
	case internal.DataTypeHash:
		data := HashModel{
			Table: d.StringTable,
			Key:   key,
		}
		err = data.deleteWithKey(ctx, tx)
	case internal.DataTypeSet:
		data := SetModel{
			Table: d.StringTable,
			Key:   key,
		}
		err = data.deleteWithKey(ctx, tx)
	case internal.DataTypeZSet:
		data := ZSetModel{
			Table: d.StringTable,
			Key:   key,
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
