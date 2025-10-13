//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-06

package xoption

import "time"

func Get(opt Reader, key Key) (any, bool) {
	if opt == nil {
		return nil, false
	}
	return opt.Get(key)
}

func GetAs[V any](opt Reader, key Key) (e V, b bool) {
	val, ok := Get(opt, key)
	if !ok {
		return e, false
	}
	vt, ok := val.(V)
	return vt, ok
}

func GetAsDefault[V any](opt Reader, key Key, def V) V {
	val, ok := GetAs[V](opt, key)
	if ok {
		return val
	}
	return def
}

func String(opt Reader, key Key, def string) string {
	return GetAsDefault[string](opt, key, def)
}

func Int64(opt Reader, key Key, def int64) int64 {
	return GetAsDefault[int64](opt, key, def)
}

func Int32(opt Reader, key Key, def int32) int32 {
	return GetAsDefault[int32](opt, key, def)
}

func Int16(opt Reader, key Key, def int16) int16 {
	return GetAsDefault[int16](opt, key, def)
}

func Int8(opt Reader, key Key, def int8) int8 {
	return GetAsDefault[int8](opt, key, def)
}

func Int(opt Reader, key Key, def int) int {
	return GetAsDefault[int](opt, key, def)
}

func Uint64(opt Reader, key Key, def uint64) uint64 {
	return GetAsDefault[uint64](opt, key, def)
}

func Uint32(opt Reader, key Key, def uint32) uint32 {
	return GetAsDefault[uint32](opt, key, def)
}

func Uint16(opt Reader, key Key, def uint16) uint16 {
	return GetAsDefault[uint16](opt, key, def)
}

func Uint8(opt Reader, key Key, def uint8) uint8 {
	return GetAsDefault[uint8](opt, key, def)
}

func Uint(opt Reader, key Key, def uint) uint {
	return GetAsDefault[uint](opt, key, def)
}

func Duration(opt Reader, key Key, def time.Duration) time.Duration {
	return GetAsDefault[time.Duration](opt, key, def)
}

func Map(opt Reader, key Key, def map[string]any) map[string]any {
	return GetAsDefault[map[string]any](opt, key, def)
}

func Bool(opt Reader, key Key, def bool) bool {
	return GetAsDefault[bool](opt, key, def)
}

type KeyValue[K comparable, V any] struct {
	K K
	V V
}
