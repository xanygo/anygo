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

// NewRing 创建新的固定长度的 Ring，length 应 > 0
func NewRing[T any](length int) *Ring[T] {
	if length <= 0 {
		panic(fmt.Errorf("invalid Ring length %d", length))
	}
	return &Ring[T]{
		length: length,
		values: make([]T, length),
	}
}

// Ring 具有固定容量的，环形结构的 slice，容量满的情况下，新元素会覆盖老元素，非并发安全的
type Ring[T any] struct {
	values []T
	length int // 容量
	size   int // 有效元素个数
	tail   int // push 写入的索引位置
	head   int // pop 出栈的索引位置
}

// Push 添加新的元素，容量满的情况下，会覆盖老的值
func (r *Ring[T]) Push(values ...T) {
	if len(values) == 0 {
		return
	}
	for _, v := range values {
		r.values[r.tail] = v
		r.pushInc()
	}
}

func (r *Ring[T]) pushInc() {
	if r.size > 0 && r.tail == r.head {
		r.popInc()
	}
	r.tail++
	if r.tail == r.length {
		r.tail = 0
	}
	if r.size < r.length {
		r.size++
	}
}

func (r *Ring[T]) popInc() {
	r.head++
	if r.head == r.length {
		r.head = 0
	}
}

// PushSwap 添加并返回被替换的值
func (r *Ring[T]) PushSwap(v T) (old T, swapped bool) {
	if r.size > r.tail {
		old = r.values[r.tail]
		swapped = true
	}
	r.values[r.tail] = v
	r.pushInc()
	return old, swapped
}

// Pop 出栈
func (r *Ring[T]) Pop() (v T, ok bool) {
	if r.size == 0 {
		return v, false
	}
	v = r.values[r.head]
	var zero T
	r.values[r.head] = zero
	r.popInc()
	r.size--
	return v, true
}

func (r *Ring[T]) Cap() int {
	return r.length
}

func (r *Ring[T]) Len() int {
	return r.size
}

// Range 遍历，先加入的会先遍历
func (r *Ring[T]) Range(fn func(v T) bool) {
	if r.size == 0 {
		return
	}
	for i := 0; i < r.size; i++ {
		index := (i + r.head) % r.length
		if !fn(r.values[index]) {
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
	if r.size == 0 {
		return nil
	}
	vs := make([]T, 0, r.size)
	for i := 0; i < r.size; i++ {
		index := (i + r.head) % r.length
		vs = append(vs, r.values[index])
	}
	return vs
}

func (r *Ring[T]) Clear() {
	r.tail = 0
	r.size = 0
	r.head = 0
	clear(r.values)
}

// NewSyncRing 创建新的 SyncRing，length-容量，应 > 0
func NewSyncRing[T any](length int) *SyncRing[T] {
	if length <= 0 {
		panic(fmt.Errorf("invalid SyncRing length %d", length))
	}
	return &SyncRing[T]{
		length: length,
		values: make([]T, length),
		mux:    &sync.RWMutex{},
	}
}

// SyncRing 具有固定容量的，环形结构的 slice，容量满的情况下，新元素会覆盖老元素，是并发安全的
type SyncRing[T any] struct {
	values []T
	length int // 长度
	size   int // 有效元素个数
	tail   int // push 写入的索引位置
	head   int // pop 出栈的索引位置
	mux    *sync.RWMutex
}

// Push 添加新的元素，容量满的情况下，会覆盖老的值
func (r *SyncRing[T]) Push(values ...T) {
	if len(values) == 0 {
		return
	}
	r.mux.Lock()
	defer r.mux.Unlock()
	for _, v := range values {
		r.values[r.tail] = v
		r.tailInc()
	}
}

func (r *SyncRing[T]) tailInc() {
	if r.size > 0 && r.tail == r.head {
		r.headInc()
	}
	r.tail++
	if r.tail == r.length {
		r.tail = 0
	}
	if r.size < r.length {
		r.size++
	}
}

func (r *SyncRing[T]) headInc() {
	r.head++
	if r.head == r.length {
		r.head = 0
	}
}

// PushSwap 添加并返回被替换的值
func (r *SyncRing[T]) PushSwap(v T) (old T, swapped bool) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.size > r.tail {
		old = r.values[r.tail]
		swapped = true
	}
	r.values[r.tail] = v
	r.tailInc()
	return old, swapped
}

// Pop 出栈
func (r *SyncRing[T]) Pop() (v T, ok bool) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.size == 0 {
		return v, false
	}
	v = r.values[r.head]
	// 可选：清空引用，便于 GC
	var zeroValue T
	r.values[r.head] = zeroValue

	r.headInc()
	r.size--
	return v, true
}

func (r *SyncRing[T]) Cap() int {
	return r.length
}

func (r *SyncRing[T]) Len() int {
	r.mux.RLock()
	defer r.mux.RUnlock()
	return r.size
}

// Range 遍历，先加入的会先遍历
func (r *SyncRing[T]) Range(fn func(v T) bool) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	if r.size == 0 {
		return
	}
	for i := 0; i < r.size; i++ {
		index := (i + r.head) % r.length
		if !fn(r.values[index]) {
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

	if r.size == 0 {
		return nil
	}
	vs := make([]T, 0, r.size)
	for i := r.head; i < r.head+r.size; i++ {
		vs = append(vs, r.values[i%r.length])
	}
	return vs
}

func (r *SyncRing[T]) Clear() {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.tail = 0
	r.size = 0
	r.head = 0
	clear(r.values)
}

func NewUniqRing[T comparable](capacity int) *UniqRing[T] {
	if capacity <= 0 {
		panic(fmt.Errorf("invalid SyncRing length %d", capacity))
	}
	return &UniqRing[T]{
		capacity:   capacity,
		values:     make([]T, capacity),
		valueIndex: make(map[T]int, capacity),
	}
}

// UniqRing 具有唯一值的 ring list，非并发安全的
type UniqRing[T comparable] struct {
	values     []T
	valueIndex map[T]int
	capacity   int
	length     int
	index      int
}

// Push 添加新的元素，容量满的情况下，会覆盖老的值
func (r *UniqRing[T]) Push(values ...T) {
	for _, v := range values {
		oldIndex, has := r.valueIndex[v]
		if has {
			r.values[oldIndex] = v
			continue
		}

		r.values[r.index] = v
		r.valueIndex[v] = r.index
		r.index++
		if r.index == r.capacity {
			r.index = 0
		}
		if r.length < r.capacity {
			r.length++
		}
	}
}

// PushSwap 添加并返回被替换的值
func (r *UniqRing[T]) PushSwap(v T) (old T, swapped bool) {
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
	if r.index == r.capacity {
		r.index = 0
	}
	if r.length < r.capacity {
		r.length++
	}

	return old, swapped
}

func (r *UniqRing[T]) Cap() int {
	return r.capacity
}

func (r *UniqRing[T]) Len() int {
	return r.length
}

// Range 遍历，先加入的会先遍历
func (r *UniqRing[T]) Range(fn func(v T) bool) {
	if r.length == 0 {
		return
	}

	if r.length != r.capacity {
		for i := 0; i < r.length; i++ {
			if !fn(r.values[i]) {
				return
			}
		}
		return
	}

	// 容量满的情况下

	for i := r.index; i < r.capacity; i++ {
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
	if length != r.capacity {
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

func NewSyncUniqRing[T comparable](capacity int) *SyncUniqRing[T] {
	if capacity <= 0 {
		panic(fmt.Errorf("invalid SyncRing length %d", capacity))
	}
	return &SyncUniqRing[T]{
		capacity:   capacity,
		values:     make([]T, capacity),
		valueIndex: make(map[T]int, capacity),
		mux:        new(sync.RWMutex),
	}
}

// SyncUniqRing 具有唯一值的 ring list,是并发安全的
type SyncUniqRing[T comparable] struct {
	values     []T
	valueIndex map[T]int
	capacity   int
	length     int
	index      int
	mux        *sync.RWMutex
}

// Push 添加新的元素，容量满的情况下，会覆盖老的值
func (r *SyncUniqRing[T]) Push(values ...T) {
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
		if r.index == r.capacity {
			r.index = 0
		}
		if r.length < r.capacity {
			r.length++
		}
	}
}

// PushSwap 添加并返回被替换的值
func (r *SyncUniqRing[T]) PushSwap(v T) (old T, swapped bool) {
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
	if r.index == r.capacity {
		r.index = 0
	}
	if r.length < r.capacity {
		r.length++
	}

	return old, swapped
}

func (r *SyncUniqRing[T]) Cap() int {
	return r.capacity
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

	if r.length != r.capacity {
		for i := 0; i < r.length; i++ {
			if !fn(r.values[i]) {
				return
			}
		}
		return
	}

	// 容量满的情况下

	for i := r.index; i < r.capacity; i++ {
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
	if length != r.capacity {
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

func NewSyncRingWriter(capacity int) *SyncRingWriter {
	return &SyncRingWriter{
		sr: NewSyncRing[string](capacity),
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
		w.sr.Push(s)
	}
	return len(s), nil
}

func (w *SyncRingWriter) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		w.sr.Push(string(p))
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
