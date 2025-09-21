//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xsession

import (
	"context"
	"encoding/json"
	"reflect"
	"time"
	"unsafe"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xctx"
	"github.com/xanygo/anygo/xerror"
)

type Session struct {
	id      string
	created xsync.Value[time.Time]
	updated xsync.Value[time.Time]
	values  xmap.Sync[string, string]
	storage Storage
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Set(key string, value string) {
	s.updated.Store(time.Now())
	s.values.Store(key, value)
}

func (s *Session) Get(key string) string {
	v, _ := s.values.Load(key)
	return v
}

func (s *Session) Load(key string) (string, bool) {
	return s.values.Load(key)
}

func (s *Session) Equal(key string, value string) bool {
	v, ok := s.values.Load(key)
	if !ok {
		return false
	}
	return value == v
}

func (s *Session) LoadAndDelete(key string) (string, bool) {
	return s.values.LoadAndDelete(key)
}

func (s *Session) Has(key string) bool {
	_, ok := s.values.Load(key)
	return ok
}

func (s *Session) CompareAndDelete(key string, value string) bool {
	deleted := s.values.CompareAndDelete(key, value)
	if deleted {
		s.updated.Store(time.Now())
	}
	return deleted
}

func (s *Session) Range(fn func(key string, value string) bool) {
	s.values.Range(fn)
}

func (s *Session) Delete(keys ...string) {
	s.updated.Store(time.Now())
	if len(keys) == 0 {
		return
	}
	for _, key := range keys {
		s.values.Delete(key)
	}
}

func (s *Session) Save(ctx context.Context) error {
	return s.storage.Save(ctx, s)
}

func (s *Session) Clear() {
	s.values.Clear()
}

func (s *Session) Created() time.Time {
	return s.created.Load()
}

func (s *Session) Updated() time.Time {
	return s.updated.Load()
}

func (s *Session) Bytes() ([]byte, error) {
	data := map[string]any{
		"c": s.created.Load().Unix(),
		"v": s.values.ToMap(),
		"u": s.updated.Load().Unix(),
	}
	return json.Marshal(data)
}

func NewValue(id string) *Value {
	now := time.Now()
	val := &Value{
		ID:      id,
		Created: now,
		Updated: now,
	}
	return val
}

type Value struct {
	ID      string
	Created time.Time
	Updated time.Time
	Values  map[string]string
}

func (v *Value) ToSession(s Storage) *Session {
	se := &Session{
		id:      v.ID,
		storage: s,
	}
	se.created.Store(v.Created)
	se.updated.Store(v.Updated)
	for k, v := range v.Values {
		se.values.Store(k, v)
	}
	return se
}

type valueData struct {
	Created int64             `json:"c"`
	Updated int64             `json:"u"`
	Values  map[string]string `json:"v"`
}

func ParserValue(bf []byte) (*Value, error) {
	var v *valueData
	if err := json.Unmarshal(bf, &v); err != nil {
		return nil, err
	}
	val := &Value{
		Values:  v.Values,
		Created: time.Unix(v.Created, 0),
		Updated: time.Unix(v.Updated, 0),
	}
	return val, nil
}

type Storage interface {
	// Get 从存储中加载 Session 数据，若不存在会报错
	Get(ctx context.Context, id string) (*Session, error)

	// GetOrCreate 从存储中加载 Session 数据，若不存在则生成一个新的
	GetOrCreate(ctx context.Context, id string) *Session

	// Save 保存数据
	Save(ctx context.Context, session *Session) error
}

var (
	ctxKeyStore   = xctx.NewKey()
	ctxKeySession = xctx.NewKey()
)

func WithStorage(ctx context.Context, store Storage) context.Context {
	return context.WithValue(ctx, ctxKeyStore, store)
}

func WithSession(ctx context.Context, store *Session) context.Context {
	return context.WithValue(ctx, ctxKeySession, store)
}

func StorageFromContext(ctx context.Context) Storage {
	return ctx.Value(ctxKeyStore).(Storage)
}

func FromContext(ctx context.Context) *Session {
	ss, _ := ctx.Value(ctxKeySession).(*Session)
	return ss
}

// Set 将数据 val 使用 json 编码，并调用 Session.Set 保存
// 注意：使用此方法写入的数据，必须使用 Load 或 Get 等来读取，不可以直接使用 Session 对象的 Load、Get 等方法
func Set[T any](s *Session, key string, val T) error {
	bf, err := json.Marshal(val)
	if err != nil {
		return err
	}
	str := unsafe.String(&bf[0], len(bf))
	s.Set(key, str)
	return nil
}

func Load[T any](s *Session, key string) (result T, err error) {
	str, ok := s.Load(key)
	if !ok {
		return result, xerror.NotFound
	}
	err = json.Unmarshal([]byte(str), &result)
	return result, err
}

func Get[T any](s *Session, key string) (result T) {
	str, ok := s.Load(key)
	if !ok {
		return result
	}
	_ = json.Unmarshal([]byte(str), &result)
	return result
}

func LoadAndDelete[T any](s *Session, key string) (result T, err error) {
	str, ok := s.LoadAndDelete(key)
	if !ok {
		return result, xerror.NotFound
	}
	err = json.Unmarshal([]byte(str), &result)
	return result, err
}

func Equal[T any](s *Session, key string, value T) bool {
	val, err := Load[T](s, key)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(val, value)
}

func EqualAndDelete[T any](s *Session, key string, value T) bool {
	str, ok := s.LoadAndDelete(key)
	if !ok {
		return false
	}
	var result T
	err := json.Unmarshal([]byte(str), &result)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(result, value)
}
