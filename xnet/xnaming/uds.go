//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-01

package xnaming

import (
	"context"

	"github.com/xanygo/anygo/xnet"
)

var _ Naming = (*Unix)(nil)

type Unix struct{}

func (d *Unix) Scheme() string {
	return xnet.NetworkUnix
}

func (d *Unix) Lookup(ctx context.Context, idc string, sockPath string) ([]xnet.AddrNode, error) {
	return []xnet.AddrNode{
		{
			HostPort: sockPath,
			Addr:     xnet.NewAddr(xnet.NetworkUnix, sockPath),
		},
	}, nil
}

func init() {
	Register(&Unix{})
}
