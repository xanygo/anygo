package xsync

import (
	"sync"
	"time"
)

// Every 返回一个带周期缓存的函数。
//
//	第一次调用时会立即执行 do，并缓存返回值，在 cycle 时间内再次调用时，直接返回已缓存的结果。
//	缓存周期结束后，下一次调用会重新执行 do，并更新缓存。
//
//	当 cycle 小于等于 0 时，每次调用都会执行 do。
func Every[T any](do func() T, cycle time.Duration) func() T {
	var (
		mu    sync.Mutex
		last  time.Time
		value T
	)

	return func() T {
		mu.Lock()
		defer mu.Unlock()

		if !last.IsZero() && time.Since(last) < cycle {
			return value
		}

		value = do()
		last = time.Now()

		return value
	}
}

// EveryErr 返回一个带周期缓存的函数。
//
//	第一次调用时会立即执行 do，并缓存返回值，在 cycle 时间内再次调用时，直接返回已缓存的结果。
//	缓存周期结束后，下一次调用会重新执行 do，并更新缓存。
//
//	当 cycle 小于等于 0 时，每次调用都会执行 do。
func EveryErr[T any](do func() (T, error), cycle time.Duration) func() (T, error) {
	var (
		mu    sync.Mutex
		last  time.Time
		value T
		err   error
	)

	return func() (T, error) {
		mu.Lock()
		defer mu.Unlock()

		if !last.IsZero() && time.Since(last) < cycle {
			return value, err
		}

		value, err = do()
		last = time.Now()

		return value, err
	}
}

// EveryErrFor 返回一个带周期缓存的函数，支持为成功和失败结果分别设置缓存周期。
//
//	第一次调用时会立即执行 do，并根据执行结果缓存返回值和错误。
//
//	在对应的缓存周期内再次调用时，直接返回已缓存的结果，不会执行 do。
//	对应的缓存周期结束后，下一次调用会重新执行 do，并根据新的执行结果更新缓存。
//
// 返回的函数支持并发调用。
//
//	当 cycleSuccess 或 cycleFail <= 0 时，其对应结果不会被缓存。
func EveryErrFor[T any](do func() (T, error), cycleSuccess time.Duration, cycleFail time.Duration) func() (T, error) {
	var (
		mu        sync.Mutex
		value     T
		err       error
		expiresAt time.Time
	)

	return func() (T, error) {
		mu.Lock()
		defer mu.Unlock()

		now := time.Now()

		if !expiresAt.IsZero() && now.Before(expiresAt) {
			return value, err
		}

		value, err = do()

		cycle := cycleSuccess
		if err != nil {
			cycle = cycleFail
		}

		if cycle > 0 {
			expiresAt = time.Now().Add(cycle)
		} else {
			expiresAt = time.Time{}
		}

		return value, err
	}
}

// Debounce 返回一个防抖函数。
//
//	每次调用都会重新开始计时。如果在 delay 时间内再次调用，则取消之前的等待并重新计时。
//	只有在连续调用停止，并持续等待 delay 时间后，do 才会执行一次。
//	因此，在短时间内连续调用多次时，通常只有最后一次调用会触发 do。
//	返回的函数支持并发调用。
//
//	当 delay 小于等于 0 时，每次调用都会异步执行 do。
func Debounce(do func(), delay time.Duration) func() {
	var (
		mu    sync.Mutex
		timer *time.Timer
		seq   uint64
	)

	return func() {
		mu.Lock()
		seq++
		current := seq

		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(delay, func() {
			mu.Lock()
			if current != seq {
				mu.Unlock()
				return
			}
			timer = nil
			mu.Unlock()

			do()
		})

		mu.Unlock()
	}
}

// Throttle 返回一个节流函数,每间隔 interval 时间段 do 方法才执行一次
//
//	返回的函数支持并发调用。
//	当 interval 小于等于 0 时，每次调用都会执行 do。
func Throttle(do func(), interval time.Duration) func() {
	var mu sync.Mutex
	var last time.Time

	return func() {
		now := time.Now()

		mu.Lock()
		if interval > 0 && !last.IsZero() && now.Sub(last) < interval {
			mu.Unlock()
			return
		}

		last = now
		mu.Unlock()

		do()
	}
}
