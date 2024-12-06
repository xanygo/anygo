//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml

import (
	"errors"
	"html/template"
)

// Element 所有 HTML 组件的基础定义
type Element interface {
	HTML() ([]byte, error)
}

type ElementFunc func() ([]byte, error)

func (fn ElementFunc) HTML() ([]byte, error) {
	return fn()
}

func ElementValue(code []byte, err error) Element {
	return elementValue{
		code: code,
		err:  err,
	}
}

var _ Element = (*elementValue)(nil)

type elementValue struct {
	code []byte
	err  error
}

func (e elementValue) HTML() ([]byte, error) {
	return e.code, e.err
}

// ToElements 转换为 Elements 类型
func ToElements(es ...Element) Elements {
	return es
}

// Elements alias of []Element
type Elements []Element

// HTML 实现 Element 接口
func (hs Elements) HTML() ([]byte, error) {
	if len(hs) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	for i := 0; i < len(hs); i++ {
		bw.Write(hs[i])
	}
	return bw.HTML()
}

// ErrEmptyTagName tag 值为空的错误
var ErrEmptyTagName = errors.New("empty tag name")

var _ Element = (*Any)(nil)
var _ AttrsMapper = (*Any)(nil)

// NewAny 创建任意的 tag
func NewAny(tag string) *Any {
	return &Any{
		Tag: tag,
	}
}

// Container 允许添加子元素
type Container interface {
	Add(children ...Element)
}

// AddTo 给指定的对象添加子元素
func AddTo(to Container, children ...Element) {
	to.Add(children...)
}

// WithAny 对 any 元素进行处理
func WithAny(a *Any, fn func(*Any)) *Any {
	fn(a)
	return a
}

// Any 一块 HTML 内容
type Any struct {
	// Tag 标签名称，必填，如 div
	Tag string

	// WithAttrs 可选，属性信息
	WithAttrs

	// Body 内容，可选
	Body Elements

	// SelfClose 当前标签是否自关闭,默认为 false
	// 如 img 标签就是自关闭的：<img src="/a.jpg"/>
	SelfClose bool
}

// Add 添加子元素
func (c *Any) Add(children ...Element) {
	c.Body = append(c.Body, children...)
}

// HTML 实现 Element 接口
func (c *Any) HTML() ([]byte, error) {
	if len(c.Tag) == 0 {
		return nil, ErrEmptyTagName
	}
	bw := newBufWriter()
	bw.Write("<", c.Tag)
	bw.WriteWithSep(" ", c.Attrs)
	if c.SelfClose {
		bw.Write("/>")
	} else {
		bw.Write(">")
		bw.Write(c.Body)
		bw.Write("</", c.Tag, ">")
	}
	return bw.HTML()
}

func (c *Any) TplHTML() template.HTML {
	bf, err := c.HTML()
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(bf)
}
