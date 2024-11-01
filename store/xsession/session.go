//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xsession

import (
	"context"
	"encoding/json"
	"time"

	"github.com/xanygo/anygo/xctx"
	"github.com/xanygo/anygo/xmap"
	"github.com/xanygo/anygo/xsync"
)

type Session struct {
	id      string
	created xsync.Value[time.Time]
	updated xsync.Value[time.Time]
	values  xmap.Sync[string, any]
	storage Storage
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Set(key string, value any) {
	s.updated.Store(time.Now())
	s.values.Store(key, value)
}

func (s *Session) Get(key string) any {
	v, _ := s.values.Load(key)
	return v
}
func (s *Session) Range(fn func(key string, value any) bool) {
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
	Values  map[string]any
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
	Created int64          `json:"c"`
	Updated int64          `json:"u"`
	Values  map[string]any `json:"v"`
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
	Get(ctx context.Context, id string) (*Session, error)
	GetOrCreate(ctx context.Context, id string) *Session
	Save(ctx context.Context, session *Session) error
}

var ctxKeyStore = xctx.NewKey()

func WithStorage(ctx context.Context, store Storage) context.Context {
	return context.WithValue(ctx, ctxKeyStore, store)
}

func StorageFromContext(ctx context.Context) Storage {
	return ctx.Value(ctxKeyStore).(Storage)
}

func FromContext(ctx context.Context) *Session {
	return StorageFromContext(ctx).GetOrCreate(ctx, IDFromContext(ctx))
}
