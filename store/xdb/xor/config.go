package xor

import (
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/store/xdb/internal/encoder"
)

type config struct {
	schema  *dbtype.TableSchema
	dialect dbtype.Dialect

	// 下面是配置参数
	noWhere bool // 是否允许没有 where

	where     string
	whereArgs []any
	orderBy   string
	groupBy   string
	limit     int
	offset    int
	columns   []string
	ignores   []string

	// 传入的配置错误
	errs []error
}

func (c *config) Clone() *config {
	return &config{
		schema:  c.schema,
		dialect: c.dialect,

		where:     c.where,
		whereArgs: slices.Clone(c.whereArgs),
		noWhere:   c.noWhere,

		orderBy: c.orderBy,
		limit:   c.limit,
		offset:  c.offset,
		columns: slices.Clone(c.columns),
		ignores: slices.Clone(c.ignores),

		errs: slices.Clone(c.errs),
	}
}

func (c *config) reset() {
	c.where = ""
	c.noWhere = false
	c.whereArgs = nil
	c.orderBy = ""
	c.limit = 0
	c.offset = 0
	c.columns = nil
	c.ignores = nil
	c.errs = nil
}

func (c *config) getEncoder(action encoder.Action) encoder.Encoder[any] {
	return encoder.Encoder[any]{
		Schema:       c.schema,
		Action:       action,
		Dialect:      c.dialect,
		OnlyFields:   c.getColumns(),
		IgnoreFields: c.getIgnores(),
	}
}

func (c *config) getColumns() []string {
	if c != nil {
		return c.columns
	}
	return nil
}

func (c *config) getIgnores() []string {
	if c != nil {
		return c.ignores
	}
	return nil
}

func (c *config) merge(opts ...Option) *config {
	for _, item := range opts {
		item.withOption(c)
	}
	return c
}

func (c *config) mergeOnClone(opts ...Option) *config {
	if len(opts) == 0 {
		return c
	}
	clone := c.Clone()
	for _, item := range opts {
		item.withOption(clone)
	}
	return clone
}

func (c *config) getSelectFields() string {
	if len(c.columns) > 0 {
		return strings.Join(xslice.MapFunc(c.columns, c.dialect.QuoteIdentifier), ",")
	}

	fields := slices.Clone(c.schema.ColumnNames)

	if len(c.ignores) != 0 {
		fields = xslice.DeleteValue(fields, c.ignores...)
	}

	if len(fields) == 0 {
		return "*"
	}

	return strings.Join(xslice.MapFunc(fields, c.dialect.QuoteIdentifier), ",")
}

func (c *config) addError(err error) {
	if err == nil {
		return
	}
	c.errs = append(c.errs, err)
}

func (c *config) getError() error {
	if len(c.errs) == 0 {
		return nil
	}
	return errors.Join(c.errs...)
}

func (c *config) getWhereArgs(paramIndesStart int) (where string, args []any, err error) {
	if err = c.getError(); err != nil {
		return "", nil, err
	}

	if len(c.whereArgs) > 0 {
		args, err = c.getEncoder(encoder.ActionSelect).EncodeArgs(c.whereArgs...)
		if err != nil {
			return "", nil, err
		}
	}

	if !c.noWhere && c.where == "" {
		return "", nil, xdb.ErrEmptyWhere
	}

	return c.getSQLTail(paramIndesStart), args, nil
}

var reOrderBy = regexp.MustCompile(`(?i)\border\s+by\b`)
var reWherePlaceholder = regexp.MustCompile(`\{([^{}]*)\}`)

func (c *config) getSQLTail(paramIndesStart int) string {
	var b strings.Builder

	if c.where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(c.replacePlaceholder(paramIndesStart, c.where))
	}

	if c.orderBy != "" {
		b.WriteString(" ORDER BY ")
		b.WriteString(c.orderBy)
	}

	str := c.dialect.LimitOffsetClause(c.limit, c.offset)
	if str != "" {
		if c.dialect.LimitOffsetRequiresOrderBy() && !reOrderBy.MatchString(b.String()) {
			// 目前只有 sqlserver 需要，SELECT NULL 使其满足语法要求
			b.WriteString(" ORDER BY (SELECT NULL)")
		}
		b.WriteString(" ")
		b.WriteString(str)
	}

	if c.groupBy != "" {
		b.WriteString(" GROUP BY ")
		b.WriteString(c.groupBy)
	}

	where := b.String()

	// 将条件中的 RAND() 换成方言
	if strings.Contains(where, KWRand) {
		if dr := c.dialect.RandomOrder(); dr != KWRand {
			where = strings.ReplaceAll(where, KWRand, dr)
		}
	}

	// 将 where 条件中的 {name} 转义
	if strings.ContainsRune(where, '{') {
		where = reWherePlaceholder.ReplaceAllStringFunc(where, func(m string) string {
			name := m[1 : len(m)-1] // 去掉 {}
			return c.dialect.QuoteIdentifier(name)
		})
	}

	return where
}

func (c *config) replacePlaceholder(indexStart int, where string) string {
	// 将 ? 替换为方言的占位符，如 $1, $2 ...
	if c.dialect.BindVar(0) != "?" {
		var sb strings.Builder
		idx := 1
		for i := 0; i < len(where); i++ {
			if where[i] == '?' {
				sb.WriteString(c.dialect.BindVar(indexStart + idx))
				idx++
			} else {
				sb.WriteByte(where[i])
			}
		}
		where = sb.String()
	}

	return where
}
