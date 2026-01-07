//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-07

package xpolicy

import (
	"context"
	"errors"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/xerror"
)

var defaultRetry = &Retry{
	Backoff:          FullJitter(50*time.Millisecond, 2*time.Second),
	RetryOnTemporary: true,
	RetryOnTimeout:   true,
	Retryable: func(ctx context.Context, attempt int, code int64, err error) bool {
		if xerror.IsNotFound(err) || xerror.IsInvalidParam(err) {
			return false
		}
		return true
	},
}

func DefaultRetry() *Retry {
	return defaultRetry
}

// Retry 重试策略
type Retry struct {
	Backoff func(attempt int) time.Duration

	RetryOnTimeout   bool // 是否对超时重试
	RetryOnTemporary bool // 是否对临时错误重试

	RetryErrorCodes []int64
	Retryable       func(ctx context.Context, attempt int, code int64, err error) bool

	once            sync.Once
	retryErrorCodes map[int64]bool
}

func (rp *Retry) GetBackoff(attempt int) time.Duration {
	if rp == nil || rp.Backoff == nil {
		return 0
	}
	return rp.Backoff(attempt)
}

type (
	typeTimeout interface {
		Timeout() bool
	}
)

func (rp *Retry) init() {
	rp.retryErrorCodes = xslice.ToMap(rp.RetryErrorCodes, true)
}

func (rp *Retry) IsRetryable(ctx context.Context, attempt int, err error) bool {
	if rp == nil {
		return false
	}
	rp.once.Do(rp.init)

	code, ok := xerror.ErrCode2(err)
	if ok && rp.retryErrorCodes[code] {
		return true
	}

	if rp.RetryOnTimeout {
		var tt typeTimeout
		if errors.As(err, &tt) && tt.Timeout() {
			return true
		}
	}

	if rp.RetryOnTemporary && xerror.IsTemporary(err) {
		return true
	}
	if rp.Retryable != nil {
		return rp.Retryable(ctx, attempt, code, err)
	}
	return false
}

func FullJitter(base time.Duration, max time.Duration) func(attempt int) time.Duration {
	return func(attempt int) time.Duration {
		exp := base * time.Duration(1<<attempt)
		if exp > max {
			exp = max
		}
		return time.Duration(rand.Int64N(int64(exp)))
	}
}
