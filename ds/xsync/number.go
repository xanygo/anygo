// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/19

package xsync

import (
	"strconv"
	"sync/atomic"
)

type Int64[T ~int64] int64

func (n *Int64[T]) Load() T {
	v := atomic.LoadInt64((*int64)(n))
	return T(v)
}

func (n *Int64[T]) Add(v T) {
	atomic.AddInt64((*int64)(n), int64(v))
}

func (n *Int64[T]) Store(v T) {
	atomic.StoreInt64((*int64)(n), int64(v))
}

func (n *Int64[T]) Swap(v T) (old T) {
	o := atomic.SwapInt64((*int64)(n), int64(v))
	return T(o)
}

func (n *Int64[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return atomic.CompareAndSwapInt64((*int64)(n), int64(old), int64(new))
}

func (n *Int64[T]) MarshalText() ([]byte, error) {
	txt := strconv.FormatInt(int64(n.Load()), 10)
	return []byte(txt), nil
}

func (n *Int64[T]) UnmarshalText(text []byte) error {
	num, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return err
	}
	n.Store(T(num))
	return nil
}

// -------------------------------------------------------------------------------

type Int32[T ~int32] int32

func (n *Int32[T]) Load() T {
	v := atomic.LoadInt32((*int32)(n))
	return T(v)
}

func (n *Int32[T]) Add(v T) {
	atomic.AddInt32((*int32)(n), int32(v))
}

func (n *Int32[T]) Store(v T) {
	atomic.StoreInt32((*int32)(n), int32(v))
}

func (n *Int32[T]) Swap(v T) (old T) {
	o := atomic.SwapInt32((*int32)(n), int32(v))
	return T(o)
}

func (n *Int32[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return atomic.CompareAndSwapInt32((*int32)(n), int32(old), int32(new))
}

func (n *Int32[T]) MarshalText() ([]byte, error) {
	txt := strconv.FormatInt(int64(n.Load()), 10)
	return []byte(txt), nil
}

func (n *Int32[T]) UnmarshalText(text []byte) error {
	num, err := strconv.ParseInt(string(text), 10, 32)
	if err != nil {
		return err
	}
	n.Store(T(num))
	return nil
}

// -------------------------------------------------------------------------------

type Uint64[T ~uint64] uint64

func (n *Uint64[T]) Load() T {
	v := atomic.LoadUint64((*uint64)(n))
	return T(v)
}

func (n *Uint64[T]) Add(v T) {
	atomic.AddUint64((*uint64)(n), uint64(v))
}

func (n *Uint64[T]) Store(v T) {
	atomic.StoreUint64((*uint64)(n), uint64(v))
}

func (n *Uint64[T]) Swap(v T) (old T) {
	o := atomic.SwapUint64((*uint64)(n), uint64(v))
	return T(o)
}

func (n *Uint64[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return atomic.CompareAndSwapUint64((*uint64)(n), uint64(old), uint64(new))
}

func (n *Uint64[T]) MarshalText() ([]byte, error) {
	txt := strconv.FormatUint(uint64(n.Load()), 10)
	return []byte(txt), nil
}

func (n *Uint64[T]) UnmarshalText(text []byte) error {
	num, err := strconv.ParseUint(string(text), 10, 64)
	if err != nil {
		return err
	}
	n.Store(T(num))
	return nil
}

// -------------------------------------------------------------------------------

type Uint32[T ~uint32] uint32

func (n *Uint32[T]) Load() T {
	v := atomic.LoadUint32((*uint32)(n))
	return T(v)
}

func (n *Uint32[T]) Add(v T) {
	atomic.AddUint32((*uint32)(n), uint32(v))
}

func (n *Uint32[T]) Store(v T) {
	atomic.StoreUint32((*uint32)(n), uint32(v))
}

func (n *Uint32[T]) Swap(v T) (old T) {
	o := atomic.SwapUint32((*uint32)(n), uint32(v))
	return T(o)
}

func (n *Uint32[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return atomic.CompareAndSwapUint32((*uint32)(n), uint32(old), uint32(new))
}

func (n *Uint32[T]) MarshalText() ([]byte, error) {
	txt := strconv.FormatUint(uint64(n.Load()), 10)
	return []byte(txt), nil
}

func (n *Uint32[T]) UnmarshalText(text []byte) error {
	num, err := strconv.ParseUint(string(text), 10, 32)
	if err != nil {
		return err
	}
	n.Store(T(num))
	return nil
}
