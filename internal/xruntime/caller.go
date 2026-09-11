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
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	if n == 0 {
		return "", 0, ""
	}

	frames := runtime.CallersFrames(pcs[:n])

	var foundPanic bool
	var goRootDir string
	for {
		frame, more := frames.Next()
		fileName := frame.File
		if !foundPanic {
			//  go build 添加 -trimpath 后，
			// 输出 fileName= runtime/panic.go ,lineNo= 855
			foundPanic = strings.Contains(fileName, "runtime/panic.go")
			if foundPanic && strings.Contains(fileName, "/src/runtime/") {
				// 若有 -trimpath，则 -trimpath 为空
				goRootDir = fileName[:(len(fileName) - len("runtime/panic.go"))]
			}
		} else {
			if goRootDir == "" {
				if !strings.HasPrefix(fileName, "internal/") && !strings.HasPrefix(fileName, "runtime/") {
					return fileName, frame.Line, frame.Function
				}
			} else {
				if !strings.HasPrefix(fileName, goRootDir) {
					return fileName, frame.Line, frame.Function
				}
			}
		}

		if !more {
			break
		}
	}
	return "", 0, ""
}

const selfPkg = "github.com/xanygo/anygo/"

// CallerPC 定位到排除框架外的 api 调用的地址
func CallerPC(skip int) uintptr {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	if n == 0 {
		return 0
	}
	frames := runtime.CallersFrames(pcs[:n])
	var found bool
	for {
		frame, more := frames.Next()
		if !found {
			found = strings.Contains(frame.File, selfPkg)
		} else if !strings.Contains(frame.File, selfPkg) {
			// 返回紧挨着框架文的下一个文件
			return frame.PC
		}
		if !more {
			break
		}
	}
	return 0
}
