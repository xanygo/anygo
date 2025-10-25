//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package xslice

import (
	"fmt"
	"io"
	"iter"
	"strings"
	"sync"
)

// NewRing 创建新的 Ring，caption-容量，应 > 0
func NewRing[T any](caption int) *Ring[T] {
	if caption <= 0 {
		panic(fmt.Errorf("invalid Ring caption %d", caption))
	}
	return &Ring[T]{
		caption: caption,
		values:  make([]T, caption),
	}
}

// Ring 具有指定最大容量的，环形结构的 slice，容量满的情况下，新元素会覆盖老元素，非并发安全的
type Ring[T any] struct {
	values  []T
	caption int // 容量
	length  int // 长度
	index   int
}

// Add 添加新的元素，容量满的情况下，会覆盖老的值
func (r *Ring[T]) Add(values ...T) {
	if len(values) == 0 {
		return
	}
	for _, v := range values {
		r.values[r.index] = v
		r.index++
		if r.index == r.caption {
			r.index = 0
		}
		if r.length < r.caption {
			r.length++
		}
	}
}

// AddSwap 添加并返回被替换的值
func (r *Ring[T]) AddSwap(v T) (old T, swapped bool) {
	if r.length > r.index {
		old = r.values[r.index]
		swapped = true
	}
	r.values[r.index] = v
	r.index++
	if r.index == r.caption {
		r.index = 0
	}
	if r.length < r.caption {
		r.length++
	}
	return old, swapped
}

func (r *Ring[T]) Len() int {
	return r.length
}

// Range 遍历，先加入的会先遍历
func (r *Ring[T]) Range(fn func(v T) bool) {
	if r.length == 0 {
		return
	}

	if r.length != r.caption {
		for i := 0; i < r.length; i++ {
			if !fn(r.values[i]) {
				return
			}
		}
		return
	}

	// 容量满的情况下
	for i := r.index; i < r.caption; i++ {
		if !fn(r.values[i]) {
			return
		}
	}

	for i := 0; i < r.index; i++ {
		if !fn(r.values[i]) {
			return
		}
	}
}

func (r *Ring[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		r.Range(yield)
	}
}

// Values 返回所有值，先加入的排在前面
func (r *Ring[T]) Values() []T {
	length := r.length
	if length == 0 {
		return nil
	}
	vs := make([]T, 0, length)
	if length != r.caption {
		vs = append(vs, r.values[:length]...)
		return vs
	}
	// 容量满的情况下
	vs = append(vs, r.values[r.index:]...)
	vs = append(vs, r.values[:r.index]...)
	return vs
}

func (r *Ring[T]) Clear() {
	r.index = 0
	r.length = 0
	clear(r.values)
}

// NewSyncRing 创建新的 SyncRing，caption-容量，应 > 0
func NewSyncRing[T any](caption int) *SyncRing[T] {
	if caption <= 0 {
		panic(fmt.Errorf("invalid SyncRing caption %d", caption))
	}
	return &SyncRing[T]{
		caption: caption,
		values:  make([]T, caption),
		mux:     &sync.RWMutex{},
	}
}

// SyncRing 具有指定最大容量的，环形结构的 slice，容量满的情况下，新元素会覆盖老元素，是并发安全的
type SyncRing[T any] struct {
	values  []T
	caption int
	length  int
	index   int
	mux     *sync.RWMutex
}

// Add 添加新的元素，容量满的情况下，会覆盖老的值
func (r *SyncRing[T]) Add(values ...T) {
	if len(values) == 0 {
		return
	}
	r.mux.Lock()
	for _, v := range values {
		r.values[r.index] = v
		r.index++
		if r.index == r.caption {
			r.index = 0
		}
		if r.length < r.caption {
			r.length++
		}
	}
	r.mux.Unlock()
}

// AddSwap 添加并返回被替换的值
func (r *SyncRing[T]) AddSwap(v T) (old T, swapped bool) {
	r.mux.Lock()
	if r.length > r.index {
		old = r.values[r.index]
		swapped = true
	}
	r.values[r.index] = v
	r.index++
	if r.index == r.caption {
		r.index = 0
	}
	if r.length < r.caption {
		r.length++
	}
	r.mux.Unlock()
	return old, swapped
}

func (r *SyncRing[T]) Len() int {
	r.mux.RLock()
	val := r.length
	r.mux.RUnlock()
	return val
}

// Range 遍历，先加入的会先遍历
func (r *SyncRing[T]) Range(fn func(v T) bool) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if r.length == 0 {
		return
	}

	if r.length != r.caption {
		for i := 0; i < r.length; i++ {
			if !fn(r.values[i]) {
				return
			}
		}
		return
	}

	// 容量满的情况下

	for i := r.index; i < r.caption; i++ {
		if !fn(r.values[i]) {
			return
		}
	}

	for i := 0; i < r.index; i++ {
		if !fn(r.values[i]) {
			return
		}
	}
}

func (r *SyncRing[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		r.Range(yield)
	}
}

// Values 返回所有值，先加入的排在前面
func (r *SyncRing[T]) Values() []T {
	r.mux.RLock()
	defer r.mux.RUnlock()
	length := r.length
	if length == 0 {
		return nil
	}
	vs := make([]T, 0, length)
	if length != r.caption {
		vs = append(vs, r.values[:length]...)
		return vs
	}
	// 容量满的情况下
	vs = append(vs, r.values[r.index:]...)
	vs = append(vs, r.values[:r.index]...)
	return vs
}

func (r *SyncRing[T]) Clear() {
	r.mux.Lock()
	r.length = 0
	r.index = 0
	clear(r.values)
	r.mux.Unlock()
}

func NewUniqRing[T comparable](caption int) *UniqRing[T] {
	if caption <= 0 {
		panic(fmt.Errorf("invalid SyncRing caption %d", caption))
	}
	return &UniqRing[T]{
		caption:    caption,
		values:     make([]T, caption),
		valueIndex: make(map[T]int, caption),
	}
}

// UniqRing 具有唯一值的 ring list，非并发安全的
type UniqRing[T comparable] struct {
	values     []T
	valueIndex map[T]int
	caption    int
	length     int
	index      int
}

// Add 添加新的元素，容量满的情况下，会覆盖老的值
func (r *UniqRing[T]) Add(values ...T) {
	for _, v := range values {
		oldIndex, has := r.valueIndex[v]
		if has {
			r.values[oldIndex] = v
			continue
		}

		r.values[r.index] = v
		r.valueIndex[v] = r.index
		r.index++
		if r.index == r.caption {
			r.index = 0
		}
		if r.length < r.caption {
			r.length++
		}
	}
}

// AddSwap 添加并返回被替换的值
func (r *UniqRing[T]) AddSwap(v T) (old T, swapped bool) {
	oldIndex, has := r.valueIndex[v]
	if has {
		old = r.values[oldIndex]
		r.values[oldIndex] = v
		return old, true
	}

	if r.length > r.index {
		old = r.values[r.index]
		swapped = true
	}
	r.values[r.index] = v
	r.valueIndex[v] = r.index
	r.index++
	if r.index == r.caption {
		r.index = 0
	}
	if r.length < r.caption {
		r.length++
	}

	return old, swapped
}

func (r *UniqRing[T]) Len() int {
	return r.length
}

// Range 遍历，先加入的会先遍历
func (r *UniqRing[T]) Range(fn func(v T) bool) {
	if r.length == 0 {
		return
	}

	if r.length != r.caption {
		for i := 0; i < r.length; i++ {
			if !fn(r.values[i]) {
				return
			}
		}
		return
	}

	// 容量满的情况下

	for i := r.index; i < r.caption; i++ {
		if !fn(r.values[i]) {
			return
		}
	}

	for i := 0; i < r.index; i++ {
		if !fn(r.values[i]) {
			return
		}
	}
}

func (r *UniqRing[T]) Iter() iter.Seq[T] {
	return func(yield func(v T) bool) {
		r.Range(yield)
	}
}

// Values 返回所有值，先加入的排在前面
func (r *UniqRing[T]) Values() []T {
	length := r.length
	if length == 0 {
		return nil
	}
	vs := make([]T, 0, length)
	if length != r.caption {
		vs = append(vs, r.values[:length]...)
		return vs
	}
	// 容量满的情况下
	vs = append(vs, r.values[r.index:]...)
	vs = append(vs, r.values[:r.index]...)
	return vs
}

func (r *UniqRing[T]) Clear() {
	r.length = 0
	r.index = 0
	clear(r.values)
	clear(r.valueIndex)
}

func NewSyncUniqRing[T comparable](caption int) *SyncUniqRing[T] {
	if caption <= 0 {
		panic(fmt.Errorf("invalid SyncRing caption %d", caption))
	}
	return &SyncUniqRing[T]{
		caption:    caption,
		values:     make([]T, caption),
		valueIndex: make(map[T]int, caption),
		mux:        new(sync.RWMutex),
	}
}

// SyncUniqRing 具有唯一值的 ring list,是并发安全的
type SyncUniqRing[T comparable] struct {
	values     []T
	valueIndex map[T]int
	caption    int
	length     int
	index      int
	mux        *sync.RWMutex
}

// Add 添加新的元素，容量满的情况下，会覆盖老的值
func (r *SyncUniqRing[T]) Add(values ...T) {
	r.mux.Lock()
	defer r.mux.Unlock()
	for _, v := range values {
		oldIndex, has := r.valueIndex[v]
		if has {
			r.values[oldIndex] = v
			continue
		}

		r.values[r.index] = v
		r.valueIndex[v] = r.index
		r.index++
		if r.index == r.caption {
			r.index = 0
		}
		if r.length < r.caption {
			r.length++
		}
	}
}

// AddSwap 添加并返回被替换的值
func (r *SyncUniqRing[T]) AddSwap(v T) (old T, swapped bool) {
	r.mux.Lock()
	defer r.mux.Unlock()

	oldIndex, has := r.valueIndex[v]
	if has {
		old = r.values[oldIndex]
		r.values[oldIndex] = v
		return old, true
	}

	if r.length > r.index {
		old = r.values[r.index]
		swapped = true
	}
	r.values[r.index] = v
	r.valueIndex[v] = r.index
	r.index++
	if r.index == r.caption {
		r.index = 0
	}
	if r.length < r.caption {
		r.length++
	}

	return old, swapped
}

func (r *SyncUniqRing[T]) Len() int {
	r.mux.RLock()
	val := r.length
	r.mux.RUnlock()
	return val
}

// Range 遍历，先加入的会先遍历
func (r *SyncUniqRing[T]) Range(fn func(v T) bool) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if r.length == 0 {
		return
	}

	if r.length != r.caption {
		for i := 0; i < r.length; i++ {
			if !fn(r.values[i]) {
				return
			}
		}
		return
	}

	// 容量满的情况下

	for i := r.index; i < r.caption; i++ {
		if !fn(r.values[i]) {
			return
		}
	}

	for i := 0; i < r.index; i++ {
		if !fn(r.values[i]) {
			return
		}
	}
}

func (r *SyncUniqRing[T]) Iter() iter.Seq[T] {
	return func(yield func(v T) bool) {
		r.Range(yield)
	}
}

// Values 返回所有值，先加入的排在前面
func (r *SyncUniqRing[T]) Values() []T {
	r.mux.RLock()
	defer r.mux.RUnlock()
	length := r.length
	if length == 0 {
		return nil
	}
	vs := make([]T, 0, length)
	if length != r.caption {
		vs = append(vs, r.values[:length]...)
		return vs
	}
	// 容量满的情况下
	vs = append(vs, r.values[r.index:]...)
	vs = append(vs, r.values[:r.index]...)
	return vs
}

func (r *SyncUniqRing[T]) Clear() {
	r.mux.Lock()
	r.length = 0
	r.index = 0
	clear(r.values)
	clear(r.valueIndex)
	r.mux.Unlock()
}

func NewSyncRingWriter(caption int) *SyncRingWriter {
	return &SyncRingWriter{
		sr: NewSyncRing[string](caption),
	}
}

var _ io.Writer = (*SyncRingWriter)(nil)
var _ io.StringWriter = (*SyncRingWriter)(nil)

// SyncRingWriter 一个会循环覆盖的 Writer，并发安全的
type SyncRingWriter struct {
	sr *SyncRing[string]
}

func (w *SyncRingWriter) WriteString(s string) (n int, err error) {
	if len(s) > 0 {
		w.sr.Add(s)
	}
	return len(s), nil
}

func (w *SyncRingWriter) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		w.sr.Add(string(p))
	}
	return len(p), nil
}

func (w *SyncRingWriter) Bytes() []byte {
	return []byte(w.String())
}

func (w *SyncRingWriter) String() string {
	return strings.Join(w.sr.Values(), "")
}

func (w *SyncRingWriter) Lines() []string {
	return w.sr.Values()
}

func (w *SyncRingWriter) Reset() {
	w.sr.Clear()
}
