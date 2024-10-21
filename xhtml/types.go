//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml

import (
	"html"
)

// Bytes 将 []byte 转换为 Element 类型，原样输出 HTML
type Bytes []byte

// HTML 实现 Element 接口
func (b Bytes) HTML() ([]byte, error) {
	return b, nil
}

// Text 文本，输出的时候会自动调用 html.EscapeString
type Text String

// HTML 实现 Element 接口
func (b Text) HTML() ([]byte, error) {
	return []byte(html.EscapeString(string(b))), nil
}

// String 将 String 转换为 Element 类型，原样输出 HTML
type String string

// HTML 实现 Element 接口
func (s String) HTML() ([]byte, error) {
	return []byte(s), nil
}

// StringSlice 将 []string 转换为 Element 类型
type StringSlice []string

// ToElements 转换为 字段 tag 的 []Element
func (ss StringSlice) ToElements(tag string, fn func(b *Any)) Elements {
	if len(ss) == 0 {
		return nil
	}
	cs := make([]Element, len(ss))
	for i := 0; i < len(ss); i++ {
		b := &Any{
			Tag:  tag,
			Body: ToElements(String(ss[i])),
		}
		if fn != nil {
			fn(b)
		}
		cs[i] = b
	}
	return cs
}

func (ss StringSlice) HTML() ([]byte, error) {
	return NewUl(ss).HTML()
}

func (ss StringSlice) toOptions() []Element {
	return toOptions(ss)
}

func (ss StringSlice) ToDatalist(id string) *Any {
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

var (
	// NL 换行: \n
	NL = Bytes("\n")

	// BR HTML 换行 br
	BR = Bytes("<br/>")

	// HR HTML 分割符 hr
	HR = Bytes("<hr/>")
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
	~float32 | ~float64
}
