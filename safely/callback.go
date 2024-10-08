//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package safely

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var recoverID atomic.Int64

// NewRecoverID 每 recover 一次，调用返回一个新的自增长 id
func NewRecoverID() int64 {
	return recoverID.Add(1)
}

// RecoverTotal 返回 recover 总次数
func RecoverTotal() int64 {
	return recoverID.Load()
}

// Recovered 接收 panic 后 recover() 的信息，之后会触发使用 OnRecovered 注册的回调函数
func Recovered(re any, data ...any) {
	RecoveredPE(NewPanicErr(re, 2, data...))
}

// RecoveredCtx 接收 panic 后 recover() 的信息，之后会触发使用 OnRecovered 注册的回调函数
func RecoveredCtx(ctx context.Context, re any, data ...any) {
	RecoveredPECtx(ctx, NewPanicErr(re, 2, data...))
}

// RecoveredPE 接收 panic recover() 后，封装的 PanicErr 信息，之后会触发使用 OnRecovered 注册的回调函数
func RecoveredPE(err *PanicErr) {
	RecoveredPECtx(context.Background(), err)
}

// RecoveredPECtx 接收 panic recover() 后，封装的 PanicErr 信息，之后会触发使用 OnRecovered 注册的回调函数
func RecoveredPECtx(ctx context.Context, err *PanicErr) {
	defCallbacks.run(ctx, err)
}

// NewPanicErr 创建一个新的 PanicErr 对象
//
//	re: recover() 的内容不应该为 nil
//	callerSkip: 追踪触发 panic 位置时，应跳过的调用层次
func NewPanicErr(re any, callerSkip int, data ...any) *PanicErr {
	_, file, line, _ := runtime.Caller(callerSkip)
	return &PanicErr{
		ID:    NewRecoverID(),
		Re:    re,
		Stack: debug.Stack(),
		File:  file,
		Line:  line,
		Data:  data,
	}
}

var _ error = (*PanicErr)(nil)

// PanicErr 一次 panic 的信息，以实现 error 接口
type PanicErr struct {
	ID    int64  // recover id
	Re    any    // recover() 的内容
	Stack []byte // 堆栈信息
	File  string // panic 或者 recover 的文件名
	Line  int    // panic 或者 recover 的文件行
	Data  []any  // 其他数据
}

func (p *PanicErr) Error() string {
	return fmt.Sprintf("panic: %v", p.Re)
}

// OnRecovered 注册 panic  recover() 之后的回调函数
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
