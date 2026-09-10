//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-27

package xruntime

import (
	"runtime"
	"strings"
)

// PanicCaller 查找触发 panic 的文件和函数名
func PanicCaller(skip int) (file string, line int, fn string) {
	pc := make([]uintptr, 10)
	n := runtime.Callers(skip, pc)
	var foundPanic bool
	for i := range n {
		fn := runtime.FuncForPC(pc[i])
		fileName, lineNo := fn.FileLine(pc[i])
		//
		// 查找到下面这几行
		//  panic({0x7ff6b7018460?, 0x14abf27deaa0?})
		// 	C:/soft/go/src/runtime/panic.go:859 +0x125
		// github.com/xanygo/anygo/xkv/xkvx.MustLoad[...]({0x7ff6b6718602?, 0x3})
		//
		isPanicFile := len(fileName) > 16 && strings.Contains(fileName, "runtime") && strings.Contains(fileName, "panic.go:")
		if foundPanic && !isPanicFile {
			return fileName, lineNo, fn.Name()
		}
		foundPanic = isPanicFile
	}
	return "", 0, ""
}
