//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xlog

import (
	"context"
	"os"
	"path/filepath"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xattr"
)

var defaultLogger = &xsync.OnceInit[Logger]{
	New: func() Logger {
		return NewSimpleWithLevel(os.Stderr, DefaultLevel)
	},
}

// Default 获取默认的 logger，默认日志内容输出到 os.Stderr,并且为 JSON 格式
func Default() Logger {
	return defaultLogger.Load()
}

// SetDefault 替换 Default Logger
func SetDefault(l Logger) {
	defaultLogger.Store(l)
}

// Info 使用 Default() Logger 打印 Info 日志
func Info(ctx context.Context, msg string, attr ...Attr) {
	Default().Output(ctx, LevelInfo, 1, msg, attr...)
}

// Debug 使用 Default() Logger 打印 Debug 日志
func Debug(ctx context.Context, msg string, attr ...Attr) {
	Default().Output(ctx, LevelDebug, 1, msg, attr...)
}

// Warn 使用 Default() Logger 打印 Warn 日志
func Warn(ctx context.Context, msg string, attr ...Attr) {
	Default().Output(ctx, LevelWarn, 1, msg, attr...)
}

// Error 使用 Default() Logger 打印 Error 日志
func Error(ctx context.Context, msg string, attr ...Attr) {
	Default().Output(ctx, LevelError, 1, msg, attr...)
}

// DefaultLoggerOpt 默认 logger 的配置，用于打印服务端内部组件运行状态的
// 配置内容:
// 1. 首选日志配置文件: conf/log/default
// 2. 若无则使用默认日志文件地址： log/default/default.log
func DefaultLoggerOpt() FileLoggerOpt {
	return FileLoggerOpt{
		CfgPath: filepath.Join(xattr.ConfDir(), "log", "default"),
		Cfg: &FileConfig{
			FileName: filepath.Join(xattr.LogDir(), "default", "default.log"),
			ExtRule:  "1hour",
			MaxFiles: 48,
			Dispatch: defaultDispatch,
		},
	}
}

// InitAllDefaultLogger 使用所有的 XXXLoggerOpt 初始化并赋值给对应的默认 Logger
func InitAllDefaultLogger() {
	stdLogger := NewSimpleWithLevel(os.Stderr, DefaultLevel)
	SetDefault(MultiLogger(stdLogger, DefaultLoggerOpt().MustNewLogger()))
	SetPanicLogger(MultiLogger(stdLogger, PanicLoggerOpt().MustNewLogger()))
	SetAccessLogger(AccessLoggerOpt().MustNewLogger())
	SetClientLogger(ClientLoggerOpt().MustNewLogger())
}

// SetAllDefaultLogger 设置所有默认内置的Logger 为同一个
func SetAllDefaultLogger(logger Logger) {
	SetDefault(logger)
	SetPanicLogger(logger)
	SetAccessLogger(logger)
	SetClientLogger(logger)
}
