//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-06

package xoption

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/xanygo/anygo/xbus"
	"github.com/xanygo/anygo/xctx"
)

var Topic = xbus.NewTopic("option")

type Option interface {
	Reader
	Writer
}

// Reader 只读的配置组件接口定义
type Reader interface {
	Get(key Key) (any, bool)
	Range(fn func(key Key, val any) bool)
}

// Writer 只写的配置组件接口定义
type Writer interface {
	Set(key Key, val any)
	Delete(keys ...Key)
}

// Key 配置的 key 类型
type Key struct {
	id   int64
	name string
	str  string
}

func (k Key) String() string {
	return k.name
}

func (k Key) Name() string {
	return k.name
}

var globalKeyID atomic.Int64

// NewKey 创建一个新的全局唯一的 Key
func NewKey(name string) Key {
	id := globalKeyID.Add(1)
	return Key{
		id:   id,
		name: name,
		str:  fmt.Sprintf("%d-%s", id, name),
	}
}

var _ Option = (*Dynamic)(nil)

func NewDynamic() *Dynamic {
	return &Dynamic{
		values: &sync.Map{},
	}
}

// Dynamic 并发安全的，可动态、并发读写的 配置存储管理器
type Dynamic struct {
	values *sync.Map
}

func (d *Dynamic) Get(key Key) (any, bool) {
	return d.values.Load(key)
}

func (d *Dynamic) Set(key Key, val any) {
	d.values.Store(key, val)
}

func (d *Dynamic) Delete(keys ...Key) {
	for _, key := range keys {
		d.values.Delete(key)
	}
}

func (d *Dynamic) Clone() *Dynamic {
	data := &sync.Map{}
	d.values.Range(func(key, value any) bool {
		data.Store(key, value)
		return true
	})
	return &Dynamic{
		values: data,
	}
}

func (d *Dynamic) Range(fn func(key Key, val any) bool) {
	d.values.Range(func(key, value any) bool {
		return fn(key.(Key), value)
	})
}

func Readers(rds ...Reader) Reader {
	return readers(rds)
}

type readers []Reader

func (rs readers) Get(key Key) (any, bool) {
	for _, reader := range rs {
		val, ok := Get(reader, key)
		if ok {
			return val, ok
		}
	}
	return nil, false
}

func (rs readers) Range(fn func(key Key, val any) bool) {
	var goOn bool
	for _, rd := range rs {
		rd.Range(func(key Key, val any) bool {
			goOn = fn(key, val)
			return goOn
		})
		if !goOn {
			return
		}
	}
}

type EmptyReader bool

func (e EmptyReader) Get(key Key) (any, bool) {
	return nil, false
}

func (e EmptyReader) Range(fn func(key Key, val any) bool) {}

func NewSimple() *Simple {
	return &Simple{
		value: make(map[Key]any, 4),
	}
}

var _ Option = (*Simple)(nil)

// Simple 一个简单的，非并发安全的配置存储管理器
type Simple struct {
	value map[Key]any
}

func (m *Simple) Get(key Key) (any, bool) {
	if m == nil || len(m.value) == 0 {
		return nil, false
	}
	v, ok := m.value[key]
	return v, ok
}

func (m *Simple) Range(fn func(key Key, val any) bool) {
	for k, v := range m.value {
		if !fn(k, v) {
			return
		}
	}
}

func (m *Simple) Set(key Key, val any) {
	m.value[key] = val
}

func (m *Simple) Delete(keys ...Key) {
	for _, key := range keys {
		delete(m.value, key)
	}
}

func (m *Simple) Empty() bool {
	return len(m.value) == 0
}

func (m *Simple) Value() *Simple {
	if m.Empty() {
		return nil
	}
	return m
}

var ctxReaderKey = xctx.NewKey()

func ContextWithReader(ctx context.Context, option Reader) context.Context {
	return context.WithValue(ctx, ctxReaderKey, option)
}

func ReaderFromContext(ctx context.Context) Reader {
	obj := ctx.Value(ctxReaderKey)
	if opt, ok := obj.(Reader); ok {
		return opt
	}
	return EmptyReader(false)
}
