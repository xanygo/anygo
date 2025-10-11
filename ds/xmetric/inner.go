//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-05

package xmetric

import (
	"context"
	"sync"
	"time"

	"github.com/xanygo/anygo/xctx"
)

var _ Provider = (*InnerProvider)(nil)

type InnerProvider struct {
	OnEnd func(ctx context.Context, name string, span Span)
}

func (p *InnerProvider) Tracer(name string) Tracer {
	return &tracer{
		OnEnd: p.OnEnd,
	}
}

var _ Tracer = (*tracer)(nil)

type tracer struct {
	OnEnd func(ctx context.Context, name string, span Span)
}

var keySpan = xctx.NewKey()

func (t *tracer) Start(ctx context.Context, name string) (context.Context, Span) {
	parent := SpanFromContext(ctx)
	if !IsNoop(parent) {
		return parent.NewChild(ctx, name)
	}
	s := newSpan(ctx, name, t.OnEnd)
	return context.WithValue(ctx, keySpan, s), s
}

func newSpan(ctx context.Context, name string, end onEndFunc) *span {
	return &span{
		name:  name,
		id:    NewSpanID(),
		start: time.Now(),
		ctx:   ctx,
		onEnd: end,
	}
}

type onEndFunc func(ctx context.Context, name string, span Span)

type span struct {
	id           SpanID
	ctx          context.Context
	name         string
	start        time.Time
	end          time.Time
	onEnd        onEndFunc
	err          error
	attrs        []Attribute
	parent       Span
	children     []SpanReader
	attemptCount int
	mux          sync.Mutex
}

func (s *span) Name() string {
	return s.name
}

func (s *span) StartTime() time.Time {
	return s.start
}

func (s *span) EndTime() time.Time {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.end
}

func (s *span) AttemptCount() int {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.attemptCount
}

func (s *span) SetAttemptCount(i int) {
	s.mux.Lock()
	s.attemptCount = i
	s.mux.Unlock()
}

func (s *span) Parent() SpanReader {
	return s.parent
}

func (s *span) ID() SpanID {
	return s.id
}

func (s *span) End() {
	s.mux.Lock()
	s.end = time.Now()
	s.mux.Unlock()
}

func (s *span) SetAttributes(kv ...Attribute) {
	s.mux.Lock()
	s.attrs = append(s.attrs, kv...)
	s.mux.Unlock()
}

func (s *span) Attributes() []Attribute {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.attrs
}

func (s *span) RecordError(err error) {
	s.mux.Lock()
	s.err = err
	s.mux.Unlock()
}

func (s *span) Error() error {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.err
}

func (s *span) IsRecording() bool {
	if s.parent != nil && !s.parent.IsRecording() {
		return false
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.end.IsZero()
}

func (s *span) isRecordingLocked() bool {
	if s.parent != nil && !s.parent.IsRecording() {
		return false
	}
	return s.end.IsZero()
}

func (s *span) NewChild(ctx context.Context, name string) (context.Context, Span) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if !s.isRecordingLocked() {
		return ctx, emptySpan
	}
	child := newSpan(ctx, name, s.onEnd)
	s.children = append(s.children, child)
	ctx = context.WithValue(ctx, keySpan, child)
	return ctx, child
}

func (s *span) Children() []SpanReader {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.children
}
