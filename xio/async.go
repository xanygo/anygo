// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package xio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

var _ io.WriteCloser = (*AsyncWriter)(nil)

// AsyncWriter 支持异步写的 Writer
type AsyncWriter struct {
	// Writer 实际 writer，必填
	Writer io.Writer

	// ChanSize 异步队列大小，可选
	// 默认为 1024。当值为 -1 时，chanSize=0，即变为同步
	ChanSize int

	// NeedStatus 是否需要 write 的状态
	NeedStatus bool

	writeStats chan WriteStatus
	buffers    chan []byte   // 异步数据
	loopExit   chan struct{} // 异步写完成后的事件

	once   sync.Once   // 用于初始化
	mux    sync.Mutex  // 初始化时的锁
	closed atomic.Bool // 是否已经调用过 Close
}

var errClosed = errors.New("writer already closed")

func (aw *AsyncWriter) getChanSize() int {
	if aw.ChanSize > 0 {
		return aw.ChanSize
	} else if aw.ChanSize == -1 {
		return 0
	}
	return 1024
}

// Write 异步写
func (aw *AsyncWriter) Write(p []byte) (n int, err error) {
	if aw.closed.Load() {
		return 0, errClosed
	}
	if len(p) == 0 {
		return 0, nil
	}
	aw.once.Do(aw.init)

	select {
	case aw.buffers <- bytes.Clone(p):
		return len(p), nil
	case <-aw.loopExit:
		return 0, errClosed
	}
}

func (aw *AsyncWriter) init() {
	aw.mux.Lock()
	defer aw.mux.Unlock()

	aw.loopExit = make(chan struct{})
	aw.buffers = make(chan []byte, aw.getChanSize())
	aw.writeStats = make(chan WriteStatus, aw.getChanSize())
	go func() {
		defer func() {
			if re := recover(); re != nil {
				aw.onRecover("AsyncWriter.loop", re)
			}
		}()
		defer close(aw.loopExit)

		for {
			aw.doLoop()
			// 放在后面判断，以避免 aw.closed 在高并发下还未执行 doLoop，closed的状态就发生变化
			if aw.closed.Load() {
				break
			}
		}
	}()
}

func (aw *AsyncWriter) onRecover(msg string, re any) {
	err := fmt.Errorf("%s  panic %v", msg, re)
	s := WriteStatus{
		Err: err,
	}
	select {
	case aw.writeStats <- s:
	default:
	}
}

func (aw *AsyncWriter) doLoop() {
	defer func() {
		if re := recover(); re != nil {
			aw.onRecover("AsyncWriter.doLoop", re)
		}
	}()

	for b := range aw.buffers {
		if b == nil {
			break
		}
		n, err := aw.Writer.Write(b)
		if aw.NeedStatus {
			s := WriteStatus{
				Wrote: n,
				Err:   err,
			}
			select {
			case aw.writeStats <- s:
			default:
			}
		}
	}
}

// WriteStatus 返回的是异步写的最新一次的状态
func (aw *AsyncWriter) WriteStatus() <-chan WriteStatus {
	aw.mux.Lock()
	defer aw.mux.Unlock()
	if aw.writeStats == nil {
		c := make(chan WriteStatus)
		close(c)
		return c
	}
	return aw.writeStats
}

// Close 关闭
func (aw *AsyncWriter) Close() error {
	if aw.closed.CompareAndSwap(false, true) {
		aw.mux.Lock()
		defer aw.mux.Unlock()
		if aw.buffers != nil {
			aw.buffers <- nil
			<-aw.loopExit
			close(aw.writeStats)
		}
	}
	return nil
}

// WriteStatus status for Write
type WriteStatus struct {
	Err   error
	Wrote int
}
