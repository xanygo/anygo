//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-15

package xproxy

import (
	"context"

	"github.com/xanygo/anygo/xnet"
)

var _ Driver = (*direct)(nil)

type direct struct {
}

func (d direct) Protocol() string {
	return "Direct"
}

func (d direct) Proxy(ctx context.Context, proxyConn *xnet.ConnNode, c *Config, target string) (*xnet.ConnNode, error) {
	return proxyConn, nil
}

func init() {
	Register(&direct{})
}
