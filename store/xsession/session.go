//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xsession

import (
	"context"
	"reflect"
	"time"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/xcodec"
)

type Session interface {
	ID() string
	Set(ctx context.Context, key string, value string) error
	MSet(ctx context.Context, kv map[string]string) error
	Get(ctx context.Context, key string) (string, error)
	MGet(ctx context.Context, keys ...string) (map[string]string, error)
	Delete(ctx context.Context, keys ...string) error
	Created(ctx context.Context) (time.Time, error)
	Save(ctx context.Context) error
	Clear(ctx context.Context) error
}

type Storage interface {
	// Get 从存储中加载 Session 数据，若不存在会报错
	Get(ctx context.Context, id string) Session
}

var (
	ctxKeyStore   = xctx.NewKey()
	ctxKeySession = xctx.NewKey()
)

func WithStorage(ctx context.Context, store Storage) context.Context {
	return context.WithValue(ctx, ctxKeyStore, store)
}

func StorageFromContext(ctx context.Context) Storage {
	return ctx.Value(ctxKeyStore).(Storage)
}

func WithSession(ctx context.Context, store Session) context.Context {
	return context.WithValue(ctx, ctxKeySession, store)
}

func FromContext(ctx context.Context) Session {
	ss, _ := ctx.Value(ctxKeySession).(Session)
	return ss
}

// Set 将数据 val 使用 json 编码，并调用 Session.Set 保存
// 注意：使用此方法写入的数据，必须使用 Load 或 Get 等来读取，不可以直接使用 Session 对象的 Load、Get 等方法
func Set[T any](ctx context.Context, s Session, key string, val T) error {
	str, err := xcodec.EncodeToString(xcodec.JSON, val)
	if err != nil {
		return err
	}
	return s.Set(ctx, key, str)
}

func GetAndDelete[T any](ctx context.Context, s Session, key string) (result T, err error) {
	val, err := Get[T](ctx, s, key)
	if err != nil {
		return result, err
	}
	_ = s.Delete(ctx, key)
	return val, nil
}

func Get[T any](ctx context.Context, s Session, key string) (result T, err error) {
	str, err := s.Get(ctx, key)
	if err != nil {
		return result, err
	}
	err = xcodec.DecodeFromString(xcodec.JSON, str, &result)
	return result, err
}

func EqualAndDelete[T any](ctx context.Context, s Session, key string, value T) bool {
	result, err := GetAndDelete[T](ctx, s, key)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(result, value)
}
