//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-17

package tplfn

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"reflect"
	"strings"
	"time"
)

func Dump(value any) template.HTML {
	var bs strings.Builder
	bs.WriteString("<pre>")
	bs.WriteString(varDump(value))
	bs.WriteString("</pre>")
	return template.HTML(bs.String())
}

func varDump(v any) string {
	bf := &bytes.Buffer{}
	printValue(reflect.ValueOf(v), bf, 0)
	return bf.String()
}

func printValue(v reflect.Value, w io.Writer, indent int) {
	indentation := strings.Repeat(" ", indent)

	kindStr := fmt.Sprintf("%-10s", v.Kind().String())
	_, _ = fmt.Fprint(w, "<span style='color:blue'>"+kindStr+"</span>")
	switch v.Kind() {
	case reflect.Invalid:
		_, _ = fmt.Fprintln(w, indentation+"nil")
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		_, _ = fmt.Fprintf(w, "\t<span style='color:red'>%v</span>\n", v.Interface())
	case reflect.String:
		str := v.String()
		_, _ = fmt.Fprintf(w, "\t<span style='color:gray'>(len=%d)</span><span style='color:green'>%q</span>\n", len(str), str)
	case reflect.Struct:
		_, _ = fmt.Fprintf(w, "\t<span style='color:blue'>%s</span>\n", v.Type().String())
		if v.Type() == reflect.TypeOf(time.Time{}) {
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
		nameFmt := fmt.Sprintf("%%-%ds", maxLen+3)
		for i := 0; i < v.NumField(); i++ {
			_, _ = fmt.Fprintf(w, "%s  <span style='color:red'>"+nameFmt+"</span>", indentation, v.Type().Field(i).Name)
			printValue(v.Field(i), w, indent+maxLen)
		}
	case reflect.Array, reflect.Slice:
		_, _ = fmt.Fprintf(w, "\t(len=%d)\n", v.Len())
		for i := 0; i < v.Len(); i++ {
			printValue(v.Index(i), w, indent+1)
		}
	case reflect.Map:
		_, _ = fmt.Fprintf(w, "\t<span style='color:gray'>(len=%d)</span>\n", v.Len())
		subIndentation := indentation[:len(indentation)*4/5]
		for idx, key := range v.MapKeys() {
			_, _ = fmt.Fprintf(w, "%s    key[%d]  ", subIndentation, idx)
			printValue(key, w, indent+1)
			_, _ = fmt.Fprintf(w, "%s  value[%d]  ", subIndentation, idx)
			printValue(v.MapIndex(key), w, indent+4)
		}
	case reflect.Ptr:
		if v.IsNil() {
			_, _ = fmt.Fprintln(w, indentation+"nil pointer")
		} else {
			_, _ = fmt.Fprintln(w, indentation+"pointer to:")
			printValue(v.Elem(), w, indent+4)
		}
	default:
		_, _ = fmt.Fprintf(w, indentation+"unknown type: %v\n", v)
	}
}
