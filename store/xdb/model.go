//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-10

package xdb

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbschema"
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/store/xdb/dialect"
	"github.com/xanygo/anygo/store/xdb/internal/encoder"
	"github.com/xanygo/anygo/xerror"
)

// HasTable 给 Model 使用的 struct 可以选择实现该接口，以自动读取数据库表名
type HasTable interface {
	TableName() string
}

// NewMode 生成一个 T 类型的 简单 ORM Model
// T 类型建议传入 struct 而不是 *struct，兼容性更好
func NewMode[T any](client HasDriver) *Model[T] {
	m := &Model[T]{
		client: client,
	}
	m.init()
	return m
}

// Model 轻量 ORM 实现，已实现数据模型常用的增删改查功能
//
// 使用此 Model 的 where 条件:
//   - 统一使用 ? 占位符，在执行前，会将 ? 替换为方言的占位符
//   - where 中可以写 order by X:RAND(), 让结果随机排序，字符串 “X:RAND()” 会被替换为方言
type Model[T any] struct {
	dialect dbtype.Dialect
	client  HasDriver

	table         string
	limit, offset int

	upsertFields       []string // insert, update 的字段列表
	upsertIgnoreFields []string // insert, update 忽略的字段列表

	selectFields       string   // select 查询的字段列表
	selectIgnoreFields []string // 查询时要忽略的字段列表。当 selectFields 为空时才生效

	schema *dbtype.TableSchema
	pk     dbtype.ColumnSchemas // 可能为 nil

	err error
}

func (m *Model[T]) init() {
	m.dialect, m.err = dialect.Find(m.client.Driver())
	if m.err == nil {
		var zero T
		m.schema, m.err = dbschema.Schema(m.dialect, zero)
	}
	if m.schema != nil {
		m.table = m.schema.Table
		m.pk = m.schema.PKColumns()
	}
}

func (m *Model[T]) Client() HasDriver {
	return m.client
}

// Reset 重置 limit、offset、upsertFields、upsertIgnore、selectFields、selectIgnore 等属性
//
// Table 属性会保留
func (m *Model[T]) Reset() *Model[T] {
	m.limit = 0
	m.offset = 0

	m.upsertFields = nil
	m.upsertIgnoreFields = nil

	m.selectFields = ""
	m.selectIgnoreFields = nil
	return m
}

func (m *Model[T]) CloneBase() *Model[T] {
	return &Model[T]{
		dialect: m.dialect,
		client:  m.client,
		table:   m.table,
		pk:      slices.Clone(m.pk),
		schema:  m.schema,
		err:     m.err,
	}
}

func (m *Model[T]) Clone() *Model[T] {
	return &Model[T]{
		dialect: m.dialect,
		client:  m.client,
		table:   m.table,
		pk:      slices.Clone(m.pk),
		schema:  m.schema,
		err:     m.err,

		limit:  m.limit,
		offset: m.offset,

		upsertFields:       slices.Clone(m.upsertFields),
		upsertIgnoreFields: slices.Clone(m.upsertIgnoreFields),

		selectFields:       m.selectFields,
		selectIgnoreFields: slices.Clone(m.selectIgnoreFields),
	}
}

// SetSelectFields 设置查询字段列表，设置为空，则查询所有字段
func (m *Model[T]) SetSelectFields(fields ...string) *Model[T] {
	if len(fields) == 1 && fields[0] == "*" {
		m.selectFields = "*"
	} else if len(fields) == 0 {
		m.selectFields = ""
	} else {
		m.selectFields = strings.Join(xslice.MapFunc(fields, m.dialect.QuoteIdentifier), ",")
	}
	return m
}

// SetSelectIgnore 设置查询时忽略的字段列表，未设置 SetSelectFields 时生效
func (m *Model[T]) SetSelectIgnore(fields ...string) *Model[T] {
	m.selectIgnoreFields = fields
	return m
}

func (m *Model[T]) getSelectFields() (string, error) {
	if m.selectFields != "" {
		return m.selectFields, nil
	}

	fields := slices.Clone(m.schema.ColumnNames)

	if len(m.selectIgnoreFields) != 0 {
		fields = xslice.DeleteValue(fields, m.selectIgnoreFields...)
	}

	if len(fields) == 0 {
		return "*", nil
	}

	return strings.Join(xslice.MapFunc(fields, m.dialect.QuoteIdentifier), ","), nil
}

// SetUpsertFields 设置 insert、update 的字段列表，默认为空时，写入所有字段
func (m *Model[T]) SetUpsertFields(fields ...string) *Model[T] {
	m.upsertFields = fields
	return m
}

func (m *Model[T]) AppendUpsertFields(fields ...string) *Model[T] {
	m.upsertFields = append(m.upsertFields, fields...)
	return m
}

// SetUpsertIgnore 设置 insert 和 update 时候，需要忽略的字段，默认为空
func (m *Model[T]) SetUpsertIgnore(fields ...string) *Model[T] {
	m.upsertIgnoreFields = fields
	return m
}

func (m *Model[T]) AppendUpsertIgnore(fields ...string) *Model[T] {
	m.upsertIgnoreFields = append(m.upsertIgnoreFields, fields...)
	return m
}

// Table 设置表名，若 T 没有实现 HasTable 接口时，可通过此设置
func (m *Model[T]) Table(table string) *Model[T] {
	m.table = table
	return m
}

func (m *Model[T]) Limit(num int) *Model[T] {
	m.limit = num
	return m
}

func (m *Model[T]) Offset(num int) *Model[T] {
	m.offset = num
	return m
}

func (m *Model[T]) getEncoder(action encoder.Action) encoder.Encoder[T] {
	return encoder.Encoder[T]{
		Schema:       m.schema,
		Action:       action,
		Dialect:      m.dialect,
		OnlyFields:   m.upsertFields,
		IgnoreFields: m.upsertIgnoreFields,
	}
}

// Insert 基本的 Insert 功能
func (m *Model[T]) Insert(ctx context.Context, v T) error {
	if m.err != nil {
		return m.err
	}
	kv, err := m.getEncoder(encoder.ActionInsert).Encode(v)
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

	db, ok := m.client.(Execer)
	if !ok {
		return fmt.Errorf("client (%T) is not Execer", m.client)
	}
	_, err = Exec(ctx, db, sqlStr, args...)
	return err
}

// QuoteIdentifier 将标识符转义
func (m *Model[T]) QuoteIdentifier(name string) string {
	return m.dialect.QuoteIdentifier(name)
}

// InsertReturningID 写入一条新数据,并返回 int 类型的主键 ID
//
// 若没有主键或者数据库不支持 LastInsertId 或者 Returning，会返回 0
func (m *Model[T]) InsertReturningID(ctx context.Context, v T) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	kv, err := m.getEncoder(encoder.ActionInsert).Encode(v)
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
			return m.execReturning(ctx, m.client, sqlStr, args...)
		}
	}

	db, ok := m.client.(Execer)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not Execer", m.client)
	}
	ret, err := Exec(ctx, db, sqlStr, args...)
	if err != nil {
		return 0, err
	}
	if sli {
		return ret.LastInsertId()
	}
	return 0, nil
}

func (m *Model[T]) execReturning(ctx context.Context, client HasDriver, sql string, args ...any) (int64, error) {
	db, ok := client.(RowQuerier)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not RowQuerier", client)
	}

	var id int64
	err := db.QueryRowContext(ctx, sql, args...).Scan(&id)
	return id, err
}

func (m *Model[T]) InsertBatch(ctx context.Context, vs ...T) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if len(vs) == 0 {
		return 0, errors.New("no values")
	}
	values, err := m.getEncoder(encoder.ActionInsert).EncodeBatch(vs...)
	if err != nil {
		return 0, err
	}
	cols := xmap.Keys(values[0])
	if len(cols) == 0 {
		return 0, errors.New("no columns")
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

	db, ok := m.client.(Execer)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not Execer", m.client)
	}

	vals := make([]any, 0, len(values)*len(cols))
	for _, item := range values {
		for _, col := range cols {
			vals = append(vals, item[col])
		}
	}

	ret, err := Exec(ctx, db, sqlStr, vals...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
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
	db, ok := m.client.(Execer)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not Execer", m.client)
	}

	if len(conflictCols) == 0 {
		conflictCols = m.pk.Names()
	}
	if len(conflictCols) == 0 {
		return 0, errors.New("empty conflict columns list")
	}

	kvSlice, err := m.getEncoder(encoder.ActionInsert).EncodeBatch(values...)
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
	ret, err := Exec(ctx, db, sqlStr, args...)
	return RowsAffected(ret, err)
}

// UpsertByGroup 根据预定义的字段分组执行 Upsert
//
// conflictGroup: 冲突的主键分组名称
// updateGroup: 冲突后更新字段分组名称。若值为空，或者 分组对于的字段列表为空，则冲突后不更新（丢弃）
//
//	比如：
//
//	type User struct{
//		ID       string    `db:"id,pk"`
//		Sign     string    `db:"sign,unique_index=uniq_sign,group=sign"`
//		Name     string    `db:"name,group=update"`
//		Class    int       `db:"class,group=update"`
//		Age      int       `db:"age,group=update"`
//	 	Updated  time.Time `db:"updated,auto=Updated,group=*"`
//	}
//
//	m.UpsertByGroup(ctx,"sign","update",user1)
func (m *Model[T]) UpsertByGroup(ctx context.Context, conflictGroup string, updateGroup string, values ...T) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	conflictCols := m.schema.FilterByGroup(conflictGroup).Names()
	updateCols := m.schema.FilterByGroup(updateGroup).Names()
	return m.Upsert(ctx, conflictCols, updateCols, values...)
}

// Update 执行 update 语句
func (m *Model[T]) Update(ctx context.Context, v T, where string, args ...any) (int64, error) {
	return m.doUpdate(ctx, v, where, args...)
}

func (m *Model[T]) doUpdate(ctx context.Context, v T, where string, args ...any) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	kv, err := m.getEncoder(encoder.ActionUpdate).Encode(v)
	if err != nil {
		return 0, err
	}
	return m.doUpdateMap(ctx, kv, where, args...)
}

func (m *Model[T]) doUpdateMap(ctx context.Context, kv map[string]any, where string, args ...any) (int64, error) {
	cols := xmap.Keys(kv)

	assigns := make([]string, 0, len(cols))
	values := make([]any, 0, len(args)+len(cols))
	for _, col := range cols {
		str := fmt.Sprintf(`%s=%s`, m.dialect.QuoteIdentifier(col), m.dialect.BindVar(len(assigns)+1))
		assigns = append(assigns, str)
		values = append(values, kv[col])
	}

	if len(assigns) == 0 {
		return 0, errors.New("no update values")
	}
	var err error
	where, args, err = m.buildWhere(len(assigns), where, args)
	if err != nil {
		return 0, err
	}

	if len(where) == 0 || len(args) == 0 {
		return 0, errors.New("empty where clause")
	}

	sqlStr := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		m.dialect.QuoteIdentifier(m.table),
		strings.Join(assigns, ", "),
		where,
	)
	values = append(values, args...)

	db, ok := m.client.(Execer)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not Execer", m.client)
	}
	ret, err := Exec(ctx, db, sqlStr, values...)
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
	pkData, err := m.getEncoder(encoder.ActionSelect).PKNameAndValues(v)
	if err != nil {
		return 0, err
	}
	where, args, err := m.mapWhere(pkData)
	if err != nil {
		return 0, err
	}

	m1 := m.Clone()
	m1.AppendUpsertIgnore(xmap.Keys(pkData)...)
	return m1.doUpdate(ctx, v, where, args...)
}

// Modify 增量更新新一条数据
// old: 更新前的旧数据
//
// update: 更新处理，返回的数据会和 old 做 diff，然后只更新 diff 字段。
// 若返回 error 是 xerror.SkipOne 或 xerror.SkipAll 则跳过。其他 error 则直接返回
func (m *Model[T]) Modify(ctx context.Context, old T, update func(nv T) (T, error), where string, args ...any) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	nv, err := update(zreflect.Clone(old))
	if err != nil {
		if xerror.IsSkip(err) {
			return 0, nil
		}
		return 0, err
	}
	return m.UpdateDiff(ctx, old, nv, where, args...)
}

// UpdateDiff 增量更新数据
func (m *Model[T]) UpdateDiff(ctx context.Context, old T, newValue T, where string, args ...any) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if reflect.DeepEqual(old, newValue) {
		return 0, nil
	}
	enc := m.getEncoder(encoder.ActionUpdate)
	diff, err := enc.Diff(newValue, old)
	if err != nil || len(diff) == 0 {
		return 0, err
	}
	return m.doUpdateMap(ctx, diff, where, args...)
}

// ModifyFirstByPK 使用主键查找，然后更新数据。若查找不到会返回错误
// q: 查询条件，主键字段必须有值。若主键字段有多个，但是只给部分字段赋值，可能会导致多条数据被更新
func (m *Model[T]) ModifyFirstByPK(ctx context.Context, q T, update func(nv T) (T, error)) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	m1 := m.CloneBase()
	where, args, err := m1.pkWhereArgs(q, encoder.ActionSelect)
	if err != nil {
		return 0, err
	}

	old, found, err := m1.First(ctx, where, args...)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, xerror.NotFound
	}
	return m1.Modify(ctx, old, update, where, args...)
}

// ModifyFirst 查找数据然后更新，若查找不到会返回错误
//
// 注意：若 where 条件返回多条，会查询第一条数据，并以此位基础更新所有数据
//
// update: 数据更新方法。若返回 error 是 xerror.SkipOne 或 xerror.SkipAll 则跳过。其他 error 则直接返回
func (m *Model[T]) ModifyFirst(ctx context.Context, update func(nv T) (T, error), where string, args ...any) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	m1 := m.CloneBase()
	old, found, err := m1.First(ctx, where, args...)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, xerror.NotFound
	}
	return m1.Modify(ctx, old, update, where, args...)
}

// ModifyEach 逐条更新满足条件的每一条数据。
//
// 由于最终执行更新是，where 添加中使用的是主键作为条件，所以数据 T 需要定义主键
//
// update: 数据更新方法。若返回 error 是 xerror.SkipOne,则跳过此条数据，若是 xerror.SkipAll 则跳过所有。其他 error 则直接返回
func (m *Model[T]) ModifyEach(ctx context.Context, update func(nv T) (T, error), where string, args ...any) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	m1 := m.CloneBase()
	var num int64
	for item, err := range m1.ListIter(ctx, where, args...) {
		if err != nil {
			return num, err
		}
		where1, args1, err1 := m1.pkWhereArgs(item, encoder.ActionSelect)
		if err1 != nil {
			return num, err1
		}
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
		n, err3 := m.UpdateDiff(ctx, item, newValue, where1, args1...)
		if err3 != nil {
			return num, err3
		}
		num += n
	}

	return num, nil
}

// Delete 执行 delete 语句 （不受 Limit offset 影响）
func (m *Model[T]) Delete(ctx context.Context, where string, args ...any) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	var err error
	where, args, err = m.buildWhere(0, where, args)
	if err != nil {
		return 0, err
	}
	if len(where) == 0 || len(args) == 0 {
		return 0, errors.New("empty where clause")
	}
	sqlStr := fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		m.dialect.QuoteIdentifier(m.table),
		where,
	)
	db, ok := m.client.(Execer)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not Execer", m.client)
	}
	ret, err := Exec(ctx, db, sqlStr, args...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

// DeleteByPK 使用主键删除数据
//
// 需要在 tag 里有 primaryKey 属性: 如 ID int64 `db:"id,pk"`
func (m *Model[T]) DeleteByPK(ctx context.Context, v T) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	where, args, err := m.pkWhereArgs(v, encoder.ActionSelect)
	if err != nil {
		return 0, err
	}
	return m.Delete(ctx, where, args...)
}

func (m *Model[T]) mapWhere(data map[string]any) (string, []any, error) {
	if len(data) == 0 {
		return "", nil, errors.New("no primary key")
	}
	cond := &Condition{}
	for key, value := range data {
		cond.And(m.dialect.QuoteIdentifier(key)+"=?", value)
	}
	return cond.Build()
}

func (m *Model[T]) pkWhereArgs(q T, action encoder.Action) (string, []any, error) {
	pkData, err := m.getEncoder(action).PKNameAndValues(q)
	if err != nil {
		return "", nil, err
	}
	return m.mapWhere(pkData)
}

var orderByReg = regexp.MustCompile(`(?i)^order\s+by\s+`)

func (m *Model[T]) connectWhere(where string) string {
	where = strings.TrimSpace(where)
	if where == "" || orderByReg.MatchString(where) {
		return where
	}
	return " where " + where
}

func (m *Model[T]) buildWhere(indexStart int, where string, args []any) (string, []any, error) {
	args, err := m.getEncoder(encoder.ActionSelect).EncodeArgs(args...)
	if err != nil {
		return "", nil, err
	}

	// 将 ? 替换为方言的占位符，如 $1, $2 ...
	if m.dialect.BindVar(0) != "?" {
		var sb strings.Builder
		idx := 1
		for i := 0; i < len(where); i++ {
			if where[i] == '?' {
				sb.WriteString(m.dialect.BindVar(indexStart + idx))
				idx++
			} else {
				sb.WriteByte(where[i])
			}
		}
		where = sb.String()
	}

	// 将条件中的 RAND() 换成方言
	if strings.Contains(where, KWRand) {
		if dr := m.dialect.RandomOrder(); dr != KWRand {
			where = strings.ReplaceAll(where, KWRand, dr)
		}
	}

	return where, args, nil
}

// First 使用 select xx from table where xxx limit 1 查询满足条件的第一条数据
//
// 可通过 SetSelectFields、SetSelectIgnore 限制查询返回的字段
func (m *Model[T]) First(ctx context.Context, where string, args ...any) (v T, ok bool, err error) {
	if m.err != nil {
		return v, false, m.err
	}
	where, args, err = m.buildWhere(0, where, args)
	if err != nil {
		return v, false, err
	}

	field, err := m.getSelectFields()
	if err != nil {
		return v, false, err
	}
	sqlStr := fmt.Sprintf(
		"SELECT %s FROM %s %s",
		field,
		m.dialect.QuoteIdentifier(m.table),
		m.whereLimitOffset(where, 1, 0),
	)
	db, ok := m.client.(Queryer)
	if !ok {
		return v, false, fmt.Errorf("client (%T) is not Queryer", m.client)
	}
	return QueryOne[T](ctx, db, sqlStr, args...)
}

var reOrderBy = regexp.MustCompile(`(?i)\border\s+by\b`)

func (m *Model[T]) whereLimitOffset(where string, limit int, offset int) string {
	where = m.connectWhere(where)
	after := m.dialect.LimitOffsetClause(limit, offset)
	if after == "" {
		return where
	}
	if !m.dialect.LimitOffsetRequiresOrderBy() || reOrderBy.MatchString(where) {
		return where + " " + after
	}
	// 目前只有 sqlserver 需要，SELECT NULL 使其满足语法要求
	return where + " ORDER BY (SELECT NULL) " + after
}

// FindByPK 使用主键查找数据
//
// 需要在 tag 里有 primaryKey 属性: 如 ID int64 `db:"id,pk"`
//
//	可通过 SetSelectFields、SetSelectIgnore 限制查询返回的字段
func (m *Model[T]) FindByPK(ctx context.Context, v T) (nv T, ok bool, err error) {
	if m.err != nil {
		return nv, false, m.err
	}
	where, args, err := m.pkWhereArgs(v, encoder.ActionSelect)
	if err != nil {
		return nv, false, err
	}
	return m.First(ctx, where, args...)
}

// List 查询并返回满足条件的数据。
//
//	可以使用 Limit(xxx).Offset(xxx) 限制返回条数和偏移量
//	可通过 SetSelectFields、SetSelectIgnore 限制查询返回的字段
func (m *Model[T]) List(ctx context.Context, where string, args ...any) ([]T, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []T
	for item, err := range m.ListIter(ctx, where, args...) {
		if err != nil {
			return result, err
		}
		result = append(result, item)
	}
	return result, nil
}

// ListIter 查询满足条件的数据并返回一个迭代器。数据是流式从数据库返回的。
//
//	可以使用 Limit(xxx).Offset(xxx) 限制返回条数和偏移量
//	可通过 SetSelectFields、SetSelectIgnore 限制查询返回的字段
func (m *Model[T]) ListIter(ctx context.Context, where string, args ...any) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		if m.err != nil {
			yield(zero, m.err)
			return
		}

		field, err := m.getSelectFields()
		if err != nil {
			yield(zero, err)
			return
		}

		where, args, err = m.buildWhere(0, where, args)
		if err != nil {
			yield(zero, err)
			return
		}

		sqlStr := fmt.Sprintf(
			"SELECT %s FROM %s %s",
			field,
			m.dialect.QuoteIdentifier(m.table),
			m.whereLimitOffset(where, m.limit, m.offset),
		)

		db, ok := m.client.(Queryer)
		if !ok {
			err = fmt.Errorf("client (%T) is not Queryer", m.client)
			yield(zero, err)
			return
		}
		for item, err := range QueryManyIter[T](ctx, db, sqlStr, args...) {
			if !yield(item, err) {
				return
			}
		}
	}
}

func (m *Model[T]) Count(ctx context.Context, field string, where string, args ...any) (num int64, err error) {
	if m.err != nil {
		return 0, m.err
	}
	where, args, err = m.buildWhere(0, where, args)
	if err != nil {
		return 0, err
	}
	if field == "" {
		field = "*"
	} else if field != "*" && !strings.ContainsRune(field, ' ') {
		field = m.dialect.QuoteIdentifier(field)
	}
	return m.doCount(ctx, field, where, args...)
}

func (m *Model[T]) doCount(ctx context.Context, field string, where string, args ...any) (num int64, err error) {
	sqlStr := fmt.Sprintf("SELECT count(%s) from %s %s",
		field,
		m.dialect.QuoteIdentifier(m.table),
		m.connectWhere(where),
	)
	db, ok := m.client.(RowQuerier)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not RowQuerier", m.client)
	}
	return Count(ctx, db, sqlStr, args...)
}

// ListPage 分页查询，适应于数据量不太大的场景
//
//	可通过 SetSelectFields、SetSelectIgnore 限制查询返回的字段
func (m *Model[T]) ListPage(ctx context.Context, page int, size int, where string, args ...any) (Pagination, []PageRecord[T], error) {
	if m.err != nil {
		return Pagination{}, nil, m.err
	}
	if size < 1 {
		return Pagination{}, nil, fmt.Errorf("invalid size=%d", size)
	}
	var err error
	where, args, err = m.buildWhere(0, where, args)
	if err != nil {
		return Pagination{}, nil, err
	}

	page = max(page, 1) // 最小值为 1
	total, err := m.doCount(ctx, "*", where, args...)
	if err != nil {
		return Pagination{}, nil, err
	}

	info := Pagination{
		TotalRecords: int(total),
		PageSize:     size,
		PageIndex:    page,
	}

	offset := (page - 1) * size
	if int64(offset) >= total {
		return info, nil, nil
	}
	m1 := m.Clone()
	result, err := m1.Limit(size).Offset(offset).List(ctx, where, args...)
	if err != nil {
		return info, nil, err
	}
	items := ToPageRecords[T](result, (page-1)*size)
	return info, items, nil
}
