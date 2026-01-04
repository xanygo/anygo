//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xlog

import (
	"path/filepath"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xattr"
)

var defaultAccessLogger = &xsync.OnceInit[Logger]{
	New: func() Logger {
		return Default()
	},
}

// AccessLogger 用于打印 server 访问日志的 logger，若没有设置，则返回 Default()
func AccessLogger() Logger {
	return defaultAccessLogger.Load()
}

// SetAccessLogger 设置用于打印 server 访问日志的 logger
func SetAccessLogger(lg Logger) {
	defaultAccessLogger.Store(lg)
}

// AccessLoggerOpt 用于打印 server 访问日志的配置
// 配置内容:
// 1. 首选日志配置文件: conf/log/access
// 2. 若无则使用默认日志文件地址： log/access/access.log
func AccessLoggerOpt() FileLoggerOpt {
	return FileLoggerOpt{
		CfgPath: filepath.Join(xattr.ConfDir(), "log", "access"),
		Cfg: &FileConfig{
			FileName: filepath.Join(xattr.LogDir(), "access", "access.log"),
			ExtRule:  "1hour",
			MaxFiles: 48,
			Dispatch: defaultDispatch,
		},
	}
}
