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
	Re any
}

func (p *PanicErr) Error() string {
	return fmt.Sprintf("panic: %v", p.Re)
}

// OnRecovered 注册 panic 被自动 recover 之后的回调函数
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
