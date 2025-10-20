//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-06

package xhtml

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/ds/xurl"
	"github.com/xanygo/anygo/internal/zbase"
	"github.com/xanygo/anygo/xhtml/internal/tplfn"
)

func NewTPLRequest(req *http.Request) *TPLRequest {
	return &TPLRequest{
		Request: req,
	}
}

type TPLRequest struct {
	Request *http.Request
	once    xsync.OnceDoValue[url.Values]
}

func (t *TPLRequest) Context() context.Context {
	return t.Request.Context()
}

func (t *TPLRequest) getQuery() url.Values {
	return t.once.Do(t.Request.URL.Query)
}

// Query 获取 url 的 query 参数值
func (t *TPLRequest) Query(field string) string {
	return t.getQuery().Get(field)
}

func (t *TPLRequest) QueryIn(field string, values []string) bool {
	value := t.getQuery().Get(field)
	return slices.Contains(values, value)
}

// BaseLink 基于当前 url，生成新的链接
//
// queryPair：url 中的 query 参数，必须成对出现，如 "a",1,"b","2","c",""
// 同名参数会将当前链接中的同名参数覆盖，值为空的则将其删除
func (t *TPLRequest) BaseLink(queryPair ...any) template.URL {
	return template.URL(xurl.BaseLink(t.Request.URL, queryPair...))
}

// NewLink 基于当前 url 的 Path，生成新的链接
//
// queryPair：url 中的 query 参数，必须成对出现，如 "a",1,"b","2","c",""
// 值为空的会忽略
func (t *TPLRequest) NewLink(queryPair ...any) template.URL {
	return template.URL(xurl.NewLink(t.Request.URL, queryPair...))
}

func (t *TPLRequest) EchoQueryEQ(field string, value any, echo any) any {
	query := t.getQuery()
	if query.Get(field) == zbase.ToString(value) {
		return echo
	}
	return nil
}

// QueryEQ 判断 Query 参数是否相等
// 参数 queryPair 必须是双数，依次为： 字段名，字段值，字段名，字段值
// 字段名必须是 string 类型
func (t *TPLRequest) QueryEQ(queryPair ...any) bool {
	if len(queryPair)%2 != 0 {
		return false
	}
	qs := t.getQuery()
	for i := 0; i < len(queryPair); i += 2 {
		key := queryPair[i].(string)
		value := zbase.ToString(queryPair[i+1])
		if qs.Get(key) != value {
			return false
		}
	}
	return true
}

func (t *TPLRequest) PathHas(sub string) bool {
	return strings.Contains(t.Request.URL.Path, sub)
}

func (t *TPLRequest) PathHasPrefix(sub string) bool {
	return strings.HasPrefix(t.Request.URL.Path, sub)
}

func (t *TPLRequest) PathHasSuffix(sub string) bool {
	return strings.HasSuffix(t.Request.URL.Path, sub)
}

func (t *TPLRequest) EchoPathHas(sub string, echo any) any {
	if strings.Contains(t.Request.URL.Path, sub) {
		return echo
	}
	return nil
}

// FuncMap 用于模版的辅助方法
var FuncMap = template.FuncMap{
	// 渲染一个 Element 为 HTML 字符串
	"xRender": render,

	// 用于 type="check" 类型的 input 的 value 和 checked 属性输出
	"xCheckedValue": tplfn.InputChecked,

	// 用于 option 类型的 input 的 value 和 checked 属性输出
	"xSelectedValue": tplfn.InputSelected,

	// 连接多个参数合并为 input 的 name，
	// 如 name='{{ xInputObjectName “widget" "name" }}' -> name='widget[name]'
	"xInputObjectName": tplfn.InputObjectName,

	// xRandStr 返回一个长度为 8 的随机字符串
	"xRandStr": func() string {
		return xstr.RandNChar(8)
	},

	// 返回指定长度的字符串， 如 {{ xRandStrN 10 }}
	"xRandStrN": xstr.RandNChar,

	// 返回指定长度的可用作 id 的字符串(首字母总是英文字母，其他为字母或数字）， 如 {{ xRandIDN 10 }}
	"xRandIDN": xstr.RandIdentN,

	// 返回长度为5 的可用作 id 的字符串(首字母总是英文字母，其他为字母或数字）
	"xRandID": func() string {
		return xstr.RandIdentN(5)
	},

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

	// 通过输入的 pair 创建一个 map，
	// 如 {{ $obj := xNewMap "k1" "v1" "k2" 100 }}, 会生成map：$obj = {"k1" : "v1", "k2" : 100 }
	// 创建的是 map[string]any 类型的 map
	"xNewMap":  xmap.Create[string, any],
	"xMapKeys": tplfn.MapKeys,

	// 若传入的 value 不为空，则返回自身。否则返回一个空的 map[sting]any
	"xOrMap": tplfn.OrMap,

	"xDateTime":  tplfn.DateTime,
	"xNowFormat": tplfn.NowTimeFormat,

	// 对输入的参数，创建一个依次轮询的顺序迭代器
	// 如 {{ $iter := xEachOfIter "a" "b" "c" }}
	//  {{ range $index,$item:= .Items }}
	//    {{ $item.Value}}
	//    {{ $item.Next }} // 依次输出 "a" "b" "c"
	//  {{ end }}
	"xEachOfIter": tplfn.EachOfIter,

	// 对输入的参数，创建一个随机迭代器
	// 如 {{ $iter := xRandOfIter "a" "b" "c" }}
	//  {{ range $index,$item:= .Items }}
	//    {{ $item.Value}}
	//    {{ $iter.Next }} // 随机输出 "a" "b" "c"
	//  {{ end }}
	"xRandOfIter": tplfn.RandOfIter,

	"xJSON": func(val any) (string, error) {
		bf, err := json.Marshal(val)
		return string(bf), err
	},
	"xJSONIndent": func(val any) (string, error) {
		bf, err := json.MarshalIndent(val, "", "  ")
		return string(bf), err
	},

	"xDump": tplfn.Dump,

	"xIsOdd":  tplfn.IsOddNumber,  //  判断是否是奇数
	"xIsEven": tplfn.IsEvenNumber, // 判断是否是偶数
	"xModEQ":  tplfn.IsRemainder,  // 判断余数是否指定值

	"xHTML": func(str string) template.HTML {
		return template.HTML(str)
	},
	"xHTMLAttr": func(str string) template.HTMLAttr {
		return template.HTMLAttr(str)
	},
	"xCss": func(str string) template.CSS {
		return template.CSS(str)
	},
	"xJs": func(str string) template.JS {
		return template.JS(str)
	},

	"xNewInts": func(start int, end int) []int {
		result := make([]int, 0, end-start)
		for i := start; i < end; i++ {
			result = append(result, i)
		}
		return result
	},
	"xNewIntsStep": func(start int, end int, step int) []int {
		result := make([]int, 0, end-start)
		for i := start; i < end; i += step {
			result = append(result, i)
		}
		return result
	},

	"xStrPrefix":   strings.HasPrefix,
	"xStrSuffix":   strings.HasSuffix,
	"xStrContains": strings.Contains,
	"xStrSplit":    strings.Split,
	"xStrFields":   strings.Fields,

	// 读取使用 SetConst 设置的常量值
	"xConst": getConst,

	// 检查传入的参数是否是非空值
	"xAssert": tplfn.Assert,

	"xJoin": tplfn.Join,

	"xMathAdd":        tplfn.MathAdd,
	"xMathSub":        tplfn.MathSub,
	"xMathMul":        tplfn.MathMul,
	"xMathDiv":        tplfn.MathDiv,
	"xMathPercent":    tplfn.MathPercent,    // 将一个小数转换为百分比的字符串
	"xMathComplement": tplfn.MathComplement, // (1-f)*100 %

	"xCat": func(items ...string) string {
		if len(items) == 0 {
			return ""
		}
		return strings.Join(items, "")
	},
	"xToLower":   strings.ToLower,
	"xToUpper":   strings.ToUpper,
	"xToTitle":   strings.ToTitle,
	"xTrimSpace": strings.TrimSpace,
	"xTrim":      strings.Trim,

	"xnl2br": tplfn.NL2BR,
}

func Dump(w io.Writer, obj any) {
	code := tplfn.Dump(obj)
	bf := unsafe.Slice(unsafe.StringData(string(code)), len(code))
	if hw, ok := w.(http.ResponseWriter); ok {
		hw.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	_, _ = w.Write(bf)
}

var constVars sync.Map

func getConst(key string, def ...any) any {
	val, ok := constVars.Load(key)
	if ok || len(def) == 0 {
		return val
	}
	return def[0]
}

// SetConst 定义/存储一个值，用于在模版中使用 xConst 方法读取到
func SetConst(key string, val any) {
	constVars.LoadOrStore(key, val)
}

func init() {
	const patternUri = `pattern="^(((https|http):\/\/\S+(\.\S+)+.*)|(\/\S+))$"`
	SetConst("pattern-uri", template.HTMLAttr(patternUri))
	SetConst("pid", os.Getpid())
	SetConst("ppid", os.Getppid())
	SetConst("startTime", time.Now())
}
