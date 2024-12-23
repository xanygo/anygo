//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/xslice"
)

const (
	// onlyKey 属性只需要 key，不需要 value
	onlyKey = ":only-key"
)

// Attrs 多个属性
type Attrs struct {
	// Sep 多个属性间的连接符，当为空时，使用默认值 " " (一个空格)
	Sep string

	// KVSep key 和 value 之间的连接符，当为空时，使用默认值 =
	KVSep string

	attrs map[string]*Attr
	keys  []string
}

// GetSep  多个属性间的连接符，当为空时，返回默认值 " " (一个空格)
func (a *Attrs) GetSep() string {
	if len(a.Sep) == 0 {
		return " "
	}
	return a.Sep
}

// GetKVSep key 和 value 之间的连接符，当为空时， 返回默认值 =
func (a *Attrs) GetKVSep() string {
	if len(a.KVSep) == 0 {
		return "="
	}
	return a.KVSep
}

// Attr 返回一个指定的属性，若不存在，返回 nil
func (a *Attrs) Attr(key string) *Attr {
	if len(a.attrs) == 0 {
		return nil
	}
	return a.attrs[key]
}

// MustAttr 返回一个指定的属性，若不存在，返回 nil
func (a *Attrs) MustAttr(key string) *Attr {
	if val := a.Attr(key); val != nil {
		return val
	}
	attr := &Attr{
		Key: key,
	}
	a.Set(attr)
	return attr
}

// Delete 删除指定 key 的属性
func (a *Attrs) Delete(keys ...string) {
	if len(a.attrs) == 0 {
		return
	}
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		delete(a.attrs, key)
		a.keys = xslice.DeleteValue(a.keys, key)
	}
}

// Keys 返回所有属性的 key
func (a *Attrs) Keys() []string {
	return a.keys
}

// Set 设置属性值
func (a *Attrs) Set(attr ...*Attr) {
	if a.attrs == nil {
		a.attrs = make(map[string]*Attr, len(attr))
	}
	for _, item := range attr {
		if _, has := a.attrs[item.Key]; !has {
			a.keys = append(a.keys, item.Key)
		}
		a.attrs[item.Key] = item
	}
}

// HTML 转换为 HTML
func (a *Attrs) HTML() ([]byte, error) {
	if a == nil {
		return nil, nil
	}
	return attrsHTML(a, a.GetKVSep(), strconv.Quote, a.GetSep())
}

func attrsHTML(attrs *Attrs, kvSep string, quote func(string) string, sep string) ([]byte, error) {
	keys := attrs.Keys()
	if len(keys) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	for i := 0; i < len(keys); i++ {
		attrKey := keys[i]
		vs := attrs.Attr(attrKey).Values
		if len(vs) == 0 {
			continue
		}
		bw.Write(attrKey)
		if vs[0] != onlyKey {
			bw.Write(kvSep, quote(strings.Join(vs, " ")))
		}
		if i != len(keys)-1 {
			bw.Write(sep)
		}
	}
	return bw.HTML()
}

// AttrsMapper 具有 AttrsMapper 方法
type AttrsMapper interface {
	MustAttrs() *Attrs
	FindAttrs() *Attrs
}

var _ AttrsMapper = (*WithAttrs)(nil)

// WithAttrs 具有 attrs 属性
type WithAttrs struct {
	Attrs *Attrs
}

// FindAttrs 返回当前的 Attrs
func (w *WithAttrs) FindAttrs() *Attrs {
	return w.Attrs
}

// MustAttrs 若 attrs 不存在，则创建 并返回 attrs
func (w *WithAttrs) MustAttrs() *Attrs {
	if w.Attrs == nil {
		w.Attrs = &Attrs{}
	}
	return w.Attrs
}

// DeleteAttr 删除指定的属性值
func DeleteAttr(w AttrsMapper, key string, values ...string) {
	as := w.FindAttrs()
	if as == nil {
		return
	}
	attr := as.Attr(key)
	if attr == nil {
		return
	}
	attr.Delete(values...)
}

// Attr  一个属性
type Attr struct {
	// Key 属性的名字
	Key string

	// Values 属性值，可以有多个
	Values []string
}

// Set 设置属性值
func (a *Attr) Set(value ...string) {
	a.Values = value
}

// First 返回首个属性值
func (a *Attr) First() string {
	if len(a.Values) == 0 {
		return ""
	}
	return a.Values[0]
}

// Add 添加新的属性值
func (a *Attr) Add(value ...string) {
	a.Values = append(a.Values, value...)
	a.Values = xslice.Unique(a.Values)
}

// Delete 删除属性值
func (a *Attr) Delete(value ...string) {
	a.Values = xslice.DeleteValue(a.Values, value...)
}

func findOrCreateAttr(w AttrsMapper, key string) *Attr {
	as := w.MustAttrs()
	attr := as.Attr(key)
	if attr != nil {
		return attr
	}
	attr = &Attr{
		Key: key,
	}
	as.Set(attr)
	return attr
}

// SetAttr 设置属性值
func SetAttr(w AttrsMapper, key string, value ...string) {
	if len(value) == 1 && len(value[0]) == 0 {
		DeleteAttr(w, key)
	} else {
		findOrCreateAttr(w, key).Set(value...)
	}
}

// SetAttrNoValue 设置只有 key，不需要 value 的属性
func SetAttrNoValue(w AttrsMapper, key string) {
	findOrCreateAttr(w, key).Set(onlyKey)
}

// SetAsync 设置 async  属性
func SetAsync(w AttrsMapper) {
	SetAttrNoValue(w, "async")
}

// SetAccept 设置 accept 属性
func SetAccept(w AttrsMapper, accept string) {
	SetAttr(w, "accept", accept)
}

// SetClass 设置 class 属性
func SetClass(w AttrsMapper, class ...string) {
	SetAttr(w, "class", class...)
}

// AddClass 添加 class 属性
func AddClass(w AttrsMapper, class ...string) {
	findOrCreateAttr(w, "class").Add(class...)
}

// DeleteClass 删除 class 属性
func DeleteClass(w AttrsMapper, class ...string) {
	DeleteAttr(w, "class", class...)
}

// SetID 设置元素的 id
func SetID(w AttrsMapper, id string) {
	SetAttr(w, "id", id)
}

// SetName 设置元素的 name
func SetName(w AttrsMapper, name string) {
	SetAttr(w, "name", name)
}

// SetWidth 设置元素的 width
func SetWidth(w AttrsMapper, width string) {
	SetAttr(w, "width", width)
}

// SetHeight 设置元素的 height
func SetHeight(w AttrsMapper, height string) {
	SetAttr(w, "height", height)
}

// SetSize 设置元素的 size
func SetSize(w AttrsMapper, size int) {
	SetAttr(w, "size", strconv.Itoa(size))
}

// SetLang 设置元素的 lang 属性
//
//	如 en-US、zh-CN
func SetLang(w AttrsMapper, lang string) {
	SetAttr(w, "lang", lang)
}

// SetTitle 设置 title 属性
func SetTitle(w AttrsMapper, title string) {
	SetAttr(w, "title", title)
}

// SetSrc 设置 src 属性
func SetSrc(w AttrsMapper, src string) {
	SetAttr(w, "src", src)
}

func SetHref(w AttrsMapper, href string) {
	SetAttr(w, "href", href)
}

// SetTarget 设置 target 属性
func SetTarget(w AttrsMapper, target string) {
	SetAttr(w, "target", target)
}

func SetTargetBlank(w AttrsMapper) {
	SetTarget(w, "_blank")
}

func AddTextContent(w Container, txt string) {
	AddTo(w, TextString(txt))
}

func AddHTMLContent(w Container, txt string) {
	AddTo(w, HTMLString(txt))
}

// SetType 设置 type 属性
func SetType(w AttrsMapper, tp string) {
	SetAttr(w, "type", tp)
}

// SetValue 设置 value 属性
func SetValue(w AttrsMapper, value string) {
	SetAttr(w, "value", value)
}

// SetMax 设置 max 属性
func SetMax[N Number](w AttrsMapper, max N) {
	SetAttr(w, "max", fmt.Sprint(max))
}

// SetMaxLength 设置 maxlength 属性
func SetMaxLength(w AttrsMapper, maxLen int) {
	SetAttr(w, "maxlength", strconv.Itoa(maxLen))
}

// SetMin 设置 min 属性
func SetMin[N Number](w AttrsMapper, min N) {
	SetAttr(w, "min", fmt.Sprint(min))
}

// SetMinLength 设置 minlength 属性
func SetMinLength(w AttrsMapper, minLen int) {
	SetAttr(w, "minlength", strconv.Itoa(minLen))
}

// SetStep 设置 step 属性
func SetStep[N Number](w AttrsMapper, max N) {
	SetAttr(w, "step", fmt.Sprint(max))
}

// SetFor 设置 for 属性
func SetFor(w AttrsMapper, f string) {
	SetAttr(w, "for", f)
}

// SetForm 设置 form 属性
func SetForm(w AttrsMapper, form string) {
	SetAttr(w, "form", form)
}

// SetFormAction 设置 formaction 属性
func SetFormAction(w AttrsMapper, formAction string) {
	SetAttr(w, "formaction", formAction)
}

// SetFormTarget 设置 formtarget 属性，可用于 input type="submit"
func SetFormTarget(w AttrsMapper, formAction string) {
	SetAttr(w, "formtarget", formAction)
}

// SetMethod 设置 method 属性
func SetMethod(w AttrsMapper, method string) {
	SetAttr(w, "method", method)
}

// SetAction 设置 action 属性
func SetAction(w AttrsMapper, action string) {
	SetAttr(w, "action", action)
}

// SetList 设置 list 属性
func SetList(w AttrsMapper, list string) {
	SetAttr(w, "list", list)
}

// SetChecked 设置 form 属性
func SetChecked(w AttrsMapper, checked bool) {
	if checked {
		SetAttr(w, "checked", "checked")
	} else {
		DeleteAttr(w, "checked")
	}
}

// SetSelected 设置 selected 属性
func SetSelected(w AttrsMapper, checked bool) {
	if checked {
		SetAttr(w, "selected", "selected")
	} else {
		DeleteAttr(w, "selected")
	}
}

// SetDisabled 设置 disabled 属性
func SetDisabled(w AttrsMapper, disabled bool) {
	if disabled {
		SetAttr(w, "disabled", "disabled")
	} else {
		DeleteAttr(w, "checked")
	}
}

// SetRequired 设置 required 属性
func SetRequired(w AttrsMapper, required bool) {
	if required {
		SetAttr(w, "required", "required")
	} else {
		DeleteAttr(w, "required")
	}
}

// SetReadOnly 设置 readonly 属性
func SetReadOnly(w AttrsMapper, readonly bool) {
	if readonly {
		SetAttr(w, "readonly", "required")
	} else {
		DeleteAttr(w, "readonly")
	}
}

// SetAutoComplete 设置 autocomplete 属性
func SetAutoComplete(w AttrsMapper, on bool) {
	if on {
		SetAttr(w, "autocomplete", "on")
	} else {
		SetAttr(w, "autocomplete", "on")
	}
}

// SetOnChange 设置 onchange 属性
func SetOnChange(w AttrsMapper, script string) {
	SetAttr(w, "onchange", script)
}

// SetOnBlur 设置 onblur 属性
func SetOnBlur(w AttrsMapper, script string) {
	SetAttr(w, "onblur", script)
}

// SetOnFocus 设置 onfocus 属性
func SetOnFocus(w AttrsMapper, script string) {
	SetAttr(w, "onfocus", script)
}

// SetOnFormChange 设置 onformchange 属性
func SetOnFormChange(w AttrsMapper, script string) {
	SetAttr(w, "onformchange", script)
}

// SetOnFormInput 设置 onforminput 属性
func SetOnFormInput(w AttrsMapper, script string) {
	SetAttr(w, "onforminput", script)
}

// SetOnInput 设置 oninput 属性
func SetOnInput(w AttrsMapper, script string) {
	SetAttr(w, "oninput", script)
}

// SetOnInvalid 设置 oninvalid 属性
func SetOnInvalid(w AttrsMapper, script string) {
	SetAttr(w, "oninvalid", script)
}

// SetOnSubmit 设置 onsubmit 属性
func SetOnSubmit(w AttrsMapper, script string) {
	SetAttr(w, "onsubmit", script)
}

// SetOnSelect 设置 onselect 属性
func SetOnSelect(w AttrsMapper, script string) {
	SetAttr(w, "onselect", script)
}

// SetOnReset 设置 onreset 属性
func SetOnReset(w AttrsMapper, script string) {
	SetAttr(w, "onreset", script)
}

// SetOnKeyUp 设置 onkeyup 属性
func SetOnKeyUp(w AttrsMapper, script string) {
	SetAttr(w, "onkeyup", script)
}

// SetOnKeyPress 设置 onkeypress 属性
func SetOnKeyPress(w AttrsMapper, script string) {
	SetAttr(w, "onkeypress", script)
}

// SetOnKeyDown 设置 onkeydown 属性
func SetOnKeyDown(w AttrsMapper, script string) {
	SetAttr(w, "onkeydown", script)
}

// SetOnClick 设置 onclick 属性
func SetOnClick(w AttrsMapper, script string) {
	SetAttr(w, "onclick", script)
}

// StyleAttr style 属性
type StyleAttr struct {
	WithAttrs
}

// set 设置 key 的属性值为 value
func (s *StyleAttr) set(key, value string) *StyleAttr {
	attrs := s.MustAttrs()
	attr := attrs.Attr(key)
	if attr == nil {
		attr = &Attr{
			Key:    key,
			Values: []string{value},
		}
		attrs.Set(attr)
	}
	attr.Set(value)
	return s
}

func attrFirstValue(w AttrsMapper, key string) string {
	attrs := w.FindAttrs()
	if attrs == nil {
		return ""
	}
	attr := attrs.Attr(key)
	if attr == nil {
		return ""
	}
	return attr.First()
}

// Get 获取属性值
func (s *StyleAttr) Get(key string) string {
	return attrFirstValue(s, key)
}

// Width 设置宽度
func (s *StyleAttr) Width(w string) *StyleAttr {
	return s.set("width", w)
}

// MinWidth 设置最小宽度
func (s *StyleAttr) MinWidth(w string) *StyleAttr {
	return s.set("min-width", w)
}

// MaxWidth 设置最大新宽度
func (s *StyleAttr) MaxWidth(w string) *StyleAttr {
	return s.set("max-width", w)
}

// Height 设置高度
func (s *StyleAttr) Height(h string) *StyleAttr {
	return s.set("height", h)
}

// MinHeight 设置最小高度
func (s *StyleAttr) MinHeight(h string) *StyleAttr {
	return s.set("min-height", h)
}

// MaxHeight 设置最大高度
func (s *StyleAttr) MaxHeight(h string) *StyleAttr {
	return s.set("max-height", h)
}

// Color 设置前景/字体颜色
func (s *StyleAttr) Color(color string) *StyleAttr {
	return s.set("color", color)
}

// BackgroundColor 设置背景演示
func (s *StyleAttr) BackgroundColor(color string) *StyleAttr {
	return s.set("background-color", color)
}

// TextAlign 设置内容对齐方式
func (s *StyleAttr) TextAlign(align string) *StyleAttr {
	return s.set("text-align", align)
}

// Margin 设置外边距
func (s *StyleAttr) Margin(val string) *StyleAttr {
	return s.set("margin", val)
}

// Padding 设置内边距
func (s *StyleAttr) Padding(val string) *StyleAttr {
	return s.set("padding", val)
}

// Font 设置字体
func (s *StyleAttr) Font(val string) *StyleAttr {
	return s.set("font", val)
}

// FontSize 设置字体大小
func (s *StyleAttr) FontSize(val string) *StyleAttr {
	return s.set("font-size", val)
}

// FontWeight 设置字体粗细
func (s *StyleAttr) FontWeight(val string) *StyleAttr {
	return s.set("font-weight", val)
}

// FontFamily 设置字体系列（字体族）
func (s *StyleAttr) FontFamily(val string) *StyleAttr {
	return s.set("font-family", val)
}

// Border 设置边框属性
func (s *StyleAttr) Border(val string) *StyleAttr {
	return s.set("border", val)
}

// HTML 实现 Element 接口
func (s *StyleAttr) HTML() ([]byte, error) {
	attrs := s.FindAttrs()
	if attrs == nil {
		return nil, nil
	}
	return attrsHTML(attrs, ":", noQuote, "; ")
}

func noQuote(t string) string {
	return t
}

// SetTo 将样式信息设置到指定的属性集合
func (s *StyleAttr) SetTo(a AttrsMapper) error {
	code, err := s.HTML()
	if err != nil {
		return err
	}
	if len(code) == 0 {
		return nil
	}
	a.MustAttrs().MustAttr("style").Set(string(code))
	return nil
}
