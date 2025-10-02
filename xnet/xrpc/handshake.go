//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-02

package xrpc

import (
	"context"

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
