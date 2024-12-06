//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-06

package xhtml

import (
	"fmt"
	"html/template"
	"math/rand/v2"
	"net/http"

	"github.com/xanygo/anygo/xstr"
)

type TPLHelper struct {
	Request *http.Request
}

func TemplateFuncs() template.FuncMap {
	return templateFuncs
}

var templateFuncs = template.FuncMap{
	"xInputChecked":   tlpFuncTypeValue.Checked,
	"xOptionSelected": tlpFuncTypeValue.OptionSelected,
	"xRender":         Render,

	"xRandStr": func() string {
		return xstr.RandNChar(8)
	},
	"xRandStrN": xstr.RandNChar,

	"xRandUint":   rand.Uint,
	"xRandUint32": rand.Uint32,
	"xRandUint64": rand.Uint64,

	"xRandUintN":   rand.UintN,
	"xRandUint32N": rand.Uint32N,
	"xRandUint64N": rand.Uint64N,

	"xRandInt":   rand.Int,
	"xRandInt32": rand.Int32,
	"xRandInt64": rand.Int64,

	"xRandIntN":   rand.IntN,
	"xRandInt32N": rand.Int32N,
	"xRandInt64N": rand.Int64N,

	"xRandFloat64": rand.Float64,
	"xRandFloat32": rand.Float32,
}

var tlpFuncTypeValue = tlpFuncType{}

type tlpFuncType struct {
}

func (t tlpFuncType) OptionSelected(value any) func(current any) template.HTMLAttr {
	valuesStr := fmt.Sprint(value)
	return func(current any) template.HTMLAttr {
		var selected string
		cstr := fmt.Sprint(current)
		if cstr == valuesStr {
			selected = " selected"
		}
		code := fmt.Sprintf("value=%q%s", template.HTMLEscapeString(cstr), selected)
		return template.HTMLAttr(code)
	}
}

func (t tlpFuncType) Checked(value any) func(current any) template.HTMLAttr {
	valuesStr := fmt.Sprint(value)
	return func(current any) template.HTMLAttr {
		var checked string
		cstr := fmt.Sprint(current)
		if cstr == valuesStr {
			checked = " checked"
		}
		code := fmt.Sprintf("value=%q%s", template.HTMLEscapeString(cstr), checked)
		return template.HTMLAttr(code)
	}
}
