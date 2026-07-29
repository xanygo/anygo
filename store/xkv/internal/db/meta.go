package db

import (
	"context"
	"fmt"
	"time"

	"github.com/xanygo/anygo/ds/xcast"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv/internal"
)

type MetaModel struct {
	Key      string            `db:"k,pk"`
	DataType internal.DataType `db:"dt"`
	Created  int64             `db:"c"`
	Updated  int64             `db:"u"`
	Meta     map[string]any    `db:"meta,codec:json"`
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

type Meta struct {
	Table    string
	DB       *xdb.Client
	Key      string
	DataType internal.DataType
}

func (m *Meta) GetTable() string {
	if m.Table == "" {
		return "xkv_meta"
	}
	return m.Table
}

func (m *Meta) orm(tx xdb.HasDriver) *xdb.Model[MetaModel] {
	orm := xdb.NewMode[MetaModel](tx)
	orm.Table(m.GetTable())
	return orm
}

func (m *Meta) delete(ctx context.Context, tx xdb.HasDriver) error {
	orm := m.orm(tx)
	_, err := orm.Delete(ctx, "k=?", m.Key)
	return err
}

func (m *Meta) save(ctx context.Context, tx xdb.TxCore, data MetaModel) error {
	now := time.Now().Unix()
	orm := m.orm(tx)
	if data.Created == 0 {
		data.Created = now
	}
	if data.Updated == 0 {
		data.Updated = now
	}
	_, err := orm.Upsert(ctx, []string{"k"}, []string{"meta", "u"}, data)
	return err
}

func (m *Meta) WithWriteTx(ctx context.Context, do func(ctx context.Context, tx xdb.TxCore) error) error {
	return m.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		if err1 := m.checkWriteType(ctx, tx); err1 != nil {
			return err1
		}
		return do(ctx, tx)
	})
}

func (m *Meta) WithReadTx(ctx context.Context, do func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error) error {
	return m.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		found, err1 := m.checkReadType(ctx, tx)
		if err1 != nil {
			return err1
		}
		return do(ctx, tx, found)
	})
}

func (m *Meta) WithTx(ctx context.Context, do func(ctx context.Context, tx xdb.TxCore) error) error {
	te, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	return xdb.WithTx(ctx, te, func(ctx context.Context, tx xdb.TxCore) error {
		return do(ctx, te)
	})
}

func (m *Meta) checkWriteType(ctx context.Context, tx xdb.TxCore) error {
	orm := m.orm(tx)
	orm.SelectFields("dt")

	old, found, err := orm.First(ctx, "k=?", m.Key)
	if err != nil {
		return err
	}
	if found {
		if old.DataType == m.DataType {
			return nil
		}
		return fmt.Errorf("canot write %s on type %s", m.DataType.String(), old.DataType.String())
	}
	now := time.Now().Unix()
	data := MetaModel{
		Key:      m.Key,
		DataType: m.DataType,
		Created:  now,
		Updated:  now,
	}
	return orm.Insert(ctx, data)
}

func (m *Meta) load(ctx context.Context, tx xdb.TxCore) (MetaModel, error) {
	v, _, err := m.loadExists(ctx, tx)
	return v, err
}

func (m *Meta) loadExists(ctx context.Context, tx xdb.TxCore) (MetaModel, bool, error) {
	orm := m.orm(tx)
	value, found, err := orm.First(ctx, "k=?", m.Key)
	if err != nil {
		return MetaModel{}, false, err
	}
	if found {
		if value.DataType == m.DataType {
			return value, true, nil
		}
		return MetaModel{}, false, fmt.Errorf("canot load %s on type %s", m.DataType.String(), value.DataType.String())
	}
	return MetaModel{
		Key:      m.Key,
		DataType: m.DataType,
	}, false, nil
}

func (m *Meta) checkReadType(ctx context.Context, tx xdb.TxCore) (bool, error) {
	orm := m.orm(tx)
	orm.SelectFields("dt")
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
