//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xlog

import (
	"context"
	"errors"
	"io"
	"log/slog"
)

type DispatchWriter interface {
	Writers() map[Level]io.WriteCloser
}

// Handler 日志记录行的处理对象，调用 Handle 方法后会触发日志落盘
type Handler interface {
	Enabled(context.Context, Level) bool
	Handle(context.Context, slog.Record) error
}

type NewHandlerFunc func(w io.Writer) Handler

func NewDispatchHandler(ws map[Level]io.WriteCloser, nhd NewHandlerFunc) Handler {
	hd := &dispatchHandler{
		writers:      ws,
		levelHandler: make(map[Level]Handler, len(ws)),
	}
	for l, w := range ws {
		hd.levelHandler[l] = nhd(w)
	}
	return hd
}

var _ io.Closer = (*dispatchHandler)(nil)
var _ Handler = (*dispatchHandler)(nil)
var _ HasLevelWriter = (*dispatchHandler)(nil)

type dispatchHandler struct {
	writers      map[Level]io.WriteCloser
	levelHandler map[Level]Handler
}

func (h *dispatchHandler) Enabled(ctx context.Context, level Level) bool {
	_, ok := h.levelHandler[level]
	return ok
}

func (h *dispatchHandler) Handle(ctx context.Context, record slog.Record) error {
	hd, ok := h.levelHandler[record.Level]
	if !ok {
		return nil
	}
	return hd.Handle(ctx, record)
}

func (h *dispatchHandler) LevelWriter(level Level) io.Writer {
	return h.writers[level]
}

func (h *dispatchHandler) Close() error {
	var errs []error
	for _, w := range h.writers {
		if err := w.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
