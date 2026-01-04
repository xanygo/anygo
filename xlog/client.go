//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xlog

import (
	"path/filepath"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xattr"
)

var defaultClientLogger = &xsync.OnceInit[Logger]{
	New: func() Logger {
		return Default()
	},
}

// ClientLogger 用于打印 client 请求/响应结果的 logger
func ClientLogger() Logger {
	return defaultClientLogger.Load()
}

// SetClientLogger 设置打印 client 请求/响应结果的 logger
func SetClientLogger(lg Logger) {
	defaultClientLogger.Store(lg)
}

// ClientLoggerOpt 打印 client 请求/响应结果的 logger 的配置
// 配置内容:
// 1. 首选日志配置文件: conf/log/client
// 2. 若无则使用默认日志文件地址： log/client/client.log
func ClientLoggerOpt() FileLoggerOpt {
	return FileLoggerOpt{
		CfgPath: filepath.Join(xattr.ConfDir(), "log", "client"),
		Cfg: &FileConfig{
			FileName: filepath.Join(xattr.LogDir(), "client", "client.log"),
			ExtRule:  "1hour",
			MaxFiles: 48,
		},
	}
}
