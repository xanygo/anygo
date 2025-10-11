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
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xoption"
)

// HandshakeHandler 用于创建连接后，tcp client 的握手
// 如 Redis 协议发送 Hello Request
type HandshakeHandler interface {
	Handshake(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader) (any, error)
}

var _ HandshakeHandler = HandshakeFunc(nil)

type HandshakeFunc func(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader) (any, error)

func (h HandshakeFunc) Handshake(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader) (any, error) {
	return h(ctx, conn, opt)
}

func WithHandshake(c Connector, h HandshakeHandler, opt xoption.Reader) Connector {
	return ConnectorFunc(func(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error) {
		conn, err := c.Connect(ctx, addr, opt)
		if err != nil {
			return conn, err
		}
		ret, err := h.Handshake(ctx, conn, opt)
		if err != nil {
			conn.Close()
			return conn, err
		}
		conn.Handshake = ret
		return conn, nil
	})
}

var protocols = &xmap.Sync[string, HandshakeHandler]{}

func RegisterHandshakeHandler(protocol string, h HandshakeHandler) error {
	if protocol == "" {
		return errors.New("protocol name is empty")
	}
	_, loaded := protocols.LoadOrStore(strings.ToUpper(protocol), h)
	if loaded {
		return fmt.Errorf("protocol %s already registered", protocol)
	}
	return nil
}

func FindHandshakeHandler(protocol string) (HandshakeHandler, error) {
	handler, ok := protocols.Load(strings.ToUpper(protocol))
	if ok {
		return handler, nil
	}
	return nil, fmt.Errorf("protocol %s not registered", protocol)
}

func Handshake(ctx context.Context, protocol string, conn *xnet.ConnNode, opt xoption.Reader) (ret any, err error) {
	ctx, span := xmetric.Start(ctx, "Handshake")
	timeout := xoption.HandshakeTimeout(opt)
	defer func() {
		span.SetAttributes(
			xmetric.AnyAttr("Protocol", protocols),
			xmetric.AnyAttr("Timeout", timeout),
		)
		span.RecordError(err)
		span.End()
	}()
	handler, err1 := FindHandshakeHandler(protocol)
	if err1 != nil {
		return conn, err1
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cc := conn.NetConn()
	cc.SetDeadline(time.Now().Add(timeout))
	ret, err = handler.Handshake(ctx, conn, opt)
	cc.SetDeadline(time.Time{})
	return ret, err
}
