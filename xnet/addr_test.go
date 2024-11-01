//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-01

package xnet

import (
	"context"
	"testing"

	"github.com/fsgo/fst"
)

func TestNewAddr(t *testing.T) {
	addr := NewAddr(NetworkTCP, "127.0.0.1:8080")
	b := NewAddr(NetworkTCP, "127.0.0.1:8080")
	fst.True(t, addr.Equal(b))
	fst.Equal(t, "tcp", addr.Network())
	fst.Equal(t, "127.0.0.1:8080", addr.String())

	fst.NotNil(t, addr.Attr())
	addr.Attr().Set("idc", "test")
	fst.Equal(t, "test", addr.Attr().GetFirst("idc"))

	ctx := ContextWithAddr(context.Background(), addr)
	g1 := AddrFromContext(ctx)
	fst.True(t, addr.Equal(g1))
	fst.Nil(t, AddrFromContext(context.Background()))
}

func TestContextWithAddr(t *testing.T) {
	addr := NewAddr(NetworkTCP, "127.0.0.1:8080")
	ctx := ContextWithAddr(context.Background(), addr)
	g1 := AddrFromContext(ctx)
	fst.True(t, addr.Equal(g1))
	fst.Nil(t, AddrFromContext(context.Background()))
}
