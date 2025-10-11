//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-23

package xkv

import (
	"context"
	"slices"
	"sort"
	"strconv"
	"sync"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xkv/internal"
	"github.com/xanygo/anygo/xcodec"
)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// NewMemoryStoreAny 创建一个值类型支持泛型类型的，全内存存储的 KV 存储对象
func NewMemoryStoreAny[V any](coder xcodec.Codec) *Transformer[V] {
	return &Transformer[V]{
		Codec:   coder,
		Storage: NewMemoryStore(),
	}
}

var _ Storage[string] = (*MemoryStore)(nil)

// MemoryStore 底层基础类型为 string 的内存存储实现
type MemoryStore struct {
	values   map[string]string
	keyTypes map[string]internal.DataType
	mux      sync.RWMutex
}

func (m *MemoryStore) tryInit() {
	if m.values == nil {
		m.values = make(map[string]string)
		m.keyTypes = make(map[string]internal.DataType)
	}
}

func (m *MemoryStore) getLocked(key string, wantType internal.DataType) (string, bool, error) {
	var value string
	var found bool
	var tp internal.DataType
	m.mux.RLock()
	if len(m.values) > 0 {
		value, found = m.values[key]
		tp = m.keyTypes[key]
	}
	m.mux.RUnlock()
	if found && tp != wantType {
		return "", false, ErrInvalidType
	}
	return value, found, nil
}

func (m *MemoryStore) setLocked(key string, value string, tp internal.DataType) {
	m.mux.Lock()
	m.tryInit()
	m.values[key] = value
	m.keyTypes[key] = tp
	m.mux.Unlock()
}

func (m *MemoryStore) withLock(fn func(db map[string]string, tps map[string]internal.DataType)) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.tryInit()
	fn(m.values, m.keyTypes)
}

func (m *MemoryStore) String(key string) String[string] {
	return &memString{
		store: m,
		key:   key,
	}
}

var _ String[string] = (*memString)(nil)

type memString struct {
	store *MemoryStore
	key   string
}

func (m *memString) Set(ctx context.Context, value string) error {
	m.store.setLocked(m.key, value, internal.DataTypeString)
	return nil
}

func (m *memString) Get(ctx context.Context) (string, bool, error) {
	return m.store.getLocked(m.key, internal.DataTypeString)
}

func (m *memString) Incr(ctx context.Context) (result int64, err error) {
	m.store.withLock(func(db map[string]string, tps map[string]internal.DataType) {
		val, ok := db[m.key]
		if !ok {
			db[m.key] = "1"
			tps[m.key] = internal.DataTypeString
			result = 1
			return
		}
		tp := tps[m.key]
		if tp != internal.DataTypeString {
			val = "0"
		}
		result, _ = strconv.ParseInt(val, 10, 64)
		result++
		db[m.key] = strconv.FormatInt(result, 10)
		tps[m.key] = internal.DataTypeString
	})
	return result, err
}

func (m *memString) Decr(ctx context.Context) (result int64, err error) {
	m.store.withLock(func(db map[string]string, tps map[string]internal.DataType) {
		val, ok := db[m.key]
		if !ok {
			db[m.key] = "-1"
			tps[m.key] = internal.DataTypeString
			result = -1
			return
		}
		tp := tps[m.key]
		if tp != internal.DataTypeString {
			val = "0"
		}
		result, _ = strconv.ParseInt(val, 10, 64)
		result--
		db[m.key] = strconv.FormatInt(result, 10)
		tps[m.key] = internal.DataTypeString
	})
	return result, err
}

func (m *MemoryStore) List(key string) List[string] {
	return &memList{
		store: m,
		key:   key,
	}
}

var _ List[string] = (*memList)(nil)

type memList struct {
	store *MemoryStore
	key   string
}

func (m *memList) getSlice() ([]string, error) {
	val, found, err := m.store.getLocked(m.key, internal.DataTypeList)
	if err != nil || !found {
		return nil, err
	}
	var result []string
	err = xcodec.JSON.Decode([]byte(val), &result)
	return result, err
}

func (m *memList) withSliceLocked(fn func([]string) ([]string, bool)) error {
	var err error
	m.store.withLock(func(db map[string]string, tps map[string]internal.DataType) {
		str, found := db[m.key]
		tp := tps[m.key]
		if found && tp != internal.DataTypeList {
			err = ErrInvalidType
			return
		}
		var result []string
		if found {
			xcodec.JSON.Decode([]byte(str), &result)
		}
		newValue, replace := fn(result)
		if replace {
			bf, _ := xcodec.JSON.Encode(newValue)
			db[m.key] = string(bf)
			tps[m.key] = internal.DataTypeList
		}
	})
	return err
}

func (m *memList) LPush(ctx context.Context, values ...string) (int, error) {
	err := m.withSliceLocked(func(list []string) ([]string, bool) {
		return slices.Insert(list, 0, values...), true
	})
	return len(values), err
}

func (m *memList) RPush(ctx context.Context, values ...string) (int, error) {
	err := m.withSliceLocked(func(list []string) ([]string, bool) {
		return append(list, values...), true
	})
	return len(values), err
}

func (m *memList) LPop(ctx context.Context) (string, bool, error) {
	var value string
	var found bool
	err := m.withSliceLocked(func(list []string) ([]string, bool) {
		if len(list) == 0 {
			return nil, false
		}
		list, value, found = xslice.PopHead(list)
		return list, true
	})

	return value, found, err
}

func (m *memList) RPop(ctx context.Context) (string, bool, error) {
	var value string
	var found bool
	err := m.withSliceLocked(func(list []string) ([]string, bool) {
		if len(list) == 0 {
			return nil, false
		}
		list, value, found = xslice.PopTail(list)
		return list, true
	})

	return value, found, err
}

func (m *memList) LRem(ctx context.Context, count int, element string) (int, error) {
	var deleted int
	err := m.withSliceLocked(func(list []string) ([]string, bool) {
		if count == 0 {
			list = slices.DeleteFunc(list, func(s string) bool {
				if s == element {
					deleted++
					return true
				}
				return false
			})
			return list, deleted > 0
		} else if count > 0 {
			newList := xslice.DeleteFuncN(list, func(s string) bool {
				return s == element
			}, count)
			deleted = len(list) - len(newList)
			return newList, deleted > 0
		} else { // if count < 0
			newList := xslice.RevDeleteFuncN(list, func(s string) bool {
				return s == element
			}, count*-1)
			deleted = len(list) - len(newList)
			return newList, deleted > 0
		}
	})
	return deleted, err
}

func (m *memList) Range(ctx context.Context, fn func(val string) bool) error {
	return m.LRange(ctx, fn)
}

func (m *memList) LRange(ctx context.Context, fn func(val string) bool) error {
	values, err := m.getSlice()
	if err != nil {
		return nil
	}
	for _, val := range values {
		if !fn(val) {
			return nil
		}
		if err = ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (m *memList) RRange(ctx context.Context, fn func(val string) bool) error {
	values, err := m.getSlice()
	if err != nil {
		return nil
	}
	for i := len(values) - 1; i >= 0; i-- {
		if !fn(values[i]) {
			return nil
		}
		if err = ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemoryStore) Hash(key string) Hash[string] {
	return &memHash{
		store: m,
		key:   key,
	}
}

var _ Hash[string] = (*memHash)(nil)

type memHash struct {
	store *MemoryStore
	key   string
}

func (m *memHash) withMapLocked(fn func(map[string]string) (map[string]string, bool)) error {
	var err error
	m.store.withLock(func(db map[string]string, tps map[string]internal.DataType) {
		str, found := db[m.key]
		tp := tps[m.key]
		if found && tp != internal.DataTypeHash {
			err = ErrInvalidType
			return
		}
		result := map[string]string{}
		if found {
			xcodec.JSON.Decode([]byte(str), &result)
		}
		newValue, replace := fn(result)
		if replace {
			bf, _ := xcodec.JSON.Encode(newValue)
			db[m.key] = string(bf)
			tps[m.key] = internal.DataTypeHash
		}
	})
	return err
}

func (m *memHash) HSet(ctx context.Context, field string, value string) error {
	return m.withMapLocked(func(m map[string]string) (map[string]string, bool) {
		m[field] = value
		return m, true
	})
}

func (m *memHash) HMSet(ctx context.Context, values map[string]string) error {
	return m.withMapLocked(func(m map[string]string) (map[string]string, bool) {
		for k, v := range values {
			m[k] = v
		}
		return m, true
	})
}

func (m *memHash) HGet(ctx context.Context, field string) (string, bool, error) {
	var value string
	var found bool
	err := m.withMapLocked(func(m map[string]string) (map[string]string, bool) {
		value, found = m[field]
		return m, false
	})
	return value, found, err
}

func (m *memHash) HDel(ctx context.Context, fields ...string) error {
	return m.withMapLocked(func(m map[string]string) (map[string]string, bool) {
		if len(m) == 0 {
			return m, false
		}
		for _, field := range fields {
			delete(m, field)
		}
		return m, false
	})
}

func (m *memHash) HRange(ctx context.Context, fn func(field string, value string) bool) error {
	values, err := m.HGetAll(ctx)
	if err != nil {
		return err
	}
	for key, value := range values {
		if !fn(key, value) {
			return nil
		}
	}
	return err
}

func (m *memHash) HGetAll(ctx context.Context) (map[string]string, error) {
	var result map[string]string
	err := m.withMapLocked(func(m map[string]string) (map[string]string, bool) {
		result = m
		return m, false
	})
	return result, err
}

func (m *MemoryStore) Set(key string) Set[string] {
	return &memSet{
		store: m,
		key:   key,
	}
}

var _ Set[string] = (*memSet)(nil)

type memSet struct {
	store *MemoryStore
	key   string
}

func (m *memSet) withSliceLocked(fn func([]string) ([]string, bool)) error {
	var err error
	m.store.withLock(func(db map[string]string, tps map[string]internal.DataType) {
		str, found := db[m.key]
		tp := tps[m.key]
		if found && tp != internal.DataTypeSet {
			err = ErrInvalidType
			return
		}
		var result []string
		if found {
			xcodec.JSON.Decode([]byte(str), &result)
		}
		newList, replace := fn(result)
		if replace {
			bf, _ := xcodec.JSON.Encode(newList)
			db[m.key] = string(bf)
			tps[m.key] = internal.DataTypeSet
		}
	})
	return err
}

func (m *memSet) SAdd(ctx context.Context, members ...string) (int, error) {
	var added int
	err := m.withSliceLocked(func(list []string) ([]string, bool) {
		for _, member := range members {
			if slices.Contains(list, member) {
				continue
			}
			added++
			list = append(list, member)
		}
		return list, added > 0
	})
	return added, err
}

func (m *memSet) SRem(ctx context.Context, members ...string) error {
	return m.withSliceLocked(func(list []string) ([]string, bool) {
		for _, member := range members {
			if !slices.Contains(list, member) {
				continue
			}
			list = xslice.DeleteValue(list, member)
		}
		return list, true
	})
}

func (m *memSet) SRange(ctx context.Context, fn func(val string) bool) error {
	list, err := m.SMembers(ctx)
	if err != nil {
		return err
	}
	for _, val := range list {
		if !fn(val) {
			return nil
		}
	}
	return nil
}

func (m *memSet) SMembers(ctx context.Context) ([]string, error) {
	var list []string
	err := m.withSliceLocked(func(values []string) ([]string, bool) {
		list = values
		return list, false
	})
	return list, err
}

func (m *MemoryStore) ZSet(key string) ZSet[string] {
	return &memZSet{
		store: m,
		key:   key,
	}
}

var _ ZSet[string] = (*memZSet)(nil)

type memZSet struct {
	store *MemoryStore
	key   string
}

func (m *memZSet) withLocked(fn func(*memZSetValue) (*memZSetValue, bool)) error {
	var err error
	m.store.withLock(func(db map[string]string, tps map[string]internal.DataType) {
		str, found := db[m.key]
		tp := tps[m.key]
		if found && tp != internal.DataTypeZSet {
			err = ErrInvalidType
			return
		}
		result := &memZSetValue{}
		if found {
			xcodec.JSON.Decode([]byte(str), result)
		}
		newList, replace := fn(result)
		if replace {
			bf, _ := xcodec.JSON.Encode(newList)
			db[m.key] = string(bf)
			tps[m.key] = internal.DataTypeZSet
		}
	})
	return err
}

func (m *memZSet) ZAdd(ctx context.Context, score float64, member string) error {
	return m.withLocked(func(zv *memZSetValue) (*memZSetValue, bool) {
		zv.add(score, member)
		return zv, true
	})
}

func (m *memZSet) ZScore(ctx context.Context, member string) (float64, bool, error) {
	var score float64
	var found bool
	err := m.withLocked(func(zv *memZSetValue) (*memZSetValue, bool) {
		score, found = zv.score(member)
		return zv, false
	})
	return score, found, err
}

func (m *memZSet) ZRange(ctx context.Context, fn func(member string, score float64) bool) error {
	var value *memZSetValue
	err := m.withLocked(func(zv *memZSetValue) (*memZSetValue, bool) {
		value = zv
		return zv, false
	})
	if err != nil {
		return err
	}
	if value != nil {
		for _, member := range value.Members {
			if !fn(member, value.Scores[member]) {
				return nil
			}
		}
	}
	return nil
}

func (m *memZSet) ZRem(ctx context.Context, members ...string) error {
	return m.withLocked(func(value *memZSetValue) (*memZSetValue, bool) {
		var changed bool
		for _, member := range members {
			if value.remove(member) {
				changed = true
			}
		}
		return value, changed
	})
}

type memZSetValue struct {
	Members []string           `json:"m"` // 按照 score 升序排序的 members 集合
	Scores  map[string]float64 `json:"s"`
}

func (mz *memZSetValue) add(score float64, member string) {
	if mz.Scores == nil {
		mz.Scores = make(map[string]float64)
	}
	mz.Scores[member] = score
	list := make([]memMemberScore, 0, len(mz.Scores))
	for k, v := range mz.Scores {
		list = append(list, memMemberScore{
			member: k,
			score:  v,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].score < list[j].score
	})
	mz.Members = make([]string, 0, len(mz.Scores))
	for _, item := range list {
		mz.Members = append(mz.Members, item.member)
	}
}

type memMemberScore struct {
	member string
	score  float64
}

func (mz *memZSetValue) score(member string) (float64, bool) {
	if len(mz.Scores) == 0 {
		return 0, false
	}
	score, found := mz.Scores[member]
	return score, found
}

func (mz *memZSetValue) remove(member string) bool {
	if len(mz.Members) == 0 {
		return false
	}
	_, found := mz.Scores[member]
	if !found {
		return false
	}
	delete(mz.Scores, member)
	mz.Members = xslice.DeleteValue(mz.Members, member)
	return true
}

func (m *MemoryStore) Delete(ctx context.Context, keys ...string) error {
	m.withLock(func(db map[string]string, tps map[string]internal.DataType) {
		for _, key := range keys {
			delete(db, key)
			delete(tps, key)
		}
	})
	return nil
}
