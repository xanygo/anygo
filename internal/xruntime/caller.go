//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-27

package xruntime

import (
	"path/filepath"
	"runtime"
	"strings"
)

var filePanicName = filepath.Join("src", "runtime", "panic.go")

// PanicCaller 查找触发 panic 的文件和函数名
func PanicCaller(skip int) (file string, line int, fn string) {
	pc := make([]uintptr, 10)
	n := runtime.Callers(skip, pc)
	var foundPanic bool
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pc[i])
		fileName, lineNo := fn.FileLine(pc[i])
		isPanicFile := strings.HasSuffix(fileName, filePanicName)
		if foundPanic && !isPanicFile {
			return fileName, lineNo, fn.Name()
		}
		foundPanic = isPanicFile
	}
	return "", 0, ""
}
