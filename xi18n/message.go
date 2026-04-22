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

	"github.com/xanygo/anygo/ds/xsync"
)

// Message 一条本地化化消息
//
// 复数规则参考了 https://cldr.unicode.org/index/cldr-spec/plural-rules
//
//	Zero, One, Two, Few, Many, Other 是各种复数规则情况下的本地化内容。
//	这些字段可以是普通纯文本，也可以是包含变量（占位符）的，如 "Hello {0}, my name is {1}" 。
//	占位符格式为 {number}，从 0 依次递增。
type Message struct {
	// Key 主键，必填，可以包含 "/"，会当做目录结构
	Key string `yaml:"Key"`

	// vars 模版中用到的参数个数，是 Zero、One、Two、Few、Many、Other 这几个字段使用参数个数的最大值
	// 如 Zero=“zero books” 使用 0 个参数，而 Other=" {0} books" 使用 1 个参数
	vars int

	// Desc 描述信息，可选
	Desc string `yaml:"Desc"`

	// Zero 参数数值为 0 的情况下使用的模版
	// 如： zero books
	Zero string `yaml:"Zero"`

	// One 参数数值为 1 的情况下使用的模版
	// 如： one book
	One string `yaml:"One"`

	//  Two 参数数值为 2 的情况下使用的模版
	// 如： {0} books
	Two string `yaml:"Two"`

	// Few 参数数值为 (2-10) 的情况下使用的模版
	Few string `yaml:"Few"`

	// Many 参数数值为 >=10 的情况下使用的模版
	Many string `yaml:"Many"`

	// Other 其他情况，必填，当 Zero - Many 之间无满足条件的情况是使用
	// 如： {0} books
	Other string `yaml:"Other"`

	once xsync.OnceDoErr
}

var varReg = regexp.MustCompile(`{\d+}`)

func (m *Message) doInit() error {
	return m.once.Do(m.initAndCheck)
}

func (m *Message) initAndCheck() error {
	if m.Key == "" {
		return errors.New("required field Key is empty")
	}
	if m.Other == "" {
		return errors.New("required field Other is empty")
	}
	for _, str := range []string{m.Zero, m.One, m.Two, m.Few, m.Many, m.Other} {
		if str == "" {
			continue
		}
		sm := varReg.FindAllString(str, -1)
		m.vars = max(m.vars, len(sm))
	}
	return nil
}

// Render 渲染本地消息, 参数个数需要和文本模版中定义的一样
func (m *Message) Render(args ...any) (string, error) {
	if err := m.doInit(); err != nil {
		return "", err
	}
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
		if m.Many != "" {
			return renderMsgSlice(m.Many, args...)
		}
	}
	return renderMsgSlice(m.Other, args...)
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
