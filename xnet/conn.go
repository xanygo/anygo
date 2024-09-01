//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xnet

import (
	"context"
	"net"
	"time"

	"github.com/xanygo/anygo/internal/zslice"
)

// NewConn  conn WithIt with Interceptor
// its 将倒序执行：后注册的先执行
func NewConn(c net.Conn, its ...*ConnInterceptor) net.Conn {
	globalConnIts := InterceptorFromGlobal[*ConnInterceptor]()
	if rc, ok := c.(*Conn); ok {
		nc := &Conn{
			raw:  rc.raw,
			args: zslice.Merge(rc.args, its),
		}
		nc.allIts = zslice.Merge(globalConnIts, nc.args)
		return nc
	}

	nc := &Conn{
		raw:    c,
		allIts: zslice.Merge(globalConnIts, its),
		args:   its,
	}
	return nc
}

// NewConnContext 取出 ctx 里的 Interceptor， 并对 conn 进行封装
func NewConnContext(ctx context.Context, conn net.Conn) net.Conn {
	cks := InterceptorFromContext[*ConnInterceptor](ctx)
	return NewConn(conn, cks...)
}

type Conn struct {
	raw net.Conn

	// 全局和创建时传入的拦截器
	allIts connInterceptors

	// 创建时传入的拦截器
	args connInterceptors
}

func (c *Conn) Unwrap() net.Conn {
	return c.raw
}

func (c *Conn) Read(b []byte) (n int, err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].Read != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		n, err = c.raw.Read(b)
	} else {
		n, err = c.allIts.CallRead(c.raw, b, c.raw.Read, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterRead != nil {
			c.allIts[i].AfterRead(c.raw, b, n, err)
		}
	}
	return n, err
}

func (c *Conn) Write(b []byte) (n int, err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].Write != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		n, err = c.raw.Write(b)
	} else {
		n, err = c.allIts.CallWrite(c.raw, b, c.raw.Write, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterWrite != nil {
			c.allIts[i].AfterWrite(c.raw, b, n, err)
		}
	}
	return n, err
}

func (c *Conn) Close() (err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].Close != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		err = c.raw.Close()
	} else {
		err = c.allIts.CallClose(c.raw, c.raw.Close, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterClose != nil {
			c.allIts[i].AfterClose(c.raw, err)
		}
	}
	return err
}

func (c *Conn) LocalAddr() net.Addr {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].LocalAddr != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		return c.raw.LocalAddr()
	}
	return c.allIts.CallLocalAddr(c.raw, c.raw.LocalAddr, idx)
}

func (c *Conn) RemoteAddr() net.Addr {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].RemoteAddr != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		return c.raw.RemoteAddr()
	}
	return c.allIts.CallRemoteAddr(c.raw, c.raw.RemoteAddr, idx)
}

func (c *Conn) SetDeadline(t time.Time) (err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].SetDeadline != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		err = c.raw.SetDeadline(t)
	} else {
		err = c.allIts.CallSetDeadline(c.raw, t, c.raw.SetDeadline, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterSetDeadline != nil {
			c.allIts[i].AfterSetDeadline(c.raw, t, err)
		}
	}
	return err
}

func (c *Conn) SetReadDeadline(t time.Time) (err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].SetReadDeadline != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		err = c.raw.SetReadDeadline(t)
	} else {
		err = c.allIts.CallSetReadDeadline(c.raw, t, c.raw.SetReadDeadline, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterSetReadDeadline != nil {
			c.allIts[i].AfterSetReadDeadline(c.raw, t, err)
		}
	}
	return err
}

func (c *Conn) SetWriteDeadline(t time.Time) (err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].SetWriteDeadline != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		err = c.raw.SetWriteDeadline(t)
	} else {
		err = c.allIts.CallSetWriteDeadline(c.raw, t, c.raw.SetWriteDeadline, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterSetWriteDeadline != nil {
			c.allIts[i].AfterSetWriteDeadline(c.raw, t, err)
		}
	}
	return err
}

// ConnInterceptor  for net.Conn
type ConnInterceptor struct {
	Read      func(info ConnInfo, b []byte, invoker func([]byte) (int, error)) (int, error)
	AfterRead func(info ConnInfo, b []byte, readSize int, err error)

	Write      func(info ConnInfo, b []byte, invoker func([]byte) (int, error)) (int, error)
	AfterWrite func(info ConnInfo, b []byte, wroteSize int, err error)

	Close      func(info ConnInfo, invoker func() error) error
	AfterClose func(info ConnInfo, err error)

	LocalAddr  func(info ConnInfo, invoker func() net.Addr) net.Addr
	RemoteAddr func(info ConnInfo, invoker func() net.Addr) net.Addr

	SetDeadline      func(info ConnInfo, tm time.Time, invoker func(tm time.Time) error) error
	AfterSetDeadline func(info ConnInfo, tm time.Time, err error)

	SetReadDeadline      func(info ConnInfo, tm time.Time, invoker func(tm time.Time) error) error
	AfterSetReadDeadline func(info ConnInfo, tm time.Time, err error)

	SetWriteDeadline      func(info ConnInfo, tm time.Time, invoker func(tm time.Time) error) error
	AfterSetWriteDeadline func(info ConnInfo, tm time.Time, err error)
}

func (it *ConnInterceptor) Interceptor() {}

type ConnInfo interface {
	// LocalAddr returns the local network address, if known.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address, if known.
	RemoteAddr() net.Addr
}

// 先注册的先执行
type connInterceptors []*ConnInterceptor

func (chs connInterceptors) CallRead(info ConnInfo, b []byte, invoker func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Read != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(b)
	}

	return chs[idx].Read(info, b, func(b []byte) (int, error) {
		return chs.CallRead(info, b, invoker, idx+1)
	})
}

func (chs connInterceptors) CallWrite(info ConnInfo, b []byte, invoker func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Write != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(b)
	}
	return chs[idx].Write(info, b, func(b []byte) (int, error) {
		return chs.CallWrite(info, b, invoker, idx+1)
	})
}

func (chs connInterceptors) CallClose(info ConnInfo, invoker func() error, idx int) (err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Close != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker()
	}
	return chs[idx].Close(info, func() error {
		return chs.CallClose(info, invoker, idx+1)
	})
}

func (chs connInterceptors) CallLocalAddr(info ConnInfo, invoker func() net.Addr, idx int) net.Addr {
	for ; idx < len(chs); idx++ {
		if chs[idx].LocalAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker()
	}
	return chs[idx].LocalAddr(info, func() net.Addr {
		return chs.CallLocalAddr(info, invoker, idx+1)
	})
}

func (chs connInterceptors) CallRemoteAddr(info ConnInfo, invoker func() net.Addr, idx int) net.Addr {
	for ; idx < len(chs); idx++ {
		if chs[idx].RemoteAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker()
	}
	return chs[idx].RemoteAddr(info, func() net.Addr {
		return chs.CallRemoteAddr(info, invoker, idx+1)
	})
}

func (chs connInterceptors) CallSetDeadline(info ConnInfo, dl time.Time, invoker func(time.Time) error, idx int) (err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(dl)
	}
	return chs[idx].SetDeadline(info, dl, func(dl time.Time) error {
		return chs.CallSetDeadline(info, dl, invoker, idx+1)
	})
}

func (chs connInterceptors) CallSetReadDeadline(info ConnInfo, dl time.Time, invoker func(time.Time) error, idx int) (err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetReadDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(dl)
	}
	return chs[idx].SetReadDeadline(info, dl, func(dl time.Time) error {
		return chs.CallSetReadDeadline(info, dl, invoker, idx+1)
	})
}

func (chs connInterceptors) CallSetWriteDeadline(info ConnInfo, dl time.Time, invoker func(time.Time) error, idx int) (err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetWriteDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(dl)
	}
	return chs[idx].SetWriteDeadline(info, dl, func(dl time.Time) error {
		return chs.CallSetWriteDeadline(info, dl, invoker, idx+1)
	})
}
