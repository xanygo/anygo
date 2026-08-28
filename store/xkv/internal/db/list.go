package db

import (
	"context"
	"io"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv"
)

type ListModel struct {
	ID      int64    `db:"id,auto_inc,pk"`
	KeyHash [32]byte `db:"k,unique_index=k_i[1]"`
	Index   int64    `db:"idx,unique_index=k_i[2]"`
	KeyRaw  string   `db:"k_raw"`
	Value   string   `db:"v"`
	Created int64    `db:"c"`
}

var _ xkv.List[string] = (*List)(nil)

type List struct {
	Table string
	Meta  *Meta
}

func (l *List) GetTable() string {
	if l.Table != "" {
		return l.Table
	}
	return "xkv_list"
}

func (l *List) deleteWithKey(ctx context.Context, tx xdb.TxCore) error {
	orm := xdb.NewMode[ListModel](tx)
	orm.Table(l.GetTable())
	_, err := orm.Delete(ctx, "k=?", l.Meta.KeyHash[:])
	return err
}

func (l *List) xxPush(ctx context.Context, field string, dealt int64, values ...string) (num int64, err error) {
	if len(values) == 0 {
		return 0, nil
	}
	now := time.Now().UnixNano()
	err = l.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		meta, err1 := l.Meta.load(ctx, tx)
		if err1 != nil {
			return err1
		}
		var items []ListModel
		for _, value := range values {
			var idx int64
			meta, idx = meta.incr(field, dealt)
			item := ListModel{
				KeyHash: l.Meta.KeyHash,
				KeyRaw:  l.Meta.KeyRaw,
				Value:   value,
				Index:   idx,
				Created: now,
			}
			items = append(items, item)
		}
		err1 = l.Meta.save(ctx, tx, meta)
		if err1 != nil {
			return err1
		}
		orm := xdb.NewMode[ListModel](tx)
		orm.Table(l.GetTable())
		_, err1 = orm.InsertBatch(ctx, items...)
		if err1 != nil {
			return err1
		}
		num, err1 = orm.Count(ctx, "*", "k=?", l.Meta.KeyHash[:])
		return err1
	})
	return num, err
}

func (l *List) LPush(ctx context.Context, values ...string) (num int64, err error) {
	return l.xxPush(ctx, "left-idx", -1, values...)
}

func (l *List) RPush(ctx context.Context, values ...string) (int64, error) {
	return l.xxPush(ctx, "right-idx", 1, values...)
}

func (l *List) lPopXX(ctx context.Context, orderBy string) (value string, found bool, err error) {
	err = l.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		_, err1 := l.Meta.load(ctx, tx)
		if err1 != nil {
			return err1
		}
		orm := xdb.NewMode[ListModel](tx)
		orm.Table(l.GetTable())
		orm.SetSelectFields("v", "idx")

		v, ok, err2 := orm.First(ctx, "k=? order by idx "+orderBy, l.Meta.KeyHash[:])
		if err2 != nil || !ok {
			return err2
		}
		value = v.Value
		found = true
		_, err2 = orm.Delete(ctx, "k=? and idx=?", l.Meta.KeyHash[:], v.Index)
		if err2 != nil {
			return err2
		}
		return l.checkExists(ctx, orm)
	})
	return value, found, err
}

func (l *List) LPop(ctx context.Context) (value string, found bool, err error) {
	return l.lPopXX(ctx, "asc")
}

func (l *List) RPop(ctx context.Context) (value string, found bool, err error) {
	return l.lPopXX(ctx, "desc")
}

func (l *List) lPopNXX(ctx context.Context, count int, orderBy string) (result []string, err error) {
	err = l.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		_, err1 := l.Meta.load(ctx, tx)
		if err1 != nil {
			return err1
		}
		orm := xdb.NewMode[ListModel](tx)
		orm.Table(l.GetTable())
		orm.SetSelectFields("v", "idx").Limit(count)
		items, err2 := orm.List(ctx, "k=? order by idx "+orderBy, l.Meta.KeyHash[:])
		if err2 != nil || len(items) == 0 {
			return err2
		}
		idxList := make([]int64, 0, len(items))
		for _, item := range items {
			result = append(result, item.Value)
			idxList = append(idxList, item.Index)
		}
		cond := xdb.Condition{}
		cond.And("k=?", l.Meta.KeyHash[:])
		cond.AndInFmt("idx in (%s)", xslice.ToAnys(idxList))
		where, args, err3 := cond.Build()
		if err3 != nil {
			return err3
		}
		_, err4 := orm.Delete(ctx, where, args...)
		if err4 != nil {
			return err4
		}
		return l.checkExists(ctx, orm)
	})
	return result, err
}

func (l *List) LPopN(ctx context.Context, count int) (result []string, err error) {
	return l.lPopNXX(ctx, count, "asc")
}

func (l *List) RPopN(ctx context.Context, count int) (result []string, err error) {
	return l.lPopNXX(ctx, count, "desc")
}

// checkExists 检查 key 是否还存在，若不存在，则删除 meta
func (l *List) checkExists(ctx context.Context, orm *xdb.Model[ListModel]) error {
	orm = orm.Clone().Reset()
	orm.Table(l.GetTable())
	orm.SetSelectFields("c")

	_, found, err := orm.First(ctx, "k=?", l.Meta.KeyHash[:])
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return l.Meta.delete(ctx, orm.Client())
}

func (l *List) LRem(ctx context.Context, count int64, element string) (num int64, err error) {
	if count == 0 { // 删除全部
		err = l.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
			_, err1 := l.Meta.load(ctx, tx)
			if err1 != nil {
				return err1
			}
			orm := xdb.NewMode[ListModel](tx)
			orm.Table(l.GetTable())
			var err2 error
			num, err2 = orm.Delete(ctx, "k=? and v=?", l.Meta.KeyHash[:], element)
			if err2 != nil {
				return err2
			}
			return l.checkExists(ctx, orm)
		})
		return num, err
	}
	if count > 0 {
		// 从头部到尾部移除 count 个等于 element 的元素。
		err = l.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
			_, err1 := l.Meta.load(ctx, tx)
			if err1 != nil {
				return err1
			}
			orm := xdb.NewMode[ListModel](tx)
			orm.Table(l.GetTable())
			orm.Limit(int(count))
			var err2 error
			num, err2 = orm.Delete(ctx, "k=? and v=? order by idx asc", l.Meta.KeyHash[:], element)
			if err2 != nil {
				return err2
			}
			return l.checkExists(ctx, orm)
		})
		return num, err
	}

	// count < 0: 从尾部到头部移除 abs(count) 个等于 element 的元素。

	count = count * -1
	err = l.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		_, err1 := l.Meta.load(ctx, tx)
		if err1 != nil {
			return err1
		}
		orm := xdb.NewMode[ListModel](tx)
		orm.Table(l.GetTable())
		orm.Limit(int(count))
		var err2 error
		num, err2 = orm.Delete(ctx, "k=? and v=? order by idx desc", l.Meta.KeyHash[:], element)
		if err2 != nil {
			return err2
		}
		return l.checkExists(ctx, orm)
	})
	return num, err
}

func (l *List) Range(ctx context.Context, fn func(val string) bool) error {
	err := l.Meta.WithReadTx(ctx, func(as context.Context, tx xdb.TxCore, hasMeta bool) error {
		orm := xdb.NewMode[ListModel](tx)
		orm.Table(l.GetTable())
		orm.SetSelectFields("v")

		for item, err1 := range orm.ListIter(ctx, "k=?", l.Meta.KeyHash[:]) {
			if err1 != nil {
				return err1
			}
			if !fn(item.Value) {
				return io.EOF
			}
		}
		return nil
	})
	return err
}

func (l *List) LRange(ctx context.Context, fn func(val string) bool) error {
	err := l.Meta.WithReadTx(ctx, func(as context.Context, tx xdb.TxCore, hasMeta bool) error {
		orm := xdb.NewMode[ListModel](tx)
		orm.Table(l.GetTable())
		orm.SetSelectFields("v")

		for item, err1 := range orm.ListIter(ctx, "k=? order by idx asc", l.Meta.KeyHash[:]) {
			if err1 != nil {
				return err1
			}
			if !fn(item.Value) {
				return io.EOF
			}
		}
		return nil
	})
	return err
}

func (l *List) RRange(ctx context.Context, fn func(val string) bool) error {
	err := l.Meta.WithReadTx(ctx, func(as context.Context, tx xdb.TxCore, hasMeta bool) error {
		orm := xdb.NewMode[ListModel](tx)
		orm.Table(l.GetTable())
		orm.SetSelectFields("v")

		for item, err1 := range orm.ListIter(ctx, "k=? order by idx desc", l.Meta.KeyHash[:]) {
			if err1 != nil {
				return err1
			}
			if !fn(item.Value) {
				return io.EOF
			}
		}
		return nil
	})
	return err
}

func (l *List) LLen(ctx context.Context) (num int64, err error) {
	err = l.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xdb.NewMode[ListModel](tx)
		orm.Table(l.GetTable())
		var err1 error
		num, err1 = orm.Count(ctx, "*", "k=?", l.Meta.KeyHash[:])
		return err1
	})
	return num, err
}
