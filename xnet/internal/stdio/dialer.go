//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-01

package stdio

import (
	"context"
	"io"
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xmeta"
	"github.com/xanygo/anygo/ds/xpool"
	"github.com/xanygo/anygo/ds/xpool/xcmdpool"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/internal/ztypes"
	"github.com/xanygo/anygo/xnet/internal"
)

var registry = xmap.Sync[string, *xcmdpool.Command]{}

// Dialer 将命令行工具的 stdin 和 stdout 封装为 Conn 拨号逻辑
type Dialer struct{}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if xp, ok := registry.Load(address); ok {
		rw, err := xp.Spawn(ctx)
		if err != nil {
			return nil, err
		}
		return &stdioConn{
			rw:     rw,
			local:  internal.NewAddr(network, xp.Path),
			remote: internal.NewAddr(network, xp.Path),
		}, nil
	}
	sc := &ztypes.ServiceCommand{}
	if err := sc.LoadFromStr(address); err != nil {
		return nil, err
	}

	poolOption := xpool.OptionFromContext(ctx)
	pc := &xcmdpool.Command{
		Path:       sc.Path,
		Args:       sc.Args,
		PoolOption: poolOption,
		Setup: func(cmd *exec.Cmd) {
			if sc.Dir != "" {
				cmd.Dir = sc.Dir
			}
		},
	}
	xp, _ := registry.LoadOrStore(address, pc)
	rw, err := xp.Spawn(ctx)
	if err != nil {
		return nil, err
	}
	return &stdioConn{
		rw:     rw,
		local:  internal.NewAddr(network, xp.Path),
		remote: internal.NewAddr(network, xp.Path),
	}, nil
}

var _ net.Conn = (*stdioConn)(nil)
var _ xpool.Recycler = (*stdioConn)(nil)

type stdioConn struct {
	rw        io.ReadWriteCloser
	lastErr   xsync.Value[error]
	meta      sync.Map
	onRecycle xsync.OnceLoadValue[func()]
	remote    net.Addr
	local     net.Addr
}

var _ xmeta.Setter = (*stdioConn)(nil)

func (s *stdioConn) SetMeta(key any, val any) {
	s.meta.Store(key, val)
}

var _ xmeta.Getter = (*stdioConn)(nil)

func (s *stdioConn) GetMeta(key any) (any, bool) {
	return s.meta.Load(key)
}

func (s *stdioConn) OnRecycle(fn func()) {
	s.onRecycle.Store(sync.OnceFunc(func() {
		fn() // 这里实际是 entry.Release(lastErr)
		s.lastErr.Clear()
	}))
}

func (s *stdioConn) Err() error {
	return s.lastErr.Load()
}

func (s *stdioConn) Read(b []byte) (n int, err error) {
	n, err = s.rw.Read(b)
	if err != nil {
		s.lastErr.Store(err)
	}
	return n, err
}

func (s *stdioConn) Write(b []byte) (n int, err error) {
	n, err = s.rw.Write(b)
	if err != nil {
		s.lastErr.Store(err)
	}
	return n, err
}

func (s *stdioConn) Close() error {
	// 回收到对象池的逻辑，这一部分只会运行一次
	// 若连接有异常或者不需要了，对象池会负责关闭（再次调用 Close()）
	if recycle := s.onRecycle.Load(); recycle != nil {
		recycle()
		return nil
	}
	return s.rw.Close()
}

func (s *stdioConn) LocalAddr() net.Addr {
	return s.local
}

func (s *stdioConn) RemoteAddr() net.Addr {
	return s.remote
}

func (s *stdioConn) SetDeadline(t time.Time) error {
	return nil
}

func (s *stdioConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (s *stdioConn) SetWriteDeadline(t time.Time) error {
	return nil
}
