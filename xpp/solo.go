//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xpp

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/safely"
)

type SoloTask struct {
	running  atomic.Bool
	disabled atomic.Bool
}

func (st *SoloTask) Execute(ctx context.Context, do func(ctx context.Context), life time.Duration, cycle time.Duration) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	if st.disabled.Load() {
		return
	}

	if !st.running.CompareAndSwap(false, true) {
		return
	}
	tm := time.NewTimer(0)
	fn := func(ctx context.Context) {
		defer st.running.Store(false)
		defer tm.Stop()
		ctx1, cancel := context.WithTimeout(ctx, life)
		defer cancel()
		for !st.disabled.Load() {
			select {
			case <-ctx.Done():
				return
			case <-tm.C:
				safely.RunCtx(ctx1, do)
				tm.Reset(cycle)
			}
		}
	}
	go fn(ctx)
}

func (st *SoloTask) SetEnabled(enabled bool) {
	st.disabled.Store(!enabled)
}
