//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xlog

import (
	"log/slog"
	"strings"
)

func ReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if len(groups) != 0 {
		return a
	}
	switch a.Key {
	case slog.SourceKey:
		if a.Value.Kind() == slog.KindAny {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				arr := strings.Split(source.File, "/")
				index := max(0, len(arr)-3)
				source.File = strings.Join(arr[index:], "/")
			}
		}
	case slog.TimeKey:
		if a.Value.Kind() == slog.KindTime {
			a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05.999"))
		}
	}
	return a
}
