//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-05

package xi18n

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"text/template"
)

// Message 一条国际化消息
// https://cldr.unicode.org/index/cldr-spec/plural-rules
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

func (m *Message) Render(data map[string]any) (string, error) {
	rule := m.plural(data)
	switch rule {
	case pluralZero:
		if m.Zero != "" {
			return m.doRender(m.Zero, data)
		}
	case pluralOne:
		if m.One != "" {
			return m.doRender(m.One, data)
		}
	case pluralTwo:
		if m.Two != "" {
			return m.doRender(m.Two, data)
		}
	case pluralFew:
		if m.Few != "" {
			return m.doRender(m.Few, data)
		}
	case pluralMany:
		if m.Few != "" {
			return m.doRender(m.Many, data)
		}
	}
	if m.Other != "" {
		return m.doRender(m.Other, data)
	}
	return "", errors.New("msg.Other is empty")
}

func (m *Message) Render1(args []any) (string, error) {
	if len(args) != len(m.vars) {
		return "", fmt.Errorf("expect %d args, but got %d", len(m.vars), len(args))
	}
	if len(args) == 0 {
		return m.Render(nil)
	}

	data := make(map[string]any, len(args))
	for i, arg := range args {
		data[m.vars[i]] = arg
	}
	return m.Render(data)
}

func (m *Message) doRender(text string, data map[string]any) (string, error) {
	tpl, err := template.New("msg").Delims("{", "}").Parse(text)
	if err != nil {
		return "", err
	}
	bf := &bytes.Buffer{}
	err = tpl.Execute(bf, data)
	return bf.String(), err
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
