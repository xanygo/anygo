//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-04

package xmetric

import (
	"context"
	"time"
)

type Noop interface {
	IsNoop()
}

func IsNoop(obj any) bool {
	if obj == nil {
		return true
	}
	_, ok := obj.(Noop)
	return ok
}

var _ Provider = (*NoopProvider)(nil)
var _ Noop = (*NoopProvider)(nil)

type NoopProvider struct {
}

func (n NoopProvider) Tracer(name string) Tracer {
	return NoopTracer{}
}

func (n NoopProvider) IsNoop() {}

var _ Tracer = (*NoopTracer)(nil)
var _ Noop = (*NoopTracer)(nil)

type NoopTracer struct{}

func (n NoopTracer) Start(ctx context.Context, name string) (context.Context, Span) {
	return ctx, emptySpan
}

func (n NoopTracer) IsNoop() {}

var emptySpan = NoopSpan(false)

var _ Span = NoopSpan(false)
var _ Noop = NoopSpan(false)

type NoopSpan bool

func (e NoopSpan) Name() string {
	return ""
}

func (e NoopSpan) StartTime() time.Time {
	return time.Time{}
}

func (e NoopSpan) EndTime() time.Time {
	return time.Time{}
}

func (e NoopSpan) IsNoop() {}

func (e NoopSpan) AttemptCount() int {
	return 0
}

func (e NoopSpan) SetAttemptCount(i int) {}

func (e NoopSpan) ID() SpanID {
	return SpanID{}
}

func (e NoopSpan) End() {}

func (e NoopSpan) SetAttributes(kv ...Attribute) {}

func (e NoopSpan) RecordError(err error) {}

func (e NoopSpan) IsRecording() bool {
	return false
}

func (e NoopSpan) Attributes() []Attribute {
	return nil
}

func (e NoopSpan) Error() error {
	return nil
}

func (e NoopSpan) NewChild(ctx context.Context, name string) (context.Context, Span) {
	return ctx, emptySpan
}

func (e NoopSpan) Children() []SpanReader {
	return nil
}

func (e NoopSpan) Parent() SpanReader {
	return emptySpan
}
