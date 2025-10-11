//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-21

package xhtml

// NewInput 创建一个 input 标签
func NewInput(tp string, name string) *Any {
	input := &Any{
		Tag:       "input",
		SelfClose: true,
	}
	SetType(input, tp)
	if name != "" {
		SetName(input, name)
	}
	return input
}

func NewInputText(name string) *Any {
	return NewInput("text", name)
}

func NewInputPassword(name string) *Any {
	return NewInput("password", name)
}

// NewInputSubmit 创建一个 submit 标签
func NewInputSubmit(value string) *Any {
	s := NewInput("submit", "")
	if value != "" {
		SetValue(s, value)
	}
	return s
}

// NewInputButton 创建一个 button 标签
func NewInputButton(value string) *Any {
	s := NewInput("button", "")
	if value != "" {
		SetValue(s, value)
	}
	return s
}

func NewInputRadio(name string, value string, checked bool) *Any {
	r := NewInput("radio", name)
	SetValue(r, value)
	SetChecked(r, checked)
	return r
}

func NewInputCheckbox(name string, value string, checked bool) *Any {
	r := NewInput("checkbox", name)
	SetValue(r, value)
	SetChecked(r, checked)
	return r
}

func NewInputNumber(name string) *Any {
	return NewInput("number", name)
}

func NewInputData(name string) *Any {
	return NewInput("date", name)
}

func NewInputColor(name string) *Any {
	return NewInput("color", name)
}

func NewInputRange[N Number](name string, min N, max N, step N) *Any {
	input := NewInput("range", name)
	SetMin(input, min)
	SetMax(input, max)
	SetStep(input, step)
	return input
}

// NewInputMonth 选择月份和年份,使用浏览器的日期选择器,选择后得到如 2024-10
func NewInputMonth(name string) *Any {
	return NewInput("month", name)
}

// NewInputWeek 选择周和年
func NewInputWeek(name string) *Any {
	return NewInput("week", name)
}

// NewInputTime 选择时间（无时区），得到如 12:02
func NewInputTime(name string) *Any {
	return NewInput("time", name)
}

// NewInputDateTimeLocal 选择时间（无时区），得到如 2024-10-21T13:07
func NewInputDateTimeLocal(name string) *Any {
	return NewInput("datetime-local", name)
}

func NewInputEmail(name string) *Any {
	return NewInput("email", name)
}

func NewInputSearch(name string) *Any {
	return NewInput("search", name)
}

// NewInputURL 包含 URL 地址的输入字段
func NewInputURL(name string) *Any {
	return NewInput("url", name)
}
