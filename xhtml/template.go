//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-06

package xhtml

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/fs"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/xanygo/anygo/ds/xcast"
	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/ds/xurl"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xattr"
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

// WithQuery 基于当前 url，生成新的链接
//
// queryPair：url 中的 query 参数，必须成对出现，如 "a",1,"b","2","c",""
// 同名参数会将当前链接中的同名参数覆盖，值为空的则将其删除
func (t *TPLRequest) WithQuery(queryPair ...any) (template.URL, error) {
	if len(queryPair) == 0 {
		return template.URL(t.Request.URL.String()), nil
	}
	qs, err := t.toQueries(queryPair...)
	if err != nil {
		return "", err
	}
	us, err := xurl.WithQuery(t.Request.URL, qs...)
	return template.URL(us), err
}

func (t *TPLRequest) toQueries(queryPair ...any) ([]string, error) {
	qs := make([]string, 0, len(queryPair))
	for _, q := range queryPair {
		str, ok := zreflect.BaseTypeToString(q)
		if !ok {
			return nil, fmt.Errorf("invalid query %#v", q)
		}
		qs = append(qs, str)
	}
	return qs, nil
}

// WithNewQuery 基于当前 url 的 Path，生成新的链接
//
// queryPair：url 中的 query 参数，必须成对出现，如 "a",1,"b","2","c",""
// 值为空的会忽略
func (t *TPLRequest) WithNewQuery(queryPair ...any) (template.URL, error) {
	if len(queryPair) == 0 {
		return template.URL(t.Request.URL.String()), nil
	}
	qs, err := t.toQueries(queryPair...)
	if err != nil {
		return "", err
	}
	str, err := xurl.WithNewQuery(t.Request.URL, qs...)
	return template.URL(str), err
}

// QueryString 只返回 query string，会自动带上"?"
func (t *TPLRequest) QueryString(queryPair ...string) (template.URL, error) {
	qs := t.Request.URL.Query()
	for i := 0; i < len(queryPair); i += 2 {
		key := queryPair[i]
		value := queryPair[i+1]
		if value != "" {
			qs.Set(key, value)
		} else {
			qs.Del(key)
		}
	}
	if len(qs) == 0 {
		return "", nil
	}
	return template.URL("?" + qs.Encode()), nil
}

func (t *TPLRequest) EchoQueryEQ(field string, value any, echo any) any {
	query := t.getQuery()
	if query.Get(field) == zreflect.ToString(value) {
		return echo
	}
	return nil
}

// QueryEQ 判断 Query 参数是否相等
// 参数 queryPair 必须是双数，依次为： 字段名，字段值，字段名，字段值
// 字段名必须是 string 类型
func (t *TPLRequest) QueryEQ(queryPair ...any) (bool, error) {
	if len(queryPair)%2 != 0 {
		return false, fmt.Errorf("invalid query %#v", queryPair)
	}
	qs := t.getQuery()
	for i := 0; i < len(queryPair); i += 2 {
		key, ok := queryPair[i].(string)
		if !ok {
			return false, fmt.Errorf("query name [%d]=%#v not string", i, queryPair[i])
		}
		value := zreflect.ToString(queryPair[i+1])
		if qs.Get(key) != value {
			return false, nil
		}
	}
	return true, nil
}

func (t *TPLRequest) PathHas(sub string) bool {
	return strings.Contains(t.Request.URL.Path, sub)
}

func (t *TPLRequest) PathHasPrefix(prefix string) bool {
	return strings.HasPrefix(t.Request.URL.Path, prefix)
}

func (t *TPLRequest) EchoPathHasPrefix(prefix string, echo any) any {
	if strings.HasPrefix(t.Request.URL.Path, prefix) {
		return echo
	}
	return nil
}

func (t *TPLRequest) PathHasSuffix(suffix string) bool {
	return strings.HasSuffix(t.Request.URL.Path, suffix)
}

func (t *TPLRequest) EchoPathHasSuffix(suffix string, echo any) any {
	if strings.HasSuffix(t.Request.URL.Path, suffix) {
		return echo
	}
	return nil
}

func (t *TPLRequest) EchoPathHas(sub string, echo any) any {
	if strings.Contains(t.Request.URL.Path, sub) {
		return echo
	}
	return nil
}

func (t *TPLRequest) PathValue(name string) string {
	return t.Request.PathValue(name)
}

func (t *TPLRequest) PathValueHas(name string) bool {
	return t.Request.PathValue(name) != ""
}

func (t *TPLRequest) Path() string {
	return t.Request.URL.Path
}

// Dir 当前地址的父目录
func (t *TPLRequest) Dir() string {
	return path.Dir(t.Request.URL.Path)
}

func (t *TPLRequest) DirN(n int) string {
	p := t.Request.URL.Path
	for range n {
		p = path.Dir(p)
	}
	return p
}

// IsWeXin 是否微信环境
func (t *TPLRequest) IsWeXin() bool {
	return strings.Contains(t.Request.UserAgent(), "MicroMessenger/")
}

var mobileKeywords = []string{"Android", "iPhone", "iPad", "iPod"}

func (t *TPLRequest) IsMobile() bool {
	ua := t.Request.UserAgent()
	for _, k := range mobileKeywords {
		if strings.Contains(ua, k) {
			return true
		}
	}
	return false
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

	// 返回长度为 5 的可用作 id 的字符串(首字母总是英文字母，其他为字母或数字）
	"xRandID": func() string {
		return xstr.RandIdentN(5)
	},

	"xRandUint":   rand.Uint,   //  生成 uint 随机数
	"xRandUint32": rand.Uint32, // 生成 uint32 随机数
	"xRandUint64": rand.Uint64, // 生成 uint64 随机数

	"xRandUintN":   rand.UintN,   // 生成在区间 n 内的 uint   随机数, 如 {{ $num := xRandUintN  100 }}
	"xRandUint32N": rand.Uint32N, // 生成在区间 n 内的 uint32 随机数, 如 {{ $num := xRandUint32N 200 }}
	"xRandUint64N": rand.Uint64N, // 生成在区间 n 内的 uint64 随机数, 如 {{ $num := xRandUint64N 300 }}

	"xRandInt":   rand.Int,   // 生成  int 随机数
	"xRandInt32": rand.Int32, // 生成  int32 随机数
	"xRandInt64": rand.Int64, // 生成  int64 随机数

	"xRandIntN":   rand.IntN,   // 生成在区间 n 内的 int   随机数, 如 {{ $num := xRandIntN  100 }}
	"xRandInt32N": rand.Int32N, // 生成在区间 n 内的 int32   随机数, 如 {{ $num := xRandInt32N  200 }}
	"xRandInt64N": rand.Int64N, // 生成在区间 n 内的 int64   随机数, 如 {{ $num := xRandInt64N  300 }}

	"xRandFloat64": rand.Float64, // 生成 float64 随机数
	"xRandFloat32": rand.Float32, // 生成 float32 随机数

	// 通过输入的 pair 创建一个 map[string]any 类型的 map，
	// 如 {{ $obj := xNewMap "k1" "v1" "k2" 100 }}, 会生成 map：$obj = {"k1" : "v1", "k2" : 100 }
	"xNewMap": xmap.Create[string, any],

	// 返回所有 map 的 keys( 结果是 []any 类型)
	"xMapKeys": tplfn.MapKeys,

	// xNewSlice 创建一个 []any 类型的 slice，如 {{ $arr := xNewSlice 1 2 3 }}
	"xNewSlice": func(arr ...any) []any {
		return arr
	},

	// 若传入的 value 不为空，则返回自身。否则返回一个空的 map[sting]any
	"xOrMap": tplfn.OrMap,

	// 将 time.Time 类型的值，格式化输出为 2006-01-02 15:04:05。
	// 若 Time.IsZero,则会输出空字符串
	"xDateTime": tplfn.DateTime,

	// 格式输出当前的时间，需要传入 format，如 {{ xNowFormat "2006" }} -> 2026
	"xNowFormat": tplfn.NowTimeFormat,

	// 对输入的参数，创建一个依次轮询的顺序迭代器
	// 如 {{ $iter := xEachOfIter "a" "b" "c" }}
	//  {{ range $index,$item:= .Items }}
	//    {{ $item.Value}}
	//    {{ $iter.Next }} // 依次输出 "a" "b" "c"
	//  {{ end }}
	"xEachOfIter": tplfn.EachOfIter,

	// 对输入的参数，创建一个随机迭代器
	// 如 {{ $iter := xRandOfIter "a" "b" "c" }}
	//  {{ range $index,$item:= .Items }}
	//    {{ $item.Value}}
	//    {{ $iter.Next }} // 随机输出 "a" "b" "c"
	//  {{ end }}
	"xRandOfIter": tplfn.RandOfIter,

	// 将对象以 JSON 编码并输出，未加 Indent 格式对齐
	"xJSON": func(val any) (string, error) {
		bf, err := json.Marshal(val)
		return string(bf), err
	},

	// 将对象以 JSON 编码，并添加 Indent 格式对齐
	"xJSONIndent": func(val any) (string, error) {
		bf, err := json.MarshalIndent(val, "", "  ")
		return string(bf), err
	},

	// 丢弃 json tag 属性，struct 会转换为 map。
	// slice、array、map 等为 struct 类型会原样返回
	// 用于调试，更方便打印出数据的原始结构信息
	// 如 {{ xJSONIndent ( xToPlainObject $item ) }}
	"xToPlainObject": zreflect.ToPlainObject,

	// xDump 调试时打印出数据内容。如 {{ xDump $item }}
	"xDump": tplfn.Dump,

	"xIsOdd":  tplfn.IsOddNumber,  //  判断是否是奇数
	"xIsEven": tplfn.IsEvenNumber, // 判断是否是偶数
	"xModEQ":  tplfn.IsRemainder,  // 判断余数是否指定值

	// 将字符串转换为 HTML
	"xHTML": func(str string) template.HTML {
		return template.HTML(str)
	},
	// 将字符串转换为 html 属性
	"xHTMLAttr": func(str string) template.HTMLAttr {
		return template.HTMLAttr(str)
	},

	// 将字符串转换为 css 代码
	"xCss": func(str string) template.CSS {
		return template.CSS(str)
	},

	// 将字符串转换为 js 代码
	"xJs": func(str string) template.JS {
		return template.JS(str)
	},

	// 将字符串转换为 url
	"xURL": func(str string) template.URL {
		return template.URL(str)
	},

	// 生成从 [start,end] 区间的 []int ，如 {{  $arr := xNewInts 1 3 }}
	"xNewInts": func(start int, end int) []int {
		result := make([]int, 0, end-start)
		for i := start; i < end; i++ {
			result = append(result, i)
		}
		return result
	},

	// 生成从 [start, end] 区间,间隔步长为 step 的 []int ，如 {{  $arr := xNewIntsStep 1 3 2 }}
	"xNewIntsStep": func(start int, end int, step int) []int {
		result := make([]int, 0, (end-start)/step+1)
		for i := start; i < end; i += step {
			result = append(result, i)
		}
		return result
	},

	"xStrPrefix":   strings.HasPrefix, // 判断字符串是否包含指定前缀 ，如 {{ if  xStrPrefix $name "han" }}
	"xStrSuffix":   strings.HasSuffix, // 判断字符串是否包含指定后缀 ，如 {{ if  xStrSuffix $name "mei" }}
	"xStrContains": strings.Contains,  // 判断字符串是否包含指定子串 ，如 {{ if  xStrContains $name "mei" }}
	"xStrSplit":    strings.Split,     // 将字符串使用子串分割为 []string, 如 {{ $arr := xStrSplit $log "\n" }}
	"xStrFields":   strings.Fields,    // 将字符串使用空白字符分割为 []string, 如 {{ $arr := xStrFields $log }}
	"xStrCount":    strings.Count,     // 统计字符串中子串的数量，{{ xStrCount $name "mei" }}

	// 读取使用 SetConst 设置的常量值
	"xConst": getConst,

	// 检查传入的参数是否是非空值，若是空值则会报错
	"xAssert": tplfn.Assert,

	// 将 array 或者 slice,使用 连接符链接为一个字符串。
	// 如 {{ $str := xJoin $arr "-" }}
	"xJoin": tplfn.Join,

	"xMathAdd":        tplfn.MathAdd,        // 数学运算，加法，如 {{ $num := xMathAdd $score 1 }}
	"xMathSub":        tplfn.MathSub,        // 数学运算，减法，如 {{ $num := xMathSub $score 1 }}
	"xMathMul":        tplfn.MathMul,        // 数学运算，乘法，如 {{ $num := xMathMul $score 2 }}
	"xMathDiv":        tplfn.MathDiv,        // 数学运算，除法，如 {{ $num := xMathDiv $score 3 }}
	"xMathPercent":    tplfn.MathPercent,    // 将一个小数转换为百分比的字符串，如 $score=0.1, {{ xMathPercent $score }} -> " 10.000%"
	"xMathComplement": tplfn.MathComplement, // 将小数转换为剩余百分比，即  (1-f)*100 %， $score=0.1, {{ xMathComplement $score }} -> " 90.000%"

	"xInt":     xcast.IntegerE[int],    // 将数值类型的数值转换为 int，失败会报错
	"xInt8":    xcast.IntegerE[int8],   // 将其他类型的数值转换为 int8，失败会报错
	"xInt16":   xcast.IntegerE[int16],  // 将其他类型的数值转换为 int16，失败会报错
	"xInt32":   xcast.IntegerE[int32],  // 将其他类型的数值转换为 int32，失败会报错
	"xInt64":   xcast.IntegerE[int64],  // 将其他类型的数值转换为 int64，失败会报错
	"xUInt":    xcast.IntegerE[uint],   // 将其他类型的数值转换为 uint，失败会报错
	"xUInt8":   xcast.IntegerE[uint8],  // 将其他类型的数值转换为 uint8，失败会报错
	"xUInt16":  xcast.IntegerE[uint16], // 将其他类型的数值转换为 uint16，失败会报错
	"xUInt32":  xcast.IntegerE[uint32], // 将其他类型的数值转换为 uint32，失败会报错
	"xUInt64":  xcast.IntegerE[uint64], // 将其他类型的数值转换为 uint64，失败会报错
	"xFloat32": xcast.FloatE[float32],  // 将其他类型的数值转换为 float32，失败会报错
	"xFloat64": xcast.FloatE[float64],  // 将其他类型的数值转换为 float64，失败会报错

	// 将多个字符串连接在一起
	"xCat": func(items ...string) string {
		if len(items) == 0 {
			return ""
		}
		return strings.Join(items, "")
	},
	"xToLower":   strings.ToLower, // 将字符串转换为小写
	"xToUpper":   strings.ToUpper, // 将字符串转换为大写
	"xToTitle":   strings.ToTitle,
	"xTrimSpace": strings.TrimSpace, // 移除字符串首尾的空白
	"xTrim":      strings.Trim,      // 移除字符串首尾的子串

	"xnl2br": tplfn.NL2BR, // 将换行符转换为 <br>

	// 统计字符串的行数( \n 的个数)，返回值不小于 min
	"xMinLines": func(min int, str string) int {
		n := strings.Count(str, "\n") + 1
		return max(min, n)
	},

	"xPathDir": path.Dir,
	"xPathDirN": func(p string, n int) string {
		for range n {
			p = path.Dir(p)
		}
		return p
	},
	"xPathClean": path.Clean,
	"xPathJoin":  path.Join,
	"xPathBase":  path.Base,
	"xPathIsAbs": path.IsAbs,
	"xPathExt":   path.Ext,

	"xFilePathToSlash": filepath.ToSlash,

	// 三元表达式，如 {{ xTernary $ok "Ok-Value" "Else Value" }}
	"xTernary": func(ok bool, x any, y any) any {
		if ok {
			return x
		}
		return y
	},

	"xSliceContains": zreflect.SliceContains, // 判断 slice 或者 array 是否包含特定的值

	"xIsDebugMode": xattr.IsDebugMode, // 判断当前应用是否调试模式

	"xMapHasKey": zreflect.MapHasKey,
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

func WithFuncMap(tpl *template.Template) *template.Template {
	tpl = tpl.Funcs(FuncMap)
	m1 := make(template.FuncMap, len(AdvancedFuncMap))
	for key, fn := range AdvancedFuncMap {
		m1[key] = fn(tpl)
	}
	return tpl.Funcs(m1)
}

// AdvancedFuncMap 支持在模版函数中读取到 Template 对象的更高级的辅助函数
var AdvancedFuncMap = map[string]func(tpl *template.Template) any{
	"xRenderEscaped": func(tmpl *template.Template) any {
		return func(name string, values ...any) (template.HTML, error) {
			var data any
			switch len(values) {
			case 0:
			case 1:
				data = values[0]
			default:
				return "", errors.New("too many values")
			}
			buf := xsync.GetBytesBuffer()
			defer xsync.PutBytesBuffer(buf)
			if err := tmpl.ExecuteTemplate(buf, name, data); err != nil {
				return "", err
			}
			return template.HTML(html.EscapeString(buf.String())), nil
		}
	},
}

// TemplateCore 用于 WalkParseFS 方法的约束，*html/template.Template 和 *text/template.Template 都是满足的
type TemplateCore[T any] interface {
	New(name string) T
	Parse(string) (T, error)
}

// WalkParseFS 遍历读取 fsys ，并将符合 pattern 的文件解析
//
// patterns: 文件名的规则，可选，不能包含目录，有效值如 *.html。若为空，则解析所有文件
//
// 注意：
//  1. 所有 define 定义的块，全局应该不出现重名，在使用 template 方法渲染的时候，不应该添加其所在目录，
//     如在 user/index.html 文件中有 {{ define "status.part" }} Ok {{ end }},
//     不管是在那个目录的那个文件，渲染该块，都不能添加目录： {{ template "status.part" .User }}
//  2. 使用 template 渲染的时候，需要使用完整的路径，如 {{ template "user/index.html" . }}
func WalkParseFS[T TemplateCore[T]](t T, fsys fs.FS, root string, patterns ...string) (emp T, err error) {
	sub, err := fs.Sub(fsys, root)
	if err != nil {
		return emp, err
	}
	fsys = sub
	parserOne := func(filename string) error {
		content, err1 := fs.ReadFile(fsys, filename)
		if err1 != nil {
			return err1
		}
		tmpl := t.New(filename)
		_, err1 = tmpl.Parse(string(content))
		return err1
	}

	err = fs.WalkDir(fsys, ".", func(fp string, d fs.DirEntry, err error) error {
		if fp == root {
			return nil
		}
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if len(patterns) == 0 {
			return parserOne(fp)
		}
		for _, pattern := range patterns {
			if ok, _ := path.Match(pattern, path.Base(fp)); !ok {
				continue
			}
			return parserOne(fp)
		}
		return nil
	})
	return t, err
}
