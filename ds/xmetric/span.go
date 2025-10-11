//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-04

package xmetric

import (
	"context"
	"crypto/rand"
	"time"
)

func SpanFromContext(ctx context.Context) Span {
	s, ok := ctx.Value(keySpan).(Span)
	if ok {
		return s
	}
	return emptySpan
}

type SpanID [8]byte

func (s SpanID) IsValid() bool {
	return s != SpanID{}
}

type Span interface {
	SpanReader
	SpanWriter
}

type SpanWriter interface {
	End()
	SetAttributes(kv ...Attribute)
	SetAttemptCount(total int) // 记录重试总次数
	RecordError(err error)
	NewChild(ctx context.Context, name string) (context.Context, Span)
}

type SpanReader interface {
	ID() SpanID
	Name() string
	Parent() SpanReader
	StartTime() time.Time
	EndTime() time.Time
	Attributes() []Attribute
	AttemptCount() int
	Error() error
	IsRecording() bool
	Children() []SpanReader
}

type Attribute struct {
	K string
	V any
}

func NewSpanID() SpanID {
	var sid SpanID
	for {
		_, _ = rand.Read(sid[:])
		if sid.IsValid() {
			return sid
		}
	}
}

var _ Span = (*span)(nil)

func AnyAttr(key string, val any) Attribute {
	return Attribute{K: key, V: val}
}
