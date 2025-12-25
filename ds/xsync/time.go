// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123+git@gmail.com>
// Date: 2025/8/19

package xsync

import (
	"sync/atomic"
	"time"
)

// TimeStamp atomic store for time stamp
// without time Location
type TimeStamp int64

// Load atomic load time
func (at *TimeStamp) Load() time.Time {
	v := atomic.LoadInt64((*int64)(at))
	if v == 0 {
		return time.Time{}
	}
	return time.Unix(v/1e9, v%1e9)
}

// Store atomic store time stamp
func (at *TimeStamp) Store(n time.Time) {
	atomic.StoreInt64((*int64)(at), n.UnixNano())
}

// Sub returns the duration t-n
func (at *TimeStamp) Sub(n time.Time) time.Duration {
	v := atomic.LoadInt64((*int64)(at))
	return time.Duration(v - n.UnixNano())
}

// Since returns the time elapsed since n.
func (at *TimeStamp) Since(n time.Time) time.Duration {
	v := atomic.LoadInt64((*int64)(at))
	return time.Duration(n.UnixNano() - v)
}

// Before reports whether the time instant t is before u.
func (at *TimeStamp) Before(n time.Time) bool {
	v := atomic.LoadInt64((*int64)(at))
	return v < n.UnixNano()
}

// After reports whether the time instant t is after u.
func (at *TimeStamp) After(n time.Time) bool {
	v := atomic.LoadInt64((*int64)(at))
	return v > n.UnixNano()
}

func (at *TimeStamp) CompareAndSwap(old time.Time, new time.Time) bool {
	var n1 int64
	if !old.IsZero() {
		n1 = old.UnixNano()
	}
	return atomic.CompareAndSwapInt64((*int64)(at), n1, new.UnixNano())
}

type TimeDuration = Int64[time.Duration]

// Interval 控制某个操作的最小间隔时间。
// 它记录上一次操作的时间戳，并提供线程安全的方法判断操作是否允许。
type Interval struct {
	last int64
}

// Allow 判断是否可以执行操作。
// 如果距离上一次允许操作的时间已经超过指定的 dur，则返回 true 并更新 last 为当前时间。
// 否则返回 false，不更新 last。
//
// 参数:
//   - dur: 最小允许的时间间隔
//
// 返回值:
//   - bool: 当前操作是否被允许
//
// 注意:
//   - 该方法线程安全，适用于并发环境。
//   - 如果 dur 为零或负值，Allow 总是返回 true。
func (it *Interval) Allow(dur time.Duration) bool {
	old := atomic.LoadInt64(&it.last)
	expire := old + dur.Nanoseconds()
	now := time.Now().UnixNano()
	return expire < now && atomic.CompareAndSwapInt64(&it.last, old, now)
}
