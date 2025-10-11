//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xlog

import (
	"errors"
	"log/slog"
	"time"

	"github.com/xanygo/anygo/xerror"
)

type Attr = slog.Attr

func Any(key string, value any) Attr {
	return slog.Any(key, value)
}

func Bool(key string, v bool) Attr {
	return slog.Bool(key, v)
}

func Time(key string, v time.Time) Attr {
	return slog.Time(key, v)
}

func Duration(key string, v time.Duration) Attr {
	return slog.Duration(key, v)
}

func DurationMS(key string, v time.Duration) Attr {
	return slog.Float64(key, float64(v.Nanoseconds())/1e6)
}

func Float64(key string, v float64) Attr {
	return slog.Float64(key, v)
}

func Float32(key string, v float32) Attr {
	return slog.Any(key, v)
}

func Group(key string, args ...any) Attr {
	return slog.Group(key, args...)
}

func GroupAttrs(key string, args ...Attr) Attr {
	return slog.Attr{Key: key, Value: slog.GroupValue(args...)}
}

func Int(key string, value int) Attr {
	return slog.Int(key, value)
}

func Int8(key string, value int8) Attr {
	return slog.Any(key, value)
}

func Int16(key string, value int16) Attr {
	return slog.Any(key, value)
}

func Int32(key string, value int32) Attr {
	return slog.Any(key, value)
}

func Int64(key string, value int64) Attr {
	return slog.Int64(key, value)
}

func Uint(key string, value uint) Attr {
	return slog.Any(key, value)
}

func Uint8(key string, value uint8) Attr {
	return slog.Any(key, value)
}

func Uint16(key string, value uint16) Attr {
	return slog.Any(key, value)
}

func Uint32(key string, value uint32) Attr {
	return slog.Any(key, value)
}

func Uint64(key string, v uint64) Attr {
	return slog.Uint64(key, v)
}

func String(key, value string) Attr {
	return slog.String(key, value)
}

func Bytes(key string, value []byte) Attr {
	return slog.Any(key, value)
}

func ErrorAttr(key string, err error) Attr {
	if err == nil {
		return String(key, "")
	}
	var pt xerror.TraceError
	if errors.As(err, &pt) {
		return Any(key, pt.TraceData())
	}
	return String(key, err.Error())
}
