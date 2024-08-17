//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package safely

import (
	"context"
	"fmt"
	"sync"
)

func RecoveredVoid(re *PanicErr) {
	RecoveredCtx(context.Background(), re)
}

func RecoveredCtx(ctx context.Context, re *PanicErr) {
	defCallbacks.run(ctx, re)
}

var _ error = (*PanicErr)(nil)

type PanicErr struct {
	Recover any
}

func (p *PanicErr) Error() string {
	return fmt.Sprintf("panic: %v", p.Recover)
}

func OnRecovered(fn func(ctx context.Context, re *PanicErr)) {
	defCallbacks.add(fn)
}

var defCallbacks = &callbacks{}

type callbacks struct {
	fns []func(ctx context.Context, re *PanicErr)
	mux sync.RWMutex
}

func (c *callbacks) add(fn func(ctx context.Context, re *PanicErr)) {
	c.mux.Lock()
	c.fns = append(c.fns, fn)
	c.mux.Unlock()
}

func (c *callbacks) run(ctx context.Context, re *PanicErr) {
	c.mux.RLock()
	fns := c.fns
	c.mux.RUnlock()
	defer func() {
		_ = recover()
	}()
	for _, fn := range fns {
		fn(ctx, re)
	}
}
