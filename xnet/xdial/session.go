//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xdial

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xoption"
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
	StartSession(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader) (SessionReply, error)
}

var _ SessionStarter = StartSessionFunc(nil)

type StartSessionFunc func(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader) (SessionReply, error)

func (h StartSessionFunc) StartSession(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader) (SessionReply, error) {
	return h(ctx, conn, opt)
}

func WithSessionStarter(c Connector, h SessionStarter, opt xoption.Reader) Connector {
	return ConnectorFunc(func(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error) {
		conn, err := c.Connect(ctx, addr, opt)
		if err != nil {
			return conn, err
		}
		ret, err := h.StartSession(ctx, conn, opt)
		if err != nil {
			conn.Close()
			return conn, err
		}
		conn.SessionReply = ret
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

func FindSessionStarter(protocol string) (SessionStarter, error) {
	handler, ok := protocols.Load(strings.ToUpper(protocol))
	if ok {
		return handler, nil
	}
	return nil, fmt.Errorf("protocol %s not registered", protocol)
}

func StartSession(ctx context.Context, protocol string, conn *xnet.ConnNode, opt xoption.Reader) (ret SessionReply, err error) {
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
	handler, err1 := FindSessionStarter(protocol)
	if err1 != nil {
		return nil, err1
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn.SetDeadline(time.Now().Add(timeout))
	ret, err = handler.StartSession(ctx, conn, opt)
	conn.SetDeadline(time.Time{})
	if err == nil && ret != nil {
		span.SetAttributes(xmetric.AnyAttr("Summary", ret.Summary()))
	}
	return ret, err
}
