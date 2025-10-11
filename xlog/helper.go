//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xlog

import (
	"log/slog"
	"path/filepath"

	"github.com/xanygo/anygo/ds/xstr"
)

// ReplaceAttr Logger Handler 的用于重写 Attr 函数，目前包含功能：
//  1. source.file 字段的文件路径简化，只保留部分路径
//  2. time 字段格式调整为 2006-01-02 15:04:05.999
func ReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if len(groups) != 0 {
		return a
	}
	switch a.Key {
	case slog.SourceKey:
		if a.Value.Kind() == slog.KindAny {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				source.File = xstr.CutLastByteNAfter(source.File, '/', 3)
				source.Function = filepath.Base(source.Function)
			}
		}
	case slog.TimeKey:
		if a.Value.Kind() == slog.KindTime {
			a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05.999"))
		}
	}
	return a
}

type WithLogger struct {
	lg Logger
}

func (wl *WithLogger) SetLogger(lg Logger) {
	wl.lg = lg
}

func (wl *WithLogger) Logger() Logger {
	return wl.lg
}

func (wl *WithLogger) AutoLogger() Logger {
	if wl.lg == nil {
		return Default()
	}
	return wl.lg
}

func (wl *WithLogger) HasLogger() bool {
	return wl.lg != nil
}
