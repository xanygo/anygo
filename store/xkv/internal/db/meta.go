package db

import (
	"context"
	"fmt"
	"time"

	"github.com/xanygo/anygo/ds/xcast"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv/internal"
)

var _ xdb.HasTable = (*MetaModel)(nil)

type MetaModel struct {
	Table    string
	DB       *xdb.Client
	Key      string            `db:"k,pk"`
	DataType internal.DataType `db:"dt"`
	Created  int64             `db:"c"`
	Updated  int64             `db:"u"`
	Meta     map[string]any    `db:"meta,codec:json"`
}

func (m MetaModel) TableName() string {
	if m.Table == "" {
		return "xkv_meta"
	}
	return m.Table
}

func (m MetaModel) delete(ctx context.Context, tx xdb.HasDriver) error {
	orm := xdb.NewMode[MetaModel](tx)
	_, err := orm.Delete(ctx, "k=?", m.Key)
	return err
}

func (m MetaModel) incr(field string, dealt int64) (nm MetaModel, num int64) {
	if m.Meta == nil {
		m.Meta = make(map[string]any)
	}
	fv, ok := m.Meta[field]
	if !ok {
		num = 0
	} else {
		num = xcast.ToInteger[int64](fv)
	}
	num = num + dealt
	m.Meta[field] = num
	return m, num
}

func (m MetaModel) save(ctx context.Context, tx xdb.TxCore, data MetaModel) error {
	mod := xdb.NewMode[MetaModel](tx)
	data.Updated = time.Now().Unix()
	_, err := mod.Upsert(ctx, []string{"k"}, []string{"meta", "u"}, data)
	return err
}

func (m MetaModel) WithWriteTx(ctx context.Context, do func(ctx context.Context, tx xdb.TxCore) error) error {
	return m.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		if err1 := m.checkWriteType(ctx, tx); err1 != nil {
			return err1
		}
		return do(ctx, tx)
	})
}

func (m MetaModel) WithReadTx(ctx context.Context, do func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error) error {
	return m.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		found, err1 := m.checkReadType(ctx, tx)
		if err1 != nil {
			return err1
		}
		return do(ctx, tx, found)
	})
}

func (m MetaModel) WithTx(ctx context.Context, do func(ctx context.Context, tx xdb.TxCore) error) error {
	te, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	return xdb.WithTx(ctx, te, func(ctx context.Context, tx xdb.TxCore) error {
		return do(ctx, te)
	})
}

func (m MetaModel) checkWriteType(ctx context.Context, tx xdb.TxCore) error {
	mod := xdb.NewMode[MetaModel](tx)
	value, found, err := mod.First(ctx, "k=?", m.Key)
	if err != nil {
		return err
	}
	if found {
		if value.DataType == m.DataType {
			return nil
		}
		return fmt.Errorf("canot write %s on type %s", m.DataType.String(), value.DataType.String())
	}
	return mod.Insert(ctx, m)
}

func (m MetaModel) load(ctx context.Context, tx xdb.TxCore) (MetaModel, error) {
	orm := xdb.NewMode[MetaModel](tx)
	value, found, err := orm.First(ctx, "k=?", m.Key)
	if err != nil {
		return MetaModel{}, err
	}
	if found {
		if value.DataType == m.DataType {
			value.Table = m.Table
			return value, nil
		}
		return MetaModel{}, fmt.Errorf("canot load %s on type %s", m.DataType.String(), value.DataType.String())
	}
	return m, nil
}

func (m MetaModel) checkReadType(ctx context.Context, tx xdb.TxCore) (bool, error) {
	orm := xdb.NewMode[MetaModel](tx)
	orm.OnlyFields("dt")
	value, found, err := orm.First(ctx, "k=?", m.Key)
	if err != nil {
		return false, err
	}
	if found {
		if value.DataType == m.DataType || m.DataType == internal.DataTypeAny {
			return true, nil
		}
		return false, fmt.Errorf("canot read %s on type %s", m.DataType.String(), value.DataType.String())
	}
	return false, nil
}
