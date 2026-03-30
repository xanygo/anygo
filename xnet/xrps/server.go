// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/6/25

package xrps

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

type (
	TCPListener = Listener[net.Conn]

	TCPHandler = Handler[net.Conn]

	TCPServer = Server[net.Conn]

	TCPAnyServer = AnyServer[net.Conn]
)

var _ Listener[net.Conn] = (net.Listener)(nil)

type Listener[C any] interface {
	Accept() (C, error)
	Close() error
}

type Handler[C any] interface {
	Handle(ctx context.Context, conn C)
}

type HandleFunc[C any] func(ctx context.Context, conn C)

func (hf HandleFunc[C]) Handle(ctx context.Context, conn C) {
	hf(ctx, conn)
}

type Server[C any] interface {
	Serve(l Listener[C]) error
}

// CanShutdown 支持优雅关闭
type CanShutdown interface {
	Shutdown(ctx context.Context) error
}

var _ CanShutdown = (*AnyServer[net.Conn])(nil)

// AnyServer 一个通用的 server
type AnyServer[C any] struct {
	// Handler 处理请求的 Handler，必填
	Handler Handler[C]

	// BeforeAccept Accept 之前的回调，可选
	BeforeAccept func(l Listener[C]) error

	// OnConn 创建新链接后的回调，可选
	OnConn func(ctx context.Context, conn C, err error) (context.Context, C, error)

	closeCancel context.CancelFunc
	serverExit  chan bool

	connections sync.Map

	status atomic.Int32
}

const (
	statusInit    int32 = iota // server 状态，初始状态
	statusRunning              // 已经调用 Serve 方法，处于运行中
	statusClosed               // 已经调用 Shutdown 方法，server 已经关闭
)

func statusTxt(s int32) string {
	switch s {
	case statusInit:
		return "init"
	case statusRunning:
		return "running"
	case statusClosed:
		return "closed"
	default:
		return "invalid status"
	}
}

var (
	ErrShutdown = errors.New("server shutdown")
)

type temporary interface {
	Temporary() bool
}

func (as *AnyServer[C]) Serve(l Listener[C]) error {
	if as.Handler == nil {
		return errors.New("handler is nil")
	}
	if !as.status.CompareAndSwap(statusInit, statusRunning) {
		s := as.status.Load()
		return fmt.Errorf("invalid status (%s) for Serve", statusTxt(s))
	}
	ctx, cancel := context.WithCancel(context.Background())
	as.closeCancel = cancel
	as.serverExit = make(chan bool, 1)

	var errResult error
	var wg sync.WaitGroup

	loopAccept := func() error {
		if as.status.Load() != statusRunning {
			return ErrShutdown
		}

		if as.BeforeAccept != nil {
			if err := as.BeforeAccept(l); err != nil {
				return err
			}
		}

		conn, err := l.Accept()
		ctxConn := ctx
		if as.OnConn != nil {
			ctxConn, conn, err = as.OnConn(ctxConn, conn, err)
		}

		if err != nil {
			var ne temporary
			if errors.As(err, &ne) && ne.Temporary() {
				return nil
			}

			if strings.Contains(err.Error(), "i/o timeout") {
				return nil
			}
			return err
		}
		wg.Go(func() {
			as.handleConn(ctxConn, conn)
		})
		return nil
	}

	for {
		if errResult = loopAccept(); errResult != nil {
			break
		}
	}

	wg.Wait()
	as.serverExit <- true
	as.status.Store(statusClosed)
	return errResult
}

func (as *AnyServer[C]) handleConn(ctx context.Context, conn C) {
	as.connections.Store(conn, struct{}{})
	defer as.connections.Delete(conn)

	ctx = ContextWithConn(ctx, conn)
	as.Handler.Handle(ctx, conn)
}

func (as *AnyServer[C]) closeAllConn() {
	as.connections.Range(func(c any, _ any) bool {
		if cc, ok := c.(io.Closer); ok {
			_ = cc.Close()
		}
		as.connections.Delete(c)
		return true
	})
}

func (as *AnyServer[C]) Shutdown(ctx context.Context) error {
	switch as.status.Load() {
	case statusClosed,
		statusInit:
		return nil
	}
	if !as.status.CompareAndSwap(statusRunning, statusClosed) {
		s := as.status.Load()
		return fmt.Errorf("invalid status (%s) for Shutdown", statusTxt(s))
	}
	select {
	case <-ctx.Done():
		as.closeAllConn()
	case <-as.serverExit:
	}
	as.closeCancel()
	return nil
}
