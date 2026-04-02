//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xdial

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xmeta"
	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xio"
	"github.com/xanygo/anygo/xnet"
)

// SessionReply 表示 SessionStarter 执行后的结果
type SessionReply interface {
	String() string // 用于打印出完整的内容

	// Summary 返回简短的描述信息
	Summary() string
}

// SessionStarter 用于创建连接后,tcp client 业务层面的握手
// 如 Redis 协议发送 Hello Request
type SessionStarter interface {
	// StartSession 开始一个会话，conn 大多数情况下是 *xnet.ConnNode
	StartSession(ctx context.Context, rw io.ReadWriter, opt xoption.Reader) (SessionReply, error)
}

var _ SessionStarter = StartSessionFunc(nil)

type StartSessionFunc func(ctx context.Context, rw io.ReadWriter, opt xoption.Reader) (SessionReply, error)

func (h StartSessionFunc) StartSession(ctx context.Context, rw io.ReadWriter, opt xoption.Reader) (SessionReply, error) {
	return h(ctx, rw, opt)
}

func WithSessionStarter(c Connector, h SessionStarter, opt xoption.Reader) Connector {
	return ConnectorFunc(func(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (io.ReadWriteCloser, error) {
		conn, err := c.Connect(ctx, addr, opt)
		if err != nil {
			return conn, err
		}
		ret, err := h.StartSession(ctx, conn, opt)
		if err != nil {
			conn.Close()
			return conn, err
		}
		xmeta.TrySetMeta(conn, xmeta.KeySessionReply, ret)
		return conn, nil
	})
}

var protocols = &xmap.Sync[string, SessionStarter]{}

func RegisterSessionStarter(protocol string, h SessionStarter) error {
	if protocol == "" {
		return errors.New("protocol name is empty")
	}
	_, loaded := protocols.LoadOrStore(strings.ToUpper(protocol), h)
	if loaded {
		return fmt.Errorf("protocol %s already registered", protocol)
	}
	return nil
}

// FindSessionStarter 按照协议查找，若找不到会返回nil
func FindSessionStarter(protocol string) SessionStarter {
	handler, _ := protocols.Load(strings.ToUpper(protocol))
	return handler
}

// StartSession 用于在连接创建完成后，业务正式使用前，执行会话开启的逻辑。
// 如身份认证，协议升级等。
// 目前在 xservice.connector 里调用
func StartSession(ctx context.Context, protocol string, rw io.ReadWriter, opt xoption.Reader) (ret SessionReply, started bool, err error) {
	ctx, span := xmetric.Start(ctx, "StartSession")
	timeout := xoption.HandshakeTimeout(opt)
	defer func() {
		span.SetAttributes(
			xmetric.AnyAttr("Protocol", protocol),
			xmetric.AnyAttr("Timeout", timeout),
		)
		span.RecordError(err)
		span.End()
	}()
	handler := SessionStarterFromContext(ctx)
	if handler == nil {
		handler = FindSessionStarter(protocol)
	}
	// 若找不到，则直接返回(跳过)
	if handler == nil {
		return nil, false, nil
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if ds, ok := rw.(xio.DeadlineSetter); ok {
		ds.SetDeadline(time.Now().Add(timeout))
		defer ds.SetDeadline(time.Time{})
	}
	ret, err = handler.StartSession(ctx, rw, opt)

	if err == nil && ret != nil {
		span.SetAttributes(xmetric.AnyAttr("Summary", ret.Summary()))
	}
	return ret, true, err
}

var ctxKey = xctx.NewKey()

func ContextWithSessionStarter(ctx context.Context, h SessionStarter) context.Context {
	return context.WithValue(ctx, ctxKey, h)
}

func SessionStarterFromContext(ctx context.Context) SessionStarter {
	val, _ := ctx.Value(ctxKey).(SessionStarter)
	return val
}

// HTTPUpgrade 构建一个能执行 HTTP Upgrade 逻辑的会话创建逻辑
func HTTPUpgrade(method string, uri string, protocol string) SessionStarter {
	hd := bytes.NewBuffer(nil)
	fmt.Fprintf(hd, "%s %s HTTP/1.1\r\n", method, uri)
	fmt.Fprintf(hd, "Upgrade: %s\r\n", protocol)
	fmt.Fprint(hd, "Connection: Upgrade\r\n")
	return StartSessionFunc(func(ctx context.Context, rw io.ReadWriter, opt xoption.Reader) (SessionReply, error) {
		conn, ok := rw.(*xnet.ConnNode)
		if !ok {
			return nil, errors.New("conn is not a net.ConnNode")
		}
		host := fmt.Sprintf("Host: %s\r\n\r\n", conn.Addr.HostPort)
		bf := bytes.NewBuffer(nil)
		bf.Grow(hd.Len() + len(host))
		bf.Write(hd.Bytes())
		bf.WriteString(host)
		_, err := rw.Write(bf.Bytes())
		if err != nil {
			return nil, err
		}
		reader := bufio.NewReader(conn)
		resp, err := http.ReadResponse(reader, nil)
		if err != nil {
			return nil, err
		}

		// 校验状态码是否为 101 Switching Protocols
		if resp.StatusCode != http.StatusSwitchingProtocols {
			return nil, fmt.Errorf("upgrade failed with status=%s, expect statusCode=101", resp.Status)
		}
		return nil, nil
	})
}
