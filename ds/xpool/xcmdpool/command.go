//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-29

package xcmdpool

import (
	"context"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/xanygo/anygo/ds/xpool"
	"github.com/xanygo/anygo/ds/xsync"
)

// Command 将 Command 封装成一个 Server
// 输出通过 Command 的 stdin 传入，输出通过 stdout 传出
type Command struct {
	Path       string          // 命令地址,必填
	Args       []string        // 命令参数列表，可选
	Setup      func(*exec.Cmd) // 初始化 Cmd 的 回调，可选
	PoolOption *xpool.Option   // 对象次参数，可轩

	pool xpool.Pool[*command]

	once       sync.Once
	rootCtx    context.Context
	rootCancel context.CancelFunc
	ch         chan io.ReadWriteCloser
}

var _ io.Closer = (*command)(nil)

type command struct {
	cmd *exec.Cmd
	rw  io.ReadWriter
}

func (c *command) Close() error {
	return c.cmd.Process.Kill()
}

func (c *Command) initOnce() {
	c.once.Do(func() {
		c.rootCtx, c.rootCancel = context.WithCancel(context.Background())
		c.ch = make(chan io.ReadWriteCloser)
		fac := &commandFactory{P: c}
		c.pool = xpool.New[*command](c.PoolOption, fac)
	})
}

func (c *Command) Close() error {
	c.initOnce()
	c.rootCancel()
	c.pool.Close()
	return nil
}

// Spawn 获取一个空闲的读写对象，若没有空闲的会等待或创建新的
// 读写完成后，使必须调用 Close() 方法，释放资源
func (c *Command) Spawn(ctx context.Context) (io.ReadWriteCloser, error) {
	c.initOnce()
	child, err := c.pool.Get(ctx)
	if err != nil {
		return nil, err
	}
	cc := child.Raw()
	return &entry{
		pe: child,
		rw: cc.rw,
	}, nil
}

var _ io.ReadWriteCloser = (*entry)(nil)

type entry struct {
	pe     xpool.Entry[*command]
	rw     io.ReadWriter
	rwErr  xsync.Value[error]
	closed atomic.Bool
}

func (c *entry) Read(p []byte) (n int, err error) {
	if c.closed.Load() {
		return 0, io.ErrClosedPipe
	}
	n, err = c.rw.Read(p)
	if err != nil {
		c.rwErr.Store(err)
	}
	return n, err
}

func (c *entry) Write(p []byte) (n int, err error) {
	if c.closed.Load() {
		return 0, io.ErrClosedPipe
	}
	n, err = c.rw.Write(p)
	if err != nil {
		c.rwErr.Store(err)
	}
	return n, err
}

func (c *entry) Close() error {
	if c.closed.CompareAndSwap(false, true) {
		c.pe.Release(c.rwErr.Load())
	}
	return nil
}
