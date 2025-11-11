//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-20

package xdb

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xslice"
)

var ErrEmptyBuilder = errors.New("builder is empty")

type Builder interface {
	Build() (string, []any, error)
}

type Condition struct {
	builder strings.Builder
	args    []any
	err     error
}

func (c *Condition) Append(op string, str string, args ...any) {
	if c.builder.Len() > 0 {
		c.builder.WriteString(" ")
		c.builder.WriteString(op)
		c.builder.WriteString(" ")
	}
	c.builder.WriteString(str)
	c.args = append(c.args, args...)
}

func (c *Condition) AppendBuilder(op string, str string, b Builder) {
	txt, args, err := b.Build()
	if err != nil {
		c.err = err
		return
	}
	c.Append(op, str+" "+txt, args...)
}

func (c *Condition) And(str string, args ...any) {
	c.Append("AND", str, args...)
}

func (c *Condition) Or(str string, args ...any) {
	c.Append("OR", str, args...)
}

func (c *Condition) Build() (string, []any, error) {
	return c.builder.String(), c.args, c.err
}

func EmptyBuilder() Builder {
	return emptyBuilder{}
}

var _ Builder = (*emptyBuilder)(nil)

type emptyBuilder struct{}

func (e emptyBuilder) Build() (string, []any, error) {
	return "", nil, ErrEmptyBuilder
}

var _ Builder = (*InsertBuilder)(nil)

func NewInsertBuilder(table string) *InsertBuilder {
	return &InsertBuilder{
		table: table,
	}
}

type InsertBuilder struct {
	table   string
	builder strings.Builder
	values  []string
	args    []any
	err     error
	fields  []string // 字段名
}

func (ib *InsertBuilder) Values(values ...map[string]any) {
	if ib.table == "" {
		ib.err = errors.New("table name is empty")
		return
	}
	for _, value := range values {
		if len(value) == 0 {
			continue
		}
		if len(ib.fields) == 0 {
			fields := xmap.Keys(value)
			slices.Sort(fields)
			ib.fields = fields
			ib.doFields()
		}
		if err := ib.doValue(value); err != nil {
			ib.err = err
			return
		}
		str := "(" + strings.Join(xslice.Repeat("?", len(ib.fields)), ",") + ")"
		ib.values = append(ib.values, str)
	}
}

func (ib *InsertBuilder) doFields() {
	ib.builder.WriteString("INSERT INTO ")
	ib.builder.WriteString(ib.table)
	ib.builder.WriteString(" (")
	ib.builder.WriteString(strings.Join(ib.fields, ","))
	ib.builder.WriteString(ibValue)
}

const ibValue = ") VALUES "

func (ib *InsertBuilder) doValue(value map[string]any) error {
	la := len(ib.fields)
	lb := len(value)
	if la != lb {
		return fmt.Errorf("fields not eq (%d!=%d), expect fields %q, got %q", la, lb, ib.fields, xmap.Keys(value))
	}
	for _, field := range ib.fields {
		val, ok := value[field]
		if !ok {
			return fmt.Errorf("field %q not found", field)
		}
		ib.args = append(ib.args, val)
	}
	return nil
}

func (ib *InsertBuilder) Append(op string, str string, args ...any) {
	if ib.builder.Len() > 0 {
		ib.builder.WriteString(" ")
		ib.builder.WriteString(op)
		ib.builder.WriteString(" ")
	}
	ib.builder.WriteString(str)
	ib.args = append(ib.args, args...)
}

func (ib *InsertBuilder) AppendBuilder(op string, b Builder) {
	str, arg, err := b.Build()
	if err != nil {
		ib.err = err
	}
	ib.Append(op, " ( "+str+" ) ", arg...)
}

func (ib *InsertBuilder) Build() (string, []any, error) {
	if len(ib.fields) == 0 {
		return "", nil, ErrEmptyBuilder
	}

	if len(ib.values) > 0 {
		ib.builder.WriteString(strings.Join(ib.values, ","))
		clear(ib.values)
		ib.values = nil
	}

	return ib.builder.String(), ib.args, ib.err
}

var _ Builder = (*InBuilder)(nil)

type InBuilder struct {
	Value []any
}

func (in *InBuilder) Build() (string, []any, error) {
	if len(in.Value) == 0 {
		return "", nil, errors.New("no value with InBuilder")
	}
	txt := strings.Join(xslice.Repeat("?", len(in.Value)), ",")
	return txt, in.Value, nil
}
