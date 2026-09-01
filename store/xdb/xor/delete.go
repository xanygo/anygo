package xor

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/store/xdb"
)

// Delete 执行 delete 语句，必须通过 Option传递删除条件
func (m *Model[T]) Delete(ctx context.Context, opts ...Option) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	cfg := m.cfg.mergeOnClone(opts...)
	where, args, err := cfg.getWhereArgs(0)
	if err != nil {
		return 0, err
	}
	sqlStr := fmt.Sprintf(
		"DELETE FROM %s %s",
		m.dialect.QuoteIdentifier(m.table),
		where,
	)
	ret, err := xdb.Exec(ctx, m.client, sqlStr, args...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

// DeleteByPK 使用主键删除数据
//
// 需要在 tag 里有 primaryKey 属性: 如 ID int64 `db:"id,pk"`
func (m *Model[T]) DeleteByPK(ctx context.Context, v T) (int64, error) {
	return m.Delete(ctx, WhereByPK(v))
}
