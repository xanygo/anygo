//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xnet

import (
	"context"
	"crypto/tls"
	"net"
	"slices"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/internal/zslice"
	"github.com/xanygo/anygo/xerror"
)

// NewConn  对 net.Conn 封装，以支持 ConnInterceptor
func NewConn(c net.Conn, its ...*ConnInterceptor) net.Conn {
	if rc, ok := c.(*Conn); ok {
		nc := &Conn{
			raw:    rc.raw,
			allIts: zslice.Merge(rc.allIts, its),
		}
		return nc
	}
	nc := &Conn{
		raw:    c,
		allIts: zslice.Merge(globalConnIts, its),
	}
	return nc
}

var globalConnIts connInterceptors

// 在 interceptor.go 里统一用 WithInterceptor 注册
func registerConnInterceptor(its ...*ConnInterceptor) {
	globalConnIts = append(globalConnIts, its...)
}

// NewContextConn 取出 ctx 里的 ConnInterceptor 作为参数， 并对 Conn 包装
func NewContextConn(ctx context.Context, conn net.Conn) net.Conn {
	its := ITsFromContext[*ConnInterceptor](ctx)
	return NewConn(conn, its...)
}

var _ net.Conn = (*Conn)(nil)

// Conn 支持拦截器 ( ConnInterceptor ) 的网络连接
type Conn struct {
	raw net.Conn

	// 全局和创建时传入的拦截器
	allIts connInterceptors
}

var _ ConnUnwrapper = (*Conn)(nil)

func (c *Conn) NetConn() net.Conn {
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
	return c.raw.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.raw.RemoteAddr()
}

func (c *Conn) SetDeadline(dl time.Time) (err error) {
	for i := 0; i < len(c.allIts); i++ {
		if it := c.allIts[i]; it.SetDeadline != nil {
			it.SetDeadline(c.raw, dl)
		}
	}
	err = c.raw.SetDeadline(dl)
	for i := 0; i < len(c.allIts); i++ {
		if it := c.allIts[i]; it.AfterSetDeadline != nil {
			it.AfterSetDeadline(c.raw, dl, err)
		}
	}
	return err
}

func (c *Conn) SetReadDeadline(dl time.Time) (err error) {
	for i := 0; i < len(c.allIts); i++ {
		if it := c.allIts[i]; it.SetReadDeadline != nil {
			it.SetReadDeadline(c.raw, dl)
		}
	}
	err = c.raw.SetReadDeadline(dl)
	for i := 0; i < len(c.allIts); i++ {
		if it := c.allIts[i]; it.AfterSetReadDeadline != nil {
			it.AfterSetReadDeadline(c.raw, dl, err)
		}
	}
	return err
}

func (c *Conn) SetWriteDeadline(dl time.Time) (err error) {
	for i := 0; i < len(c.allIts); i++ {
		if it := c.allIts[i]; it.SetWriteDeadline != nil {
			it.SetWriteDeadline(c.raw, dl)
		}
	}
	err = c.raw.SetWriteDeadline(dl)
	for i := 0; i < len(c.allIts); i++ {
		if it := c.allIts[i]; it.AfterSetWriteDeadline != nil {
			it.AfterSetWriteDeadline(c.raw, dl, err)
		}
	}
	return err
}

var _ Interceptor = (*ConnInterceptor)(nil)

// ConnInterceptor  net.Conn 的拦截器定义
type ConnInterceptor struct {
	Read      func(info ConnInfo, b []byte, invoker func([]byte) (int, error)) (int, error)
	AfterRead func(info ConnInfo, b []byte, readSize int, err error)

	Write      func(info ConnInfo, b []byte, invoker func([]byte) (int, error)) (int, error)
	AfterWrite func(info ConnInfo, b []byte, wroteSize int, err error)

	Close      func(info ConnInfo, invoker func() error) error
	AfterClose func(info ConnInfo, err error)

	SetDeadline      func(info ConnInfo, tm time.Time)
	AfterSetDeadline func(info ConnInfo, tm time.Time, err error)

	SetReadDeadline      func(info ConnInfo, tm time.Time)
	AfterSetReadDeadline func(info ConnInfo, tm time.Time, err error)

	SetWriteDeadline      func(info ConnInfo, tm time.Time)
	AfterSetWriteDeadline func(info ConnInfo, tm time.Time, err error)
}

func (it *ConnInterceptor) Interceptor() {}

type ConnInfo interface {
	// LocalAddr 本地网络地址
	LocalAddr() net.Addr

	// RemoteAddr 远端网络地址
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

var _ net.Conn = (*ConnNode)(nil)

type ConnNode struct {
	Conn         net.Conn   // 最原始的网络连接
	Wraps        []net.Conn // 包括 tls.Conn、被代理逻辑封装后的 conn 等，
	Addr         AddrNode   // 创建链接的的地址信息
	SessionReply any        // 业务握手后得到的信息
	CreatTime    time.Time  // 创建时间
	LongPool     bool       // 是否来自长连接连接池

	// OnClose 调用 Close 的时候调用
	OnClose func() error

	readTotalBytes  atomic.Int64
	firstErr        xerror.OnceSet
	writeTotalBytes atomic.Int64
	readTotalTime   xsync.TimeDuration
	writeTotalTime  xsync.TimeDuration
	usage           atomic.Int64 // 使用次数
}

var _ ConnUnwrapper = (*ConnNode)(nil)

func (t *ConnNode) NetConn() net.Conn {
	if t == nil {
		return nil
	}
	return t.Conn
}

func (t *ConnNode) AddWrap(w net.Conn) {
	t.Wraps = append(t.Wraps, w)
}

func (t *ConnNode) Outer() net.Conn {
	if t == nil {
		return nil
	}
	if len(t.Wraps) == 0 {
		return t.Conn
	}
	return t.Wraps[len(t.Wraps)-1]
}

func (t *ConnNode) Read(b []byte) (n int, err error) {
	start := time.Now()
	n, err = t.Outer().Read(b)
	t.readTotalTime.Add(time.Since(start))
	if err != nil {
		t.firstErr.SetOnce(err)
	}

	t.readTotalBytes.Add(int64(n))
	return n, err
}

func (t *ConnNode) Write(b []byte) (n int, err error) {
	start := time.Now()
	n, err = t.Outer().Write(b)
	t.writeTotalTime.Add(time.Since(start))
	if err != nil {
		t.firstErr.SetOnce(err)
	}

	t.writeTotalBytes.Add(int64(n))
	return n, err
}

func (t *ConnNode) Close() error {
	if t.OnClose != nil {
		return t.OnClose()
	}
	return t.Outer().Close()
}

func (t *ConnNode) LocalAddr() net.Addr {
	return t.Outer().LocalAddr()
}

func (t *ConnNode) RemoteAddr() net.Addr {
	return t.Outer().RemoteAddr()
}

func (t *ConnNode) SetDeadline(tm time.Time) error {
	err := t.Outer().SetDeadline(tm)
	if err != nil {
		t.firstErr.SetOnce(err)
	}
	return err
}

func (t *ConnNode) SetReadDeadline(tm time.Time) error {
	err := t.Outer().SetReadDeadline(tm)
	if err != nil {
		t.firstErr.SetOnce(err)
	}
	return err
}

func (t *ConnNode) SetWriteDeadline(tm time.Time) error {
	err := t.Outer().SetWriteDeadline(tm)
	if err != nil {
		t.firstErr.SetOnce(err)
	}
	return err
}

func (t *ConnNode) ReadBytes() int64 {
	return t.readTotalBytes.Load()
}

func (t *ConnNode) WriteBytes() int64 {
	return t.writeTotalBytes.Load()
}

func (t *ConnNode) ReadCost() time.Duration {
	return t.readTotalTime.Load()
}

func (t *ConnNode) WriteCost() time.Duration {
	return t.writeTotalTime.Load()
}

// UsageCount 被复用的，使用次数
func (t *ConnNode) UsageCount() int64 {
	return t.usage.Load() + 1
}

// Err 获取其首次 error 信息
func (t *ConnNode) Err() error {
	return t.firstErr.Unwrap()
}

func (t *ConnNode) ResetStats() {
	t.usage.Add(1)
	t.readTotalBytes.Store(0)
	t.writeTotalTime.Store(0)
	t.readTotalTime.Store(0)
	t.writeTotalTime.Store(0)
}

func (t *ConnNode) Clone() *ConnNode {
	if t == nil {
		return nil
	}
	return &ConnNode{
		Conn:  t.Conn,
		Wraps: slices.Clone(t.Wraps),
		Addr:  t.Addr,
	}
}

// ConnUnwrapper 返回底层的 Conn
type ConnUnwrapper interface {
	NetConn() net.Conn
}

var _ ConnUnwrapper = (*tls.Conn)(nil)

func UnwrapConn(conn net.Conn) net.Conn {
	for {
		uc, ok := conn.(ConnUnwrapper)
		if !ok {
			return conn
		}
		conn = uc.NetConn()
	}
}
