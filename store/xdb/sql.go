//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-20

package xdb

import (
	"context"
	"strings"
)

type Builder interface {
	Build() (string, []any)
}

type Condition struct {
	builder strings.Builder
	args    []any
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

func (c *Condition) And(str string, args ...any) {
	c.Append("AND", str, args...)
}

func (c *Condition) Or(str string, args ...any) {
	c.Append("OR", str, args...)
}

func (c *Condition) Build() (string, []any) {
	return c.builder.String(), c.args
}

func EmptyBuilder() Builder {
	return emptyBuilder{}
}

var _ Builder = (*emptyBuilder)(nil)

type emptyBuilder struct{}

func (e emptyBuilder) Build() (string, []any) {
	return "", nil
}

// PageList 分页查询
//
// page: 当前页码，总是 >=1
// size: 查询结果数，总是 >=1
func PageList[T any](ctx context.Context, b Builder, page int, size int, search CountSearch[T]) (Pagination, []Record[T], error) {
	if b == nil {
		b = EmptyBuilder()
	}
	total, datas, err := search(ctx, b, page, size)
	if err != nil {
		return Pagination{}, nil, err
	}
	pageInfo := Pagination{
		Page:  page,
		Total: int(total),
		Size:  size,
	}
	items := make([]Record[T], len(datas))
	for idx, value := range datas {
		items[idx] = Record[T]{
			Value: value,
			Order: idx,
			Index: (page-1)*size + idx,
			Ext:   map[string]any{},
		}
	}
	return pageInfo, items, nil
}

// CountSearch 统计结果数以及分页的结果集
type CountSearch[T any] func(ctx context.Context, b Builder, page int, size int) (int64, []T, error)

type Record[T any] struct {
	Value T
	Order int // 当前页面索引
	Index int // 在所有页面的索引
	Ext   map[string]any
}

func (r Record[T]) HumanIndex() int {
	return r.Index + 1
}
