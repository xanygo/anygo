//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

//
// https://html.spec.whatwg.org/#element-interfaces

package xhtml

import "html/template"

// NewHTML html 标签
func NewHTML() *Any {
	return NewAny("html")
}

// NewHead 创建一个 <head>
func NewHead() *Any {
	return NewAny("head")
}

// NewTitle 创建一个 <title>
func NewTitle(c Element) *Any {
	return &Any{
		Tag:  "title",
		Body: ToElements(c),
	}
}

// NewBody 创建一个 <body>
func NewBody() *Any {
	return NewAny("body")
}

// NewDiv 创建一个 <div>
func NewDiv() *Any {
	return NewAny("div")
}

// NewNav 创建一个 <nav>
func NewNav() *Any {
	return NewAny("nav")
}

// NewP 创建一个 <p>
func NewP() *Any {
	return NewAny("p")
}

// NewDL 创建一个 <dl>
func NewDL() *Any {
	return NewAny("dl")
}

// NewDT 创建一个 <dt>
func NewDT() *Any {
	return NewAny("dt")
}

// NewArticle 创建一个 <article>
func NewArticle() *Any {
	return NewAny("article")
}

// NewPre 创建一个 <pre>
func NewPre() *Any {
	return NewAny("pre")
}

// NewCode 创建一个 <code>
func NewCode() *Any {
	return NewAny("code")
}

// NewFigure 创建一个 <figure>
func NewFigure() *Any {
	return NewAny("figure")
}

// NewFigCaption 创建一个 <figcaption>
func NewFigCaption() *Any {
	return NewAny("figcaption")
}

type selfCloseTag struct {
	WithAttrs
}

func (m *selfCloseTag) set(key string, value string) {
	attr := &Attr{
		Key:    key,
		Values: []string{value},
	}
	m.MustAttrs().Set(attr)
}

func (m *selfCloseTag) html(begin string) ([]byte, error) {
	bw := newBufWriter()
	bw.Write(begin)
	bw.WriteWithSep(" ", m.Attrs)
	bw.Write("/>")
	return bw.HTML()
}

// NewIMG 创建一个  img  标签
func NewIMG(src string) *IMG {
	return (&IMG{}).SRC(src)
}

// IMG 图片 img 标签
type IMG struct {
	selfCloseTag
}

func (m *IMG) set(key string, value string) *IMG {
	m.selfCloseTag.set(key, value)
	return m
}

// SRC 设置 src 属性
func (m *IMG) SRC(src string) *IMG {
	return m.set("src", src)
}

// ALT 设置 alt 属性
func (m *IMG) ALT(alt string) *IMG {
	return m.set("alt", alt)
}

// HTML 转换为 html
func (m *IMG) HTML() ([]byte, error) {
	return m.html("<img")
}

// NewA 创建一个 <a>
func NewA(href string) *Any {
	a := NewAny("a")
	SetHref(a, href)
	return a
}

// NewMeta 创建一个新的 <meta>
func NewMeta() *Meta {
	return &Meta{}
}

var _ Element = (*Meta)(nil)

// Meta 页面元信息标签 meth
type Meta struct {
	selfCloseTag
}

func (a *Meta) set(key string, value string) *Meta {
	a.selfCloseTag.set(key, value)
	return a
}

// Name 设置 name 属性
func (a *Meta) Name(name string) *Meta {
	return a.set("name", name)
}

// Charset 设置 charset 属性
func (a *Meta) Charset(charset string) *Meta {
	return a.set("charset", charset)
}

// HTTPEquiv 设置 http-equiv 属性
func (a *Meta) HTTPEquiv(equiv string) *Meta {
	return a.set("http-equiv", equiv)
}

// Content 设置 content 属性
func (a *Meta) Content(content string) *Meta {
	return a.set("content", content)
}

// Media 设置 media 属性
func (a *Meta) Media(media string) *Meta {
	return a.set("media", media)
}

// HTML 转换为 html
func (a *Meta) HTML() ([]byte, error) {
	return a.html("<meta")
}

// NewScript script 标签
func NewScript() *Any {
	return NewAny("script")
}

// NewStyle style 标签
func NewStyle() *Any {
	a := NewAny("style")
	SetType(a, "text/css")
	return a
}

// NewLink 创建一个新的 link 标签
func NewLink() *Link {
	return &Link{}
}

// Link 页面元素 link 标签
type Link struct {
	selfCloseTag
}

// Rel 设置 rel 属性
func (a *Link) Rel(rel string) *Link {
	a.set("rel", rel)
	return a
}

// Type 设置 tp 属性
func (a *Link) Type(tp string) *Link {
	a.set("type", tp)
	return a
}

// Href 设置 href 属性
func (a *Link) Href(href string) *Link {
	a.set("href", href)
	return a
}

// HTML 转换为 html
func (a *Link) HTML() ([]byte, error) {
	return a.html("<link")
}

// NewForm 创建一个 form
func NewForm(method string, action string) *Any {
	f := NewAny("form")
	SetMethod(f, method)
	SetAction(f, action)
	return f
}

// NewFieldSet 创建一个 fieldset
func NewFieldSet() *Any {
	return NewAny("fieldset")
}

// NewLegend 创建一个 legend
func NewLegend(e Element) *Any {
	a := &Any{
		Tag:  "legend",
		Body: ToElements(e),
	}
	return a
}

// WithLabel 使用 label 包裹元素
func WithLabel(es ...Element) *Any {
	return &Any{
		Tag:  "label",
		Body: es,
	}
}

func NewLabel(html string) *Any {
	l := &Any{
		Tag:  "label",
		Body: ToElements(HTMLString(html)),
	}
	return l
}

// NewSelect 创建一个 select
func NewSelect(name string, opts ...Element) *Any {
	input := &Any{
		Tag:  "select",
		Body: opts,
	}
	SetName(input, name)
	return input
}

// NewOption 创建一个 option
func NewOption(value string, b Element) *Any {
	if b == nil {
		b = TextString(value)
	}
	input := &Any{
		Tag:  "option",
		Body: ToElements(b),
	}
	SetValue(input, value)
	return input
}

func NewButton(text string) *Any {
	return &Any{
		Tag:  "button",
		Body: ToElements(TextString(text)),
	}
}

func toOptions[T ~string](ss []T) []Element {
	result := make([]Element, len(ss))
	for i := range ss {
		opt := &Any{
			Tag:       "option",
			SelfClose: true,
		}
		SetValue(opt, string(ss[i]))
		result[i] = opt
	}
	return result
}

func NewDatalist[T ~string](id string, ss []T) *Any {
	a := &Any{
		Tag:  "datalist",
		Body: ToElements(toOptions(ss)...),
	}
	return a
}

func WithAddress(es ...Element) *Any {
	return &Any{
		Tag:  "address",
		Body: es,
	}
}

func Render(e Element) template.HTML {
	bf, err := e.HTML()
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(bf)
}

func render(e Element) (template.HTML, error) {
	bf, err := e.HTML()
	return template.HTML(bf), err
}
