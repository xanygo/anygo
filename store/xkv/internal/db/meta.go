package db

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/xanygo/anygo/ds/xcast"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/xor"
	"github.com/xanygo/anygo/store/xkv/internal"
)

// KeyHash 计算字符串的 hash 值
// 由于部分数据库（如mssql）的索引长度限制，同时为了让查询效率更高。
// 所以对于字符串类型的字段，分别存储 hash 值和原文，给 hash 值添加索引用于查询
func KeyHash(str string) [32]byte {
	return sha256.Sum256([]byte(str))
}

func keyHashBytes(str string) []byte {
	h := sha256.Sum256([]byte(str))
	return h[:]
}

type MetaModel struct {
	ID       int64             `db:"id,pk,auto_inc"`
	TypeID   uint32            `db:"t,unique_index=t_k[1]"` // value 实际类型签名
	KeyHash  [32]byte          `db:"k,unique_index=t_k[2]"` // key 的 hash
	KeyRaw   string            `db:"k_raw"`                 // 原始的 key
	DataType internal.DataType `db:"dt"`
	Created  int64             `db:"c"`
	Updated  int64             `db:"u"`
	Meta     map[string]any    `db:"meta,codec=json"`
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
	TypeID   uint32   // 数据类型的签名
	KeyRaw   string   // 原始的 KeyRaw
	KeyHash  [32]byte // key 的 hash 值
	DataType internal.DataType
}

func (m *Meta) GetTable() string {
	if m.Table == "" {
		return "xkv_meta"
	}
	return m.Table
}

func (m *Meta) orm(tx xdb.DBCore) *xor.Model[MetaModel] {
	orm := xor.New[MetaModel](tx)
	orm.Table(m.GetTable())
	return orm
}

func (m *Meta) delete(ctx context.Context, tx xdb.DBCore) error {
	_, err := m.orm(tx).Delete(ctx, xor.Where("t=? and k=?", m.TypeID, m.KeyHash[:]))
	return err
}

func (m *Meta) save(ctx context.Context, tx xdb.DBCore, data MetaModel) error {
	now := time.Now().UnixNano()
	if data.Created == 0 {
		data.Created = now
	}
	if data.Updated == 0 {
		data.Updated = now
	}
	data.ID = 0 // k 是唯一主键，若upsert 时 id!=0, mssql 会报错
	_, err := m.orm(tx).Upsert(ctx, []string{"t", "k"}, []string{"meta", "u"}, data)
	return err
}

func (m *Meta) WithWriteTx(ctx context.Context, do func(ctx context.Context, tx xdb.DBCore) error) error {
	return m.WithTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		if err1 := m.checkWriteType(ctx, tx); err1 != nil {
			return err1
		}
		return do(ctx, tx)
	})
}

func (m *Meta) WithReadTx(ctx context.Context, do func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error) error {
	return m.WithTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		found, err1 := m.checkReadType(ctx, tx)
		if err1 != nil {
			return err1
		}
		return do(ctx, tx, found)
	})
}

func (m *Meta) WithTx(ctx context.Context, do func(ctx context.Context, tx xdb.DBCore) error) error {
	te, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	return xdb.WithTx(ctx, te, func(ctx context.Context, tx xdb.DBCore) error {
		return do(ctx, te)
	})
}

func (m *Meta) checkWriteType(ctx context.Context, tx xdb.DBCore) error {
	orm := m.orm(tx)

	old, found, err := orm.First(ctx, xor.Where("t=? and k=?", m.TypeID, m.KeyHash[:]), xor.Columns("dt"))
	if err != nil {
		return err
	}
	if found {
		if old.DataType == m.DataType {
			return nil
		}
		return fmt.Errorf("canot write %s on type %s", m.DataType.String(), old.DataType.String())
	}
	now := time.Now().UnixNano()
	data := MetaModel{
		TypeID:   m.TypeID,
		KeyHash:  m.KeyHash,
		KeyRaw:   m.KeyRaw,
		DataType: m.DataType,
		Created:  now,
		Updated:  now,
	}
	return orm.Insert(ctx, data)
}

func (m *Meta) load(ctx context.Context, tx xdb.DBCore) (MetaModel, error) {
	v, _, err := m.loadExists(ctx, tx)
	return v, err
}

func (m *Meta) loadExists(ctx context.Context, tx xdb.DBCore) (MetaModel, bool, error) {
	orm := m.orm(tx)
	value, found, err := orm.First(ctx, xor.Where("t=? and k=?", m.TypeID, m.KeyHash[:]))
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
		TypeID:   m.TypeID,
		KeyRaw:   m.KeyRaw,
		KeyHash:  m.KeyHash,
		DataType: m.DataType,
	}, false, nil
}

func (m *Meta) checkReadType(ctx context.Context, tx xdb.DBCore) (bool, error) {
	orm := m.orm(tx)
	value, found, err := orm.First(ctx, xor.Where("t=? and k=?", m.TypeID, m.KeyHash[:]), xor.Columns("dt"))
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
