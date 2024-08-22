//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xlog

import (
	"context"
	"os"

	"github.com/xanygo/anygo/xsync"
)

var defaultLogger xsync.Value[Logger]

// Default 获取默认的 logger，默认日志内容输出到 os.Stderr,并且为 JSON 格式
func Default() Logger {
	return defaultLogger.Load()
}

// SetDefault 替换 Default Logger
func SetDefault(l Logger) {
	defaultLogger.Store(l)
}

func init() {
	SetDefault(NewSimple(os.Stderr))
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