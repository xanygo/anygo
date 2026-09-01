package xor

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/internal/encoder"
	"github.com/xanygo/anygo/xerror"
)

// Update 执行 update 语句
func (m *Model[T]) Update(ctx context.Context, v T, opts ...Option) (int64, error) {
	cfg := m.cfg.Clone().merge(opts...)
	return m.doUpdate(ctx, v, cfg)
}

func (m *Model[T]) doUpdate(ctx context.Context, v T, cfg *config) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	kv, err := m.getEncoder(encoder.ActionUpdate, cfg).Encode(v)
	if err != nil {
		return 0, err
	}
	return m.doUpdateMap(ctx, kv, cfg)
}

func (m *Model[T]) doUpdateMap(ctx context.Context, kv map[string]any, cfg *config) (int64, error) {
	cols := xmap.Keys(kv)

	assigns := make([]string, 0, len(cols))
	values := make([]any, 0, len(cfg.whereArgs)+len(cols))
	for _, col := range cols {
		str := fmt.Sprintf(`%s=%s`, m.dialect.QuoteIdentifier(col), m.dialect.BindVar(len(assigns)+1))
		assigns = append(assigns, str)
		values = append(values, kv[col])
	}

	if len(assigns) == 0 {
		return 0, errors.New("no update values")
	}
	where, args, err := cfg.getWhereArgs(len(assigns))
	if err != nil {
		return 0, err
	}

	sqlStr := fmt.Sprintf(
		"UPDATE %s SET %s %s",
		m.dialect.QuoteIdentifier(m.table),
		strings.Join(assigns, ", "),
		where,
	)
	values = append(values, args...)

	ret, err := xdb.Exec(ctx, m.client, sqlStr, values...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

// UpdateByPK 使用主键更新数据
//
// 需要在 tag 里有 primaryKey 属性: 如 ID int64 `db:"id,pk"`。支持联合主键。
func (m *Model[T]) UpdateByPK(ctx context.Context, v T) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	cfg := m.cfg.mergeOnClone(WhereByPK(v))

	// 主键字段不被更新，否则部分数据库（如 mssql ）会报错
	cfg.ignores = append(cfg.ignores, m.pk.Names()...)

	return m.doUpdate(ctx, v, cfg)
}

// Modify 增量更新新一条数据
// old: 更新前的旧数据
//
// update: 更新处理，返回的数据会和 old 做 diff，然后只更新 diff 字段。
// 若返回 error 是 xerror.SkipOne 或 xerror.SkipAll 则跳过。其他 error 则直接返回
func (m *Model[T]) Modify(ctx context.Context, old T, update func(nv T) (T, error), opts ...Option) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	clonedValue, err := zreflect.Clone(old)
	if err != nil {
		return 0, err
	}
	nv, err := update(clonedValue)
	if err != nil {
		if xerror.IsSkip(err) {
			return 0, nil
		}
		return 0, err
	}
	return m.UpdateDiff(ctx, old, nv, opts...)
}

// UpdateDiff 增量更新数据
//
// 应采用 Option 传递更新条件，若没有传递则将 old 数据的主键当作更新条件
func (m *Model[T]) UpdateDiff(ctx context.Context, old T, newValue T, opts ...Option) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if reflect.DeepEqual(old, newValue) {
		return 0, nil
	}

	cfg := m.cfg.Clone().merge(opts...)
	if cfg.where == "" {
		cfg.merge(WhereByPK(old))
	}

	enc := m.getEncoder(encoder.ActionUpdate, cfg)
	diff, err := enc.Diff(newValue, old)
	if err != nil || len(diff) == 0 {
		return 0, err
	}
	return m.doUpdateMap(ctx, diff, cfg)
}

// ModifyFirstByPK 使用主键查找，然后更新数据。若查找不到会返回错误
// q: 查询条件，主键字段必须有值。若主键字段有多个，但是只给部分字段赋值，可能会导致多条数据被更新
func (m *Model[T]) ModifyFirstByPK(ctx context.Context, q T, update func(nv T) (T, error)) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.ModifyFirst(ctx, update, WhereByPK(q))
}

// ModifyFirst 查找数据然后更新，若查找不到会返回错误
//
// 注意：若 where 条件返回多条，会查询第一条数据，并以此位基础更新所有数据
//
// update: 数据更新方法。若返回 error 是 xerror.SkipOne 或 xerror.SkipAll 则跳过。其他 error 则直接返回
func (m *Model[T]) ModifyFirst(ctx context.Context, update func(nv T) (T, error), opts ...Option) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	var num int64
	err := xdb.BeginTx(ctx, m.client, nil, func(ctx context.Context, tx xdb.DBCore) error {
		m1 := m.Clone()
		m1.client = tx
		old, found, err1 := m1.First(ctx, opts...)
		if err1 != nil {
			return err1
		}
		if !found {
			return xerror.NotFound
		}
		num, err1 = m1.Modify(ctx, old, update, opts...)
		return err1
	})
	return num, err
}

// ModifyEach 逐条更新满足条件的每一条数据。
//
// 由于最终执行更新是，where 添加中使用的是主键作为条件，所以数据 T 需要定义主键
//
// update: 数据更新方法。若返回 error 是 xerror.SkipOne,则跳过此条数据，若是 xerror.SkipAll 则跳过所有。其他 error 则直接返回
func (m *Model[T]) ModifyEach(ctx context.Context, update func(nv T) (T, error), opts ...Option) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	m1 := m.New()
	var num int64
	for item, err := range m1.ListIter(ctx, opts...) {
		if err != nil {
			return num, err
		}
		where := WhereByPK(item) // 在 update 之前构建好，避免 update 修改了数据
		newValue, err2 := update(item)
		if err2 != nil {
			if errors.Is(err2, xerror.SkipOne) {
				continue
			}
			if errors.Is(err2, xerror.SkipAll) {
				break
			}
			return num, err2
		}
		n, err3 := m1.UpdateDiff(ctx, item, newValue, where)
		if err3 != nil {
			return num, err3
		}
		num += n
	}

	return num, nil
}
