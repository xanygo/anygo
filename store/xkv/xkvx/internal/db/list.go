package db

import (
	"context"
	"io"
	"time"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv"
)

type ListModel struct {
	Table   string
	Key     string `db:"k,unique_index:idx_k_i"`
	Index   int64  `db:"idx,unique_index:idx_k_i"`
	Value   string `db:"v"`
	Created int64  `db:"c"`
}

func (dt ListModel) TableName() string {
	if dt.Table == "" {
		return "xkv_list"
	}
	return dt.Table
}

func (dt ListModel) deleteWithKey(ctx context.Context, tx xdb.TxCore) error {
	ms := xdb.NewMode[ListModel](tx)
	_, err := ms.Delete(ctx, "k=?", dt.Key)
	return err
}

var _ xkv.List[string] = (*List)(nil)

type List struct {
	Table string
	Key   string
	Meta  MetaModel
}

func (l *List) xxPush(ctx context.Context, field string, dealt int64, values ...string) (num int64, err error) {
	if len(values) == 0 {
		return 0, nil
	}
	now := time.Now().Unix()
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
				Table:   l.Table,
				Key:     l.Key,
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
		_, err1 = orm.InsertBatch(ctx, items...)
		if err1 == nil {
			num = int64(len(items))
		}
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

func (l *List) LPop(ctx context.Context) (value string, found bool, err error) {
	err = l.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		_, err1 := l.Meta.load(ctx, tx)
		if err1 != nil {
			return err1
		}
		orm := xdb.NewMode[ListModel](tx)
		v, ok, err2 := orm.First(ctx, "k=? order by idx asc", l.Key)
		if err2 != nil || !ok {
			return err2
		}
		value = v.Value
		found = true
		_, err2 = orm.Delete(ctx, "k=? and idx=?", v.Key, v.Index)
		if err2 != nil {
			return err2
		}
		return l.checkExists(ctx, orm)
	})
	return value, found, err
}

// checkExists 检查 key 是否还存在，若不存在，则删除 meta
func (l *List) checkExists(ctx context.Context, orm *xdb.Model[ListModel]) error {
	orm = orm.Clone().Reset().OnlyFields("c")
	_, found, err := orm.First(ctx, "k=?", l.Key)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return l.Meta.delete(ctx, orm.Client())
}

func (l *List) RPop(ctx context.Context) (value string, found bool, err error) {
	err = l.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
		_, err1 := l.Meta.load(ctx, tx)
		if err1 != nil {
			return err1
		}
		orm := xdb.NewMode[ListModel](tx)
		orm.OnlyFields("v")
		v, ok, err2 := orm.First(ctx, "k=? order by idx desc", l.Key)
		if err2 != nil || !ok {
			return err2
		}
		value = v.Value
		found = true
		_, err2 = orm.Delete(ctx, "k=? and idx=?", v.Key, v.Index)
		if err2 != nil {
			return err2
		}
		return l.checkExists(ctx, orm)
	})
	return value, found, err
}

func (l *List) LRem(ctx context.Context, count int64, element string) (num int64, err error) {
	if count == 0 { // 删除全部
		err = l.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.TxCore) error {
			_, err1 := l.Meta.load(ctx, tx)
			if err1 != nil {
				return err1
			}
			orm := xdb.NewMode[ListModel](tx)
			var err2 error
			num, err2 = orm.Delete(ctx, "k=? and v=?", l.Key, element)
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
			orm.Limit(int(count))
			var err2 error
			num, err2 = orm.Delete(ctx, "k=? and v=? order by idx asc", l.Key, element)
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
		orm.Limit(int(count))
		var err2 error
		num, err2 = orm.Delete(ctx, "k=? and v=? order by idx desc", l.Key, element)
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
		orm.OnlyFields("v")
		for item, err1 := range orm.ListIter(ctx, "k=?", l.Key) {
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
		orm.OnlyFields("v")
		for item, err1 := range orm.ListIter(ctx, "k=? order by idx asc", l.Key) {
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
		orm.OnlyFields("v")
		for item, err1 := range orm.ListIter(ctx, "k=? order by idx desc", l.Key) {
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
		var err1 error
		num, err1 = orm.Count(ctx, "*", "k=?", l.Key)
		return err1
	})
	return num, err
}
