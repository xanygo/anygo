//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-07

package xstruct

import (
	"encoding/json"
	"iter"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/xanygo/anygo/internal/zcache"
	"github.com/xanygo/anygo/internal/zstr"
)

type Tag struct {
	name   string // tag 的名字
	keys   []string
	values map[string]string
}

func (t Tag) Name() string {
	return t.name
}

func (t Tag) String() string {
	data := map[string]any{
		"Name":   t.name,
		"Keys":   t.keys,
		"Values": t.values,
	}
	bf, _ := json.Marshal(data)
	return string(bf)
}

func (t Tag) Empty() bool {
	return len(t.values) == 0
}

func (t Tag) Keys() []string {
	return t.keys
}

func (t Tag) Iter() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, k := range t.keys {
			if !yield(k, t.values[k]) {
				return
			}
		}
	}
}

func (t Tag) Values() map[string]string {
	return t.values
}

func (t Tag) Get(key string) (string, bool) {
	if t.values == nil {
		return "", false
	}
	v, ok := t.values[key]
	return v, ok
}

func (t Tag) Value(key string) string {
	if t.values == nil {
		return ""
	}
	return t.values[key]
}

func (t Tag) Has(key string) bool {
	if t.values == nil {
		return false
	}
	_, has := t.values[key]
	return has
}

func (t Tag) WithName(name string) Tag {
	return Tag{
		name:   name,
		values: maps.Clone(t.values),
		keys:   slices.Clone(t.keys),
	}
}

// First 读取首个 key，value
func (t Tag) First() (string, string) {
	if len(t.keys) > 0 {
		key := t.keys[0]
		return key, t.values[key]
	}
	return "", ""
}

// Shift 将首个key 作为 name，并返回新 tag
//
// 对于 `db:"id,codec=csv"` 这种，使用后，Name 会读取到 id
func (t Tag) Shift() Tag {
	name, _ := t.First()
	keys := slices.Clone(t.keys)
	if len(keys) > 1 {
		keys = keys[1:]
	}
	nt := Tag{
		name:   name,
		keys:   keys,
		values: maps.Clone(t.values),
	}
	delete(nt.values, name)
	return nt
}

// ParserTag 解析 struct 的 tag 内容
//
// 支持的格式参考 zstr.ScanPairs
func ParserTag(name string, str string) Tag {
	if len(str) == 0 {
		return Tag{
			name: name,
		}
	}
	n := strings.Count(str, ",")
	t := Tag{
		name:   name,
		keys:   make([]string, 0, n),
		values: make(map[string]string, n),
	}
	zstr.ScanPairs(str, ',', func(key, value string) bool {
		key = strings.TrimSpace(key)
		if key == "" {
			return true
		}
		t.keys = append(t.keys, key)
		t.values[key] = strings.TrimSpace(value)
		return true
	})
	return t
}

var structTagCache = &zcache.MapCache[[2]string, Tag]{}

// CachedTag 解析 `xv:"required"` 这种 tag
func CachedTag(tag reflect.StructTag, name string) Tag {
	key := [2]string{string(tag), name}
	val, ok := structTagCache.Load(key)
	if ok {
		return val
	}
	val = ParserTag(name, tag.Get(name))
	structTagCache.Set(key, val)
	return val
}

var structDBTagCache = &zcache.MapCache[[2]string, Tag]{}

// CachedShiftTag 解析 `db:"id,codec=csv"` 这种 tag
func CachedShiftTag(tag reflect.StructTag, name string) Tag {
	key := [2]string{string(tag), name}
	val, ok := structDBTagCache.Load(key)
	if ok {
		return val
	}
	val = ParserTag(name, tag.Get(name)).Shift()
	structDBTagCache.Set(key, val)
	return val
}
