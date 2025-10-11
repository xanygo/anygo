//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml

import (
	"html"
	"unsafe"
)

var (
	// NL 换行: \n
	NL = HTMLBytes("\n")

	// BR HTML 换行 br
	BR = HTMLBytes("<br/>")

	// HR HTML 分割符 hr
	HR = HTMLBytes("<hr/>")
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// HTMLBytes 将 []byte 转换为 Element 类型，原样输出 HTML
type HTMLBytes []byte

// HTML 实现 Element 接口
func (b HTMLBytes) HTML() ([]byte, error) {
	return b, nil
}

func (b HTMLBytes) Pre() ([]byte, error) {
	return b.Wrap("pre")
}

func (b HTMLBytes) Div() ([]byte, error) {
	return b.Wrap("div")
}

func (b HTMLBytes) Li() ([]byte, error) {
	return b.Wrap("li")
}

func (b HTMLBytes) P() ([]byte, error) {
	return b.Wrap("p")
}

func (b HTMLBytes) Wrap(tag string) ([]byte, error) {
	return wrap(tag, b)
}

func wrap[T ~string | ~[]byte](tag string, code T) ([]byte, error) {
	bf := make([]byte, 0, len(code)+1+len(tag)*2)
	bf = append(bf, '<')
	bf = append(bf, tag...)
	bf = append(bf, '>')
	bf = append(bf, code...)
	bf = append(bf, '<', '/')
	bf = append(bf, tag...)
	bf = append(bf, '>')
	return bf, nil
}

// HTMLString 将 String 转换为 Element 类型，原样输出 HTML
type HTMLString string

// HTML 实现 Element 接口
func (b HTMLString) HTML() ([]byte, error) {
	return unsafe.Slice(unsafe.StringData(string(b)), len(b)), nil
}

func (b HTMLString) Pre() ([]byte, error) {
	return b.Wrap("pre")
}

func (b HTMLString) Div() ([]byte, error) {
	return b.Wrap("div")
}

func (b HTMLString) Li() ([]byte, error) {
	return b.Wrap("li")
}

func (b HTMLString) P() ([]byte, error) {
	return b.Wrap("p")
}

func (b HTMLString) Wrap(tag string) ([]byte, error) {
	return wrap(tag, b)
}

// TextBytes 将 []byte 转换为 Element 类型，会转换为 html.EscapeString
type TextBytes []byte

// HTML 实现 Element 接口
func (b TextBytes) HTML() ([]byte, error) {
	be := html.EscapeString(string(b))
	return unsafe.Slice(unsafe.StringData(be), len(be)), nil
}

func (b TextBytes) Pre() ([]byte, error) {
	return b.Wrap("pre")
}

func (b TextBytes) Div() ([]byte, error) {
	return b.Wrap("div")
}

func (b TextBytes) Li() ([]byte, error) {
	return b.Wrap("li")
}

func (b TextBytes) P() ([]byte, error) {
	return b.Wrap("p")
}

func (b TextBytes) Wrap(tag string) ([]byte, error) {
	be := html.EscapeString(string(b))
	return wrap(tag, be)
}

// TextString 文本，输出的时候会自动调用 html.EscapeString
type TextString string

// HTML 实现 Element 接口
func (b TextString) HTML() ([]byte, error) {
	be := html.EscapeString(string(b))
	return unsafe.Slice(unsafe.StringData(be), len(be)), nil
}

func (b TextString) Pre() ([]byte, error) {
	return b.Wrap("pre")
}

func (b TextString) Div() ([]byte, error) {
	return b.Wrap("div")
}

func (b TextString) Li() ([]byte, error) {
	return b.Wrap("li")
}

func (b TextString) P() ([]byte, error) {
	return b.Wrap("p")
}

func (b TextString) Wrap(tag string) ([]byte, error) {
	be := html.EscapeString(string(b))
	return wrap(tag, be)
}

// HTMLStringSlice 将 []string 转换为 Element 类型
type HTMLStringSlice[T ~string] []T

// Elements 转换为 字段 tag 的 []Element
func (ss HTMLStringSlice[T]) Elements(tag string, fn func(b *Any)) Elements {
	if len(ss) == 0 {
		return nil
	}
	cs := make([]Element, len(ss))
	for i := 0; i < len(ss); i++ {
		b := &Any{
			Tag:  tag,
			Body: ToElements(HTMLString(ss[i])),
		}
		if fn != nil {
			fn(b)
		}
		cs[i] = b
	}
	return cs
}

func (ss HTMLStringSlice[T]) HTML() ([]byte, error) {
	return ss.UL()
}

func (ss HTMLStringSlice[T]) UL() ([]byte, error) {
	ul := &Any{
		Tag:  "ul",
		Body: ss.Elements("li", nil),
	}
	return ul.HTML()
}

func (ss HTMLStringSlice[T]) OL() ([]byte, error) {
	ol := &Any{
		Tag:  "ol",
		Body: ss.Elements("li", nil),
	}
	return ol.HTML()
}

func (ss HTMLStringSlice[T]) Datalist(id string) *Any {
	return NewDatalist(id, ss)
}

// TextStringSlice 将 []string 转换为 Element 类型
type TextStringSlice []string

// Elements 转换为 字段 tag 的 []Element
func (ss TextStringSlice) Elements(tag string, fn func(b *Any)) Elements {
	if len(ss) == 0 {
		return nil
	}
	cs := make([]Element, len(ss))
	for i := 0; i < len(ss); i++ {
		b := &Any{
			Tag:  tag,
			Body: ToElements(TextString(ss[i])),
		}
		if fn != nil {
			fn(b)
		}
		cs[i] = b
	}
	return cs
}

func (ss TextStringSlice) HTML() ([]byte, error) {
	return ss.UL()
}

func (ss TextStringSlice) UL() ([]byte, error) {
	ul := &Any{
		Tag:  "ul",
		Body: ss.Elements("li", nil),
	}
	return ul.HTML()
}

func (ss TextStringSlice) OL() ([]byte, error) {
	ol := &Any{
		Tag:  "ol",
		Body: ss.Elements("li", nil),
	}
	return ol.HTML()
}

func (ss TextStringSlice) Datalist(id string) *Any {
	return NewDatalist(id, ss)
}

// Pre 输出 HTML 时添加 pre 标签
type Pre string

func (p Pre) HTML() ([]byte, error) {
	return []byte("<pre>" + p + "</pre>"), nil
}

// PreByte 输出 HTML 时添加 pre 标签
type PreByte []byte

func (p PreByte) HTML() ([]byte, error) {
	bf := make([]byte, 0, len(p)+5+6)
	bf = append(bf, "<pre>"...)
	bf = append(bf, p...)
	bf = append(bf, "</pre>"...)
	return bf, nil
}

// Comment 注释
type Comment string

// HTML 转换为 HTML
func (c Comment) HTML() ([]byte, error) {
	if len(c) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	bw.Write("<!-- ", html.EscapeString(string(c)), " -->\n")
	return bw.HTML()
}
