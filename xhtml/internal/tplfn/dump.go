//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-17

package tplfn

import (
	"bytes"
	"cmp"
	"fmt"
	"html/template"
	"io"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/xanygo/anygo/ds/xsync"
)

func Dump(value any) template.HTML {
	var bs strings.Builder
	bs.WriteString("<pre class='x-dump'>\n")
	bs.WriteString(varDump(value))
	bs.WriteString("</pre>")
	return template.HTML(bs.String())
}

func varDump(v any) string {
	bf := xsync.GetBytesBuffer()
	defer xsync.PutBytesBuffer(bf)
	printValue(reflect.ValueOf(v), bf, 0, "")
	return bf.String()
}

var bgKindColor = map[reflect.Kind]string{
	reflect.Array:   "background-color:#cce5ff;margin-top:5px",
	reflect.Slice:   "background-color:#d4edda;margin-top:5px",
	reflect.Struct:  "background-color:#fff3cd;margin-top:5px",
	reflect.Map:     "background-color:#ffc107;margin-top:5px",
	reflect.Pointer: "background-color:#f8d7da;margin-top:5px",
}

func printValue(v reflect.Value, w io.Writer, indent int, prefix string) {
	indentation := strings.Repeat(" ", indent)

	var typeStr string
	if v.IsValid() {
		typeStr = strings.ReplaceAll(v.Type().String(), "interface {}", "any")
	} else {
		typeStr = v.Kind().String()
	}
	typeStr += "  "

	code := indentation + prefix + "<span style='color:blue'>" + typeStr + "</span>"
	switch v.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map, reflect.Pointer:
		style := bgKindColor[v.Kind()]
		_, _ = fmt.Fprint(w, "<div style='"+style+"'>"+code) // div 的结尾在具体的类型逻辑里
	default:
		_, _ = fmt.Fprint(w, code)
	}

	switch v.Kind() {
	case reflect.Invalid:
		_, _ = fmt.Fprintln(w, indentation+"nil")
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		if v.CanInterface() {
			_, _ = fmt.Fprintf(w, "\t<span style='color:red'>%v</span>\n", v.Interface())
		} else {
			_, _ = fmt.Fprintf(w, "\t<span style='color:red'>%v</span>\n", v)
		}
	case reflect.String:
		str := v.String()
		_, _ = fmt.Fprintf(w, "\t<span style='color:gray'>(%d)</span><span style='color:green'>%q</span>\n", len(str), str)
	case reflect.Struct:
		fmt.Fprint(w, "</div>")
		if v.Type() == reflect.TypeFor[time.Time]() {
			tm := v.Interface().(time.Time)
			_, _ = fmt.Fprintf(w, indentation+"  Time.String   : %s\n", tm.String())
			_, _ = fmt.Fprintf(w, indentation+"  Time.UnixNano : %d\n", tm.UnixNano())
			_, _ = fmt.Fprintf(w, indentation+"  Time.Unix     : %d\n", tm.Unix())
			return
		}
		var maxLen int
		for i := 0; i < v.NumField(); i++ {
			maxLen = max(maxLen, len(v.Type().Field(i).Name))
		}
		fmt.Fprint(w, "<div data-dump-type='struct'>")
		nameFmt := fmt.Sprintf("%%-%ds", maxLen+3)
		for i := 0; i < v.NumField(); i++ {
			pp := fmt.Sprintf("<span style='color:gray'>[%d]</span><span style='color:red'>"+nameFmt+"</span>", i, v.Type().Field(i).Name)
			printValue(v.Field(i), w, indent+2, pp)
		}
		fmt.Fprint(w, "</div>")
	case reflect.Array, reflect.Slice:
		_, _ = fmt.Fprintf(w, "\t(len=%d)</div><div data-dump-type='array'>", v.Len())
		for i := 0; i < v.Len(); i++ {
			printValue(v.Index(i), w, indent+2, fmt.Sprintf("<span style='color:gray'>[%d]</span>", i))
		}
		fmt.Fprint(w, "</div>")
	case reflect.Map:
		// tt := "<span style='color:blue'>" + strings.ReplaceAll(v.Type().String(), "interface {}", "any") + "</span>"
		_, _ = fmt.Fprintf(w, "&nbsp;<span style='color:gray'>(len=%d)</span></div><div data-dump-type='map'>", v.Len())
		subIndentation := indentation[:len(indentation)*4/5]

		bw := xsync.GetBytesBuffer()
		keys := v.MapKeys()
		slices.SortFunc(keys, func(a, b reflect.Value) int {
			return cmp.Compare(a.String(), b.String())
		})
		for idx, key := range keys {
			_, _ = fmt.Fprintf(w, "%s  (%d)key    ", subIndentation, idx)
			bw.Reset()
			printValue(key, bw, indent+2, "")
			w.Write(bytes.TrimLeftFunc(bw.Bytes(), unicode.IsSpace))

			_, _ = fmt.Fprintf(w, "%s  (%d)value  ", subIndentation, idx)

			bw.Reset()
			printValue(v.MapIndex(key), bw, indent+2, "")
			w.Write(bytes.TrimLeftFunc(bw.Bytes(), unicode.IsSpace))

			w.Write([]byte("\n"))
		}
		fmt.Fprint(w, "</div>")
		xsync.PutBytesBuffer(bw)
	case reflect.Pointer:
		if v.IsNil() {
			_, _ = fmt.Fprintln(w, "\tnil pointer</div>")
		} else {
			_, _ = fmt.Fprintf(w, "\t%s</div>\n", v.Type().String())
			_, _ = fmt.Fprintln(w, "<div data-dump-type='pointer'>")
			printValue(v.Elem(), w, indent+2, "")
			_, _ = fmt.Fprintln(w, "</div>")
		}
	default:
		if v.CanInterface() {
			vvr := reflect.ValueOf(v.Interface())
			printValue(vvr, w, indent+2, "")
		} else {
			_, _ = fmt.Fprintf(w, "\t<span>%v</span>\n", v)
		}
	}
}
