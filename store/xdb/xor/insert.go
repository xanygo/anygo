package xor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/store/xdb/internal/encoder"
)

// Insert 基本的 Insert 功能
func (m *Model[T]) Insert(ctx context.Context, v T, opts ...Option) error {
	if m.err != nil {
		return m.err
	}
	cfg := m.cfg.mergeOnClone(opts...)
	kv, err := m.getEncoder(encoder.ActionInsert, cfg).Encode(v)
	if err != nil {
		return err
	}
	if len(kv) == 0 {
		return fmt.Errorf("no columns found in: %T", v)
	}

	qcols := make([]string, 0, len(kv))
	args := make([]any, 0, len(kv))
	for field, value := range kv {
		qcols = append(qcols, m.dialect.QuoteIdentifier(field))
		args = append(args, value)
	}

	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		m.dialect.QuoteIdentifier(m.table),
		strings.Join(qcols, ", "),
		m.dialect.PlaceholderList(len(kv), 1),
	)
	_, err = xdb.Exec(ctx, m.client, sqlStr, args...)
	return err
}

// InsertReturningID 写入一条新数据,并返回 int 类型的主键 ID
//
// 若没有主键或者数据库不支持 LastInsertId 或者 Returning，会返回 0
func (m *Model[T]) InsertReturningID(ctx context.Context, v T, opts ...Option) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	cfg := m.cfg.mergeOnClone(opts...)
	kv, err := m.getEncoder(encoder.ActionInsert, cfg).Encode(v)
	if err != nil {
		return 0, err
	}
	if len(kv) == 0 {
		return 0, errors.New("no columns")
	}

	qcols := make([]string, 0, len(kv))
	args := make([]any, 0, len(kv))
	for k, v := range kv {
		qcols = append(qcols, m.dialect.QuoteIdentifier(k))
		args = append(args, v)
	}

	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		m.dialect.QuoteIdentifier(m.table),
		strings.Join(qcols, ", "),
		m.dialect.PlaceholderList(len(kv), 1),
	)
	sli := m.dialect.SupportLastInsertId()
	if !sli && m.dialect.SupportReturning() {
		rd, ok := m.dialect.(dbtype.ReturningDialect)
		if ok && len(m.pk) == 1 && m.pk[0].AutoIncrement {
			sqlStr += " " + rd.ReturningClause(m.pk[0].Name)
			return m.execReturning(ctx, sqlStr, args...)
		}
	}
	ret, err := xdb.Exec(ctx, m.client, sqlStr, args...)
	if err != nil {
		return 0, err
	}
	if sli {
		return ret.LastInsertId()
	}
	return 0, nil
}

func (m *Model[T]) execReturning(ctx context.Context, sql string, args ...any) (int64, error) {
	var id int64
	err := m.client.QueryRowContext(ctx, sql, args...).Scan(&id)
	return id, err
}

func (m *Model[T]) InsertBatch(ctx context.Context, items []T, opts ...Option) error {
	if m.err != nil {
		return m.err
	}
	if len(items) == 0 {
		return errors.New("no values")
	}
	cfg := m.cfg.mergeOnClone(opts...)
	values, err := m.getEncoder(encoder.ActionInsert, cfg).EncodeBatch(items...)
	if err != nil {
		return err
	}
	cols := xmap.Keys(values[0])
	if len(cols) == 0 {
		return errors.New("no columns")
	}

	qCols := xslice.MapFunc(cols, m.dialect.QuoteIdentifier)

	valuePlaceHolders := make([]string, len(values))
	for i := range len(values) {
		valuePlaceHolders[i] = "(" + m.dialect.PlaceholderList(len(cols), i*len(cols)+1) + ")"
	}

	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		m.dialect.QuoteIdentifier(m.table),
		strings.Join(qCols, ","),
		strings.Join(valuePlaceHolders, ", "),
	)

	vals := make([]any, 0, len(values)*len(cols))
	for _, item := range values {
		for _, col := range cols {
			vals = append(vals, item[col])
		}
	}

	_, err = xdb.Exec(ctx, m.client, sqlStr, vals...)
	return err
}

// Upsert 批量 insert or update
//
// 输入参数：
//
//	conflictCols: 冲突字段名，可选，若为空则，自动读取 pk 字段。
//	updateCols: 若冲突发生，执行更新的字段列表，可选。若为空，则冲突发生后，该条数据丢弃。
//	values: 数据列表，必填
//
//	返回值：受影响条数，错误
func (m *Model[T]) Upsert(ctx context.Context, conflictCols []string, updateCols []string, values ...T) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if len(values) == 0 {
		return 0, errors.New("no values")
	}
	dup, ok := m.dialect.(dbtype.UpsertDialect)
	if !ok {
		return 0, fmt.Errorf("dialect (%T) is not UpsertDialect", m.dialect)
	}

	if len(conflictCols) == 0 {
		conflictCols = m.pk.Names()
	}
	if len(conflictCols) == 0 {
		return 0, errors.New("empty conflict columns list")
	}

	kvSlice, err := m.getEncoder(encoder.ActionInsert, m.cfg).EncodeBatch(values...)
	if err != nil {
		return 0, err
	}
	cols := xmap.Keys(kvSlice[0])
	if miss, ok := xslice.AllContains(cols, updateCols); !ok {
		return 0, fmt.Errorf("invalid updateCols: %q not in %q", miss, cols)
	}
	args := make([]any, 0, len(values)*len(cols))
	for _, item := range kvSlice {
		for _, col := range cols {
			args = append(args, item[col])
		}
	}
	sqlStr := dup.UpsertSQL(m.table, len(values), cols, conflictCols, updateCols, nil)
	ret, err := xdb.Exec(ctx, m.client, sqlStr, args...)
	return xdb.RowsAffected(ret, err)
}

// UpsertByGroup 根据预定义的字段分组执行 Upsert
//
// conflictGroup: 冲突的主键分组名称，若为空则使用主键字段
// updateGroup: 冲突后更新字段分组名称。若值为空则冲突后不更新（丢弃）
//
// conflictGroup 和 updateGroup 都可以使用英文逗号连接多个 group
//
//	比如：
//
//	type User struct{
//		ID       string    `db:"id,pk"`
//		Sign     string    `db:"sign,unique_index=uniq_sign,group=sign"`
//		Name     string    `db:"name,group=update"`
//		Class    int       `db:"class,group=update"`
//		Age      int       `db:"age,group=update"`
//	 	Updated  time.Time `db:"updated,auto=Updated,group=update"`
//	}
//
//	m.UpsertByGroup(ctx,"sign","update",user1)
func (m *Model[T]) UpsertByGroup(ctx context.Context, conflictGroup string, updateGroup string, values ...T) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	conflictCols := m.schema.FilterByGroup(conflictGroup).Names()
	if conflictGroup != "" && len(conflictCols) == 0 {
		return 0, fmt.Errorf("conflict group %q column list is empty", conflictGroup)
	}
	updateCols := m.schema.FilterByGroup(updateGroup).Names()
	if updateGroup != "" && len(updateCols) == 0 {
		return 0, fmt.Errorf("update group %q column list is empty", updateGroup)
	}
	return m.Upsert(ctx, conflictCols, updateCols, values...)
}
