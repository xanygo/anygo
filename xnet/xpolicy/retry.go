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
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/ds/xtype"
	"github.com/xanygo/anygo/xerror"
)

// Idempotent 请求多次发送，是否幂等，具体 RPC 协议的 Request 可选实现
type Idempotent interface {
	Idempotent() bool
}

var defaultRetry = &xsync.OnceInit[*Retry]{
	New: func() *Retry {
		return &Retry{
			Backoff:          FullJitter(50*time.Millisecond, 2*time.Second),
			RetryOnTemporary: true,
			RetryOnTimeout:   true,
			Retryable: func(ctx context.Context, req any, attempt int, code int64, err error) xtype.TriState {
				if xerror.IsNotFound(err) || xerror.IsInvalidParam(err) {
					return xtype.TriFalse
				}
				if ri, ok := req.(Idempotent); ok {
					if ri.Idempotent() {
						return xtype.TriTrue
					}
					return xtype.TriFalse
				}
				return xtype.TriNull
			},
		}
	},
}

// DefaultRetry 默认重试策略
func DefaultRetry() *Retry {
	return defaultRetry.Load()
}

func SetDefaultRetry(p *Retry) {
	defaultRetry.Store(p)
}

var alwaysRetry = &xsync.OnceInit[*Retry]{
	New: func() *Retry {
		return &Retry{
			Backoff: FullJitter(50*time.Millisecond, 2*time.Second),
			Retryable: func(ctx context.Context, req any, attempt int, code int64, err error) xtype.TriState {
				return xtype.TriTrue
			},
		}
	},
}

func AlwaysRetry() *Retry {
	return alwaysRetry.Load()
}

// Retry 重试策略
type Retry struct {
	Backoff func(attempt int) time.Duration

	RetryOnTimeout   bool // 是否对超时重试
	RetryOnTemporary bool // 是否对临时错误重试

	RetryErrorCodes []int64
	Retryable       func(ctx context.Context, req any, attempt int, code int64, err error) xtype.TriState

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

func (rp *Retry) IsRetryable(ctx context.Context, req any, attempt int, err error) bool {
	if rp == nil {
		return false
	}
	rp.once.Do(rp.init)

	code, ok := xerror.ErrCode2(err)
	if ok && rp.retryErrorCodes[code] {
		return true
	}

	if rp.Retryable != nil {
		if state := rp.Retryable(ctx, req, attempt, code, err); state.NotNull() {
			return state.IsTrue()
		}
	}

	if rp.RetryOnTemporary {
		var et xerror.TemporaryFailure
		if errors.As(err, &et) {
			// rpc client 各个协议的实现，可以返回这种 error，以标记是否临时错误
			return et.Temporary()
		}
	}

	if rp.RetryOnTimeout {
		var tt typeTimeout
		if errors.As(err, &tt) && tt.Timeout() {
			return true
		}
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
