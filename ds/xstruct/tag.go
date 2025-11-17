//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-07

package xstruct

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/internal/zcache"
)

type Tag struct {
	name   string
	values map[string]string
}

func (t Tag) String() string {
	data := map[string]any{
		"Name":   t.name,
		"Values": t.values,
	}
	bf, _ := json.Marshal(data)
	return string(bf)
}

func (t Tag) Name() string {
	return t.name
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
		values: t.values,
	}
}

func ParserTag(str string) Tag {
	if len(str) == 0 {
		return Tag{}
	}
	arr := strings.Split(str, ",")
	t := Tag{}
	for index, value := range arr {
		if index == 0 {
			t.name = strings.TrimSpace(value)
			continue
		}
		if t.values == nil {
			t.values = make(map[string]string)
		}
		st := strings.SplitN(value, ":", 2)
		switch len(st) {
		case 2:
			t.values[strings.TrimSpace(st[0])] = strings.TrimSpace(st[1])
		case 1:
			t.values[strings.TrimSpace(st[0])] = ""
		default:
			panic(fmt.Sprintf("unexpect: %#v", st))
		}
	}
	return t
}

var structTagCache = &zcache.MapCache[[2]string, Tag]{}

func ParserTagCached(tag reflect.StructTag, name string) Tag {
	key := [2]string{string(tag), name}
	val, ok := structTagCache.Load(key)
	if ok {
		return val
	}
	val = ParserTag(tag.Get(name))
	structTagCache.Set(key, val)
	return val
}
