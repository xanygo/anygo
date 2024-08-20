//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xlog

import (
	"context"
	"log/slog"
)

type Level = slog.Level

var (
	LevelDebug Level = slog.LevelDebug
	LevelInfo  Level = slog.LevelInfo
	LevelWarn  Level = slog.LevelWarn
	LevelError Level = slog.LevelError
)

type Attr = slog.Attr

type Logger interface {
	Info(ctx context.Context, msg string, attr ...Attr)
	Debug(ctx context.Context, msg string, attr ...Attr)
	Warn(ctx context.Context, msg string, attr ...Attr)
	Error(ctx context.Context, msg string, attr ...Attr)
	Output(ctx context.Context, level Level, callerSkip int, msg string, attr ...Attr)
}

var _ Logger = (*NopLogger)(nil)

type NopLogger struct{}

func (n NopLogger) Info(ctx context.Context, msg string, attr ...Attr) {}

func (n NopLogger) Debug(ctx context.Context, msg string, attr ...Attr) {}

func (n NopLogger) Warn(ctx context.Context, msg string, attr ...Attr) {}

func (n NopLogger) Error(ctx context.Context, msg string, attr ...Attr) {}

func (n NopLogger) Output(ctx context.Context, level Level, callerSkip int, msg string, attr ...Attr) {
}
