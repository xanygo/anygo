//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xlog

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"time"
)

// Level 日志等级
type Level = slog.Level

const (
	LevelDebug Level = slog.LevelDebug
	LevelInfo  Level = slog.LevelInfo
	LevelWarn  Level = slog.LevelWarn
	LevelError Level = slog.LevelError
)

var allLevels = []Level{
	LevelDebug,
	LevelInfo,
	LevelWarn,
	LevelError,
}

// Logger 输出日志的 Logger 接口定义
type Logger interface {
	Info(ctx context.Context, msg string, attr ...Attr)
	Debug(ctx context.Context, msg string, attr ...Attr)
	Warn(ctx context.Context, msg string, attr ...Attr)
	Error(ctx context.Context, msg string, attr ...Attr)
	Output(ctx context.Context, level Level, callerSkip int, msg string, attr ...Attr)
}

type NopType interface {
	Nop() bool
}

// IsNop 判断是否是一个空的 Logger
func IsNop(l Logger) bool {
	if l == nil {
		return true
	}
	if nl, ok := l.(NopType); ok && nl.Nop() {
		return true
	}
	return false
}

var _ Logger = (*NopLogger)(nil)

// NopLogger 会丢弃所有日志的 logger
type NopLogger struct{}

func (n NopLogger) Info(context.Context, string, ...Attr) {}

func (n NopLogger) Debug(context.Context, string, ...Attr) {}

func (n NopLogger) Warn(context.Context, string, ...Attr) {}

func (n NopLogger) Error(context.Context, string, ...Attr) {}

func (n NopLogger) Output(context.Context, Level, int, string, ...Attr) {}

func (n NopLogger) Nop() bool {
	return true
}

func NewSimple(w io.Writer) *Simple {
	if dw, ok := w.(DispatchWriter); ok {
		return &Simple{
			Handler: NewDispatchHandler(dw.Writers(), defaultJSONHandler),
			w:       w,
		}
	}
	return &Simple{
		Handler: defaultJSONHandler(w),
		w:       w,
	}
}

func defaultJSONHandler(w io.Writer) Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       LevelDebug,
		AddSource:   true,
		ReplaceAttr: ReplaceAttr,
	})
}

var _ Logger = (*Simple)(nil)

// Simple 一个 Logger 默认实现
type Simple struct {
	w       io.Writer    // 只供 LevelWriter 方法使用
	Handler Handler      // 必填
	Errors  chan<- error // 可选，当输出日志内容出错时,将错误输出到此
}

func (sl *Simple) Info(ctx context.Context, msg string, attr ...Attr) {
	sl.Output(ctx, LevelInfo, 1, msg, attr...)
}

func (sl *Simple) Debug(ctx context.Context, msg string, attr ...Attr) {
	sl.Output(ctx, LevelDebug, 1, msg, attr...)
}

func (sl *Simple) Warn(ctx context.Context, msg string, attr ...Attr) {
	sl.Output(ctx, LevelWarn, 1, msg, attr...)
}

func (sl *Simple) Error(ctx context.Context, msg string, attr ...Attr) {
	sl.Output(ctx, LevelError, 1, msg, attr...)
}

func (sl *Simple) Output(ctx context.Context, level Level, callerSkip int, msg string, attrs ...Attr) {
	err := handlerOutput(ctx, sl.Handler, level, callerSkip+1, msg, attrs...)
	if err != nil && sl.Errors != nil {
		select {
		case sl.Errors <- err:
		default:
		}
	}
}

var _ HasLevelWriter = (*Simple)(nil)

func (sl *Simple) LevelWriter(level Level) io.Writer {
	if hl, ok := sl.Handler.(HasLevelWriter); ok {
		if w := hl.LevelWriter(level); w != nil {
			return w
		}
	}
	return sl.w
}

func handlerOutput(ctx context.Context, handler Handler, level Level, callerSkip int, msg string, attrs ...Attr) error {
	var pcs [1]uintptr
	runtime.Callers(callerSkip+2, pcs[:])
	rec := slog.NewRecord(time.Now(), level, msg, pcs[0])
	meta := MetaAttrsFromCtx(ctx)
	data := AttrsFromCtx(ctx)
	if len(attrs) > 0 {
		data = append(data, attrs...)
	}
	rec.AddAttrs(
		slog.GroupAttrs("meta", meta...),
		slog.GroupAttrs("attr", data...),
	)
	return handler.Handle(ctx, rec)
}

// MultiLogger 将多个 logger 封装为一个，实现一份日志多个输出目标
func MultiLogger(loggers ...Logger) Logger {
	allLoggers := make([]Logger, 0, len(loggers))
	for _, w := range loggers {
		if mw, ok := w.(*multiLogger); ok {
			allLoggers = append(allLoggers, mw.loggers...)
		} else {
			allLoggers = append(allLoggers, w)
		}
	}
	return &multiLogger{
		loggers: allLoggers,
	}
}

var _ Logger = (*multiLogger)(nil)
var _ HasLevelWriter = (*multiLogger)(nil)

type multiLogger struct {
	loggers []Logger
}

func (m *multiLogger) Info(ctx context.Context, msg string, attr ...Attr) {
	m.Output(ctx, LevelInfo, 1, msg, attr...)
}

func (m *multiLogger) Debug(ctx context.Context, msg string, attr ...Attr) {
	m.Output(ctx, LevelDebug, 1, msg, attr...)
}

func (m *multiLogger) Warn(ctx context.Context, msg string, attr ...Attr) {
	m.Output(ctx, LevelWarn, 1, msg, attr...)
}

func (m *multiLogger) Error(ctx context.Context, msg string, attr ...Attr) {
	m.Output(ctx, LevelError, 1, msg, attr...)
}

func (m *multiLogger) Output(ctx context.Context, level Level, callerSkip int, msg string, attr ...Attr) {
	for _, l := range m.loggers {
		l.Output(ctx, level, callerSkip+1, msg, attr...)
	}
}

func (m *multiLogger) LevelWriter(level Level) io.Writer {
	for _, l := range m.loggers {
		if hl, ok := l.(HasLevelWriter); ok {
			if w := hl.LevelWriter(level); w != nil {
				return w
			}
		}
	}
	return nil
}

type HasLevelWriter interface {
	LevelWriter(level Level) io.Writer
}

func AsLevelWriter(l Logger, level Level) io.Writer {
	return AsLevelWriter3(l, level, 1)
}

func AsLevelWriter3(l Logger, level Level, callerSkip int) io.Writer {
	if lw, ok := l.(HasLevelWriter); ok {
		if w := lw.LevelWriter(level); w != nil {
			return w
		}
	}
	return &lgWriter{
		logger:     l,
		level:      level,
		callerSkip: callerSkip,
	}
}

var _ io.Writer = (*lgWriter)(nil)

type lgWriter struct {
	logger     Logger
	level      Level
	callerSkip int
}

func (w *lgWriter) Write(p []byte) (n int, err error) {
	w.logger.Output(context.Background(), w.level, w.callerSkip, string(p))
	return len(p), nil
}
