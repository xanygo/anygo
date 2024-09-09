//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-05

package xi18n

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"text/template"

	"github.com/xanygo/anygo/xsync"
)

// Message 一条本地化化消息
//
// 复数规则参考了 https://cldr.unicode.org/index/cldr-spec/plural-rules
type Message struct {
	Key string `yaml:"Key"`

	// vars 模版中用到的参数名
	vars []string

	Desc string `yaml:"Desc"`

	// Zero 0 个元素
	Zero string `yaml:"Zero"`

	// One 1 个元素
	One string `yaml:"One"`

	//  Two 2 个元素
	Two string `yaml:"Two"`

	// Few 一些，元素个数在 （2-10） 之间
	Few string `yaml:"Few"`

	// Many 很多，元素个数 >= 10
	Many string `yaml:"Many"`

	// Other 其他情况，当 Zero - Many 之间无满足条件的情况是使用
	Other string `yaml:"Other"`
}

var varReg = regexp.MustCompile(`{\.(\w+)}`)

func (m *Message) initAndCheck() error {
	if m.Key == "" {
		return errors.New("required field Key is empty")
	}
	if m.Other == "" {
		return errors.New("required field Other is empty")
	}
	sm := varReg.FindAllStringSubmatch(m.Other, -1)
	if len(sm) > 0 {
		for _, v := range sm {
			m.vars = append(m.vars, v[1])
		}
	}
	return nil
}

// RenderMap 渲染本地消息
func (m *Message) RenderMap(data map[string]any) (string, error) {
	rule := m.plural(data)
	switch rule {
	case pluralZero:
		if m.Zero != "" {
			return renderMsgMap(m.Zero, data)
		}
	case pluralOne:
		if m.One != "" {
			return renderMsgMap(m.One, data)
		}
	case pluralTwo:
		if m.Two != "" {
			return renderMsgMap(m.Two, data)
		}
	case pluralFew:
		if m.Few != "" {
			return renderMsgMap(m.Few, data)
		}
	case pluralMany:
		if m.Few != "" {
			return renderMsgMap(m.Many, data)
		}
	}
	if m.Other != "" {
		return renderMsgMap(m.Other, data)
	}
	return "", errors.New("msg.Other is empty")
}

// RenderSlice 渲染本地消息, 参数个数需要和文本模版中定义的一样
func (m *Message) RenderSlice(args ...any) (string, error) {
	if len(args) != len(m.vars) {
		return "", fmt.Errorf("expect %d args, but got %d", len(m.vars), len(args))
	}
	if len(args) == 0 {
		return m.RenderMap(nil)
	}

	data := make(map[string]any, len(args))
	for i, arg := range args {
		data[m.vars[i]] = arg
	}
	return m.RenderMap(data)
}

var bp = xsync.NewBytesBufferPool(2048)

func renderMsgMap(text string, data map[string]any) (string, error) {
	tpl, err := template.New("msg").Delims("{", "}").Parse(text)
	if err != nil {
		return "", err
	}
	bf := bp.Get()
	defer bp.Put(bf)
	err = tpl.Execute(bf, data)
	return bf.String(), err
}

func renderMsgSlice(text string, args ...any) (string, error) {
	var data map[string]any
	if len(args) > 0 {
		sm := varReg.FindAllStringSubmatch(text, -1)
		if len(sm) != len(args) {
			return "", fmt.Errorf("%q got %d args, but expect has %d", text, len(args), len(sm))
		}
		data = make(map[string]any, len(sm))
		for i, val := range sm {
			data[val[1]] = args[i]
		}
	}
	return renderMsgMap(text, data)
}

type pluralRule uint8

const (
	pluralZero pluralRule = iota
	pluralOne
	pluralTwo
	pluralFew
	pluralMany
	pluralOther
)

type ruleNumber interface {
	~int64 | ~uint64 | ~float64
}

func useRule[T ruleNumber](rule pluralRule, num T) pluralRule {
	if num == 0 {
		rule = max(rule, pluralZero)
	} else if num == 1 {
		rule = max(rule, pluralOne)
	} else if num == 2 {
		rule = max(rule, pluralTwo)
	} else if num > 2 && num < 10 {
		rule = max(rule, pluralFew)
	} else if num >= 10 {
		rule = max(rule, pluralMany)
	} else {
		rule = max(rule, pluralOther)
	}
	return rule
}

func (m *Message) plural(data map[string]any) pluralRule {
	if len(data) == 0 {
		return pluralOther
	}
	var rule pluralRule

	for _, v := range data {
		rv := reflect.ValueOf(v)
		if rv.CanInt() {
			rule = useRule(rule, rv.Int())
		} else if rv.CanUint() {
			rule = useRule(rule, rv.Uint())
		} else if rv.CanFloat() {
			rule = useRule(rule, rv.Float())
		}
	}
	return rule
}
