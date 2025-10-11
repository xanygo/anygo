//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-05

package xi18n

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Message 一条本地化化消息
//
// 复数规则参考了 https://cldr.unicode.org/index/cldr-spec/plural-rules
//
// Zero, One, Two, Few, Many, Other 是各种复数规则情况下的本地化内容。
// 这些字段可以是普通纯文本，也可以是包含变量（占位符）的，如 "Hello {0}, my name is {1}" 。
// 占位符格式为 {number}，从 0 依次递增
type Message struct {
	// Key 主键，必填
	Key string `yaml:"Key"`

	// vars 模版中用到的参数个数
	vars int

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

	// Other 其他情况，必填，当 Zero - Many 之间无满足条件的情况是使用
	Other string `yaml:"Other"`
}

var varReg = regexp.MustCompile(`{\d+}`)

func (m *Message) initAndCheck() error {
	if m.Key == "" {
		return errors.New("required field Key is empty")
	}
	if m.Other == "" {
		return errors.New("required field Other is empty")
	}
	sm := varReg.FindAllString(m.Other, -1)
	m.vars = len(sm)
	return nil
}

// Render 渲染本地消息, 参数个数需要和文本模版中定义的一样
func (m *Message) Render(args ...any) (string, error) {
	if len(args) != m.vars {
		return "", fmt.Errorf("expect %d args, but got %d", m.vars, len(args))
	}
	rule := m.plural(args...)
	switch rule {
	case pluralZero:
		if m.Zero != "" {
			return renderMsgSlice(m.Zero, args...)
		}
	case pluralOne:
		if m.One != "" {
			return renderMsgSlice(m.One, args...)
		}
	case pluralTwo:
		if m.Two != "" {
			return renderMsgSlice(m.Two, args...)
		}
	case pluralFew:
		if m.Few != "" {
			return renderMsgSlice(m.Few, args...)
		}
	case pluralMany:
		if m.Few != "" {
			return renderMsgSlice(m.Many, args...)
		}
	}
	if m.Other != "" {
		return renderMsgSlice(m.Other, args...)
	}
	return "", errors.New("msg.Other is empty")
}

func renderMsgSlice(text string, args ...any) (string, error) {
	if len(args) == 0 {
		return text, nil
	}
	for i, v := range args {
		rpk := "{" + strconv.Itoa(i) + "}"
		rpv := fmt.Sprint(v)
		text = strings.ReplaceAll(text, rpk, rpv)
	}
	strings.NewReplacer()
	return text, nil
}

type pluralRule int8

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

func (m *Message) plural(args ...any) pluralRule {
	if len(args) == 0 {
		return pluralOther
	}
	rule := pluralRule(-1)
	for _, v := range args {
		rv := reflect.ValueOf(v)
		if rv.CanInt() {
			rule = useRule(rule, rv.Int())
		} else if rv.CanUint() {
			rule = useRule(rule, rv.Uint())
		} else if rv.CanFloat() {
			rule = useRule(rule, rv.Float())
		}
	}
	if rule < pluralZero {
		return pluralOther
	}
	return rule
}
