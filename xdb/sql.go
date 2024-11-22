//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-20

package xdb

import "strings"

type Builder interface {
	Build() (string, []any)
}

type Condition struct {
	builder strings.Builder
	args    []any
	orderBy []string
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
