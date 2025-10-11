//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xlog

import (
	"path/filepath"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xattr"
)

var defaultPanicLogger xsync.Value[Logger]

// PanicLogger 用于打印 panic 信息的 logger，若没有设置，则返回 Default()
func PanicLogger() Logger {
	lg := defaultPanicLogger.Load()
	if lg == nil {
		return Default()
	}
	return lg
}

// SetPanicLogger 设置 panic logger
func SetPanicLogger(lg Logger) {
	defaultPanicLogger.Store(lg)
}

// PanicLoggerOpt 用于打印 panic 日志的配置
// 配置内容:
// 1. 首选日志配置文件: conf/log/panic
// 2. 若无则使用默认日志文件地址： log/panic/panic.log
func PanicLoggerOpt() FileLoggerOpt {
	return FileLoggerOpt{
		CfgPath: filepath.Join(xattr.ConfDir(), "log", "panic"),
		Cfg: &FileConfig{
			FileName: filepath.Join(xattr.LogDir(), "panic", "panic.log"),
			ExtRule:  "1hour",
			MaxFiles: 48,
		},
	}
}
