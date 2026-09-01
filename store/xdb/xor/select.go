package xor

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xerror"
)

// First 使用 select xx from table where xxx limit 1 查询满足条件的第一条数据
//
// 可通过 SetSelectFields、SetSelectIgnore 限制查询返回的字段
func (m *Model[T]) First(ctx context.Context, opts ...Option) (v T, ok bool, err error) {
	if m.err != nil {
		return v, false, m.err
	}
	opts = append(opts, Limit(1))
	cfg := m.cfg.mergeOnClone(opts...)
	where, args, err := cfg.getWhereArgs(0)
	if err != nil {
		return v, false, err
	}

	sqlStr := fmt.Sprintf(
		"SELECT %s FROM %s %s",
		cfg.getSelectFields(),
		m.dialect.QuoteIdentifier(m.table),
		where,
	)
	return xdb.QueryOne[T](ctx, m.client, sqlStr, args...)
}

// GetFirst 查询满足添加的首条数据，若没有则返回错误：xerror.NotFound
func (m *Model[T]) GetFirst(ctx context.Context, opts ...Option) (v T, err error) {
	value, found, err := m.First(ctx, opts...)
	if err != nil {
		return v, err
	}
	if found {
		return value, nil
	}
	return v, xerror.NotFound
}

// FindByPK 使用主键查找数据
//
// 需要在 tag 里有 primaryKey 属性: 如 ID int64 `db:"id,pk"`
//
//	可通过 SetSelectFields、SetSelectIgnore 限制查询返回的字段
//
// 若查询不到，会返回: zero,false,nil
func (m *Model[T]) FindByPK(ctx context.Context, v T) (nv T, ok bool, err error) {
	return m.First(ctx, WhereByPK(v))
}

// GetByPK 使用主键查找数据,若数据查询不到，会返回 error：xerror.NotFound
//
// 需要在 tag 里有 primaryKey 属性: 如 ID int64 `db:"id,pk"`
//
//	可通过 SetSelectFields、SetSelectIgnore 限制查询返回的字段
func (m *Model[T]) GetByPK(ctx context.Context, v T) (nv T, err error) {
	value, found, err := m.FindByPK(ctx, v)
	if err != nil {
		return nv, err
	}
	if found {
		return value, nil
	}
	return nv, fmt.Errorf("%w: cond=%v", xerror.NotFound, v)
}

// List 查询并返回满足条件的数据
func (m *Model[T]) List(ctx context.Context, opts ...Option) ([]T, error) {
	var result []T
	for item, err := range m.ListIter(ctx, opts...) {
		if err != nil {
			return result, err
		}
		result = append(result, item)
	}
	return result, nil
}

// ListIter 查询满足条件的数据并返回一个迭代器。数据是流式从数据库返回的。
func (m *Model[T]) ListIter(ctx context.Context, opts ...Option) iter.Seq2[T, error] {
	cfg := m.cfg.mergeOnClone(opts...)
	return func(yield func(T, error) bool) {
		var zero T
		if m.err != nil {
			yield(zero, m.err)
			return
		}

		where, args, err := cfg.getWhereArgs(0)
		if err != nil {
			yield(zero, err)
			return
		}

		sqlStr := fmt.Sprintf(
			"SELECT %s FROM %s %s",
			cfg.getSelectFields(),
			m.dialect.QuoteIdentifier(m.table),
			where,
		)

		for item, err := range xdb.QueryManyIter[T](ctx, m.client, sqlStr, args...) {
			if !yield(item, err) {
				return
			}
		}
	}
}

func (m *Model[T]) Count(ctx context.Context, field string, opts ...Option) (num int64, err error) {
	if m.err != nil {
		return 0, m.err
	}
	if field == "" {
		field = "*"
	} else if field != "*" && !strings.ContainsRune(field, ' ') {
		field = m.dialect.QuoteIdentifier(field)
	}
	return m.doCount(ctx, field, opts...)
}

func (m *Model[T]) doCount(ctx context.Context, field string, opts ...Option) (num int64, err error) {
	cfg := m.cfg.Clone().merge(opts...)
	where, args, err := cfg.getWhereArgs(0)
	if err != nil {
		return 0, err
	}
	sqlStr := fmt.Sprintf("SELECT count(%s) from %s %s",
		field,
		m.dialect.QuoteIdentifier(m.table),
		where,
	)

	return xdb.Count(ctx, m.client, sqlStr, args...)
}

// ListPage 分页查询，适应于数据量不太大的场景
func (m *Model[T]) ListPage(ctx context.Context, page int, size int, opts ...Option) (xdb.Pagination, []xdb.PageRecord[T], error) {
	if m.err != nil {
		return xdb.Pagination{}, nil, m.err
	}
	if size < 1 {
		return xdb.Pagination{}, nil, fmt.Errorf("invalid size=%d", size)
	}
	total, err := m.doCount(ctx, "*", opts...)
	if err != nil {
		return xdb.Pagination{}, nil, err
	}

	page = max(page, 1) // 最小值为 1

	info := xdb.Pagination{
		TotalRecords: int(total),
		PageSize:     size,
		PageIndex:    page,
	}

	offset := (page - 1) * size
	if int64(offset) >= total {
		return info, nil, nil
	}
	opts = append(opts, LimitOffset(size, offset))
	result, err := m.List(ctx, opts...)
	if err != nil {
		return info, nil, err
	}
	items := xdb.ToPageRecords(result, (page-1)*size)
	return info, items, nil
}
