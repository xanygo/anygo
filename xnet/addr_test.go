//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-01

package xnet

import (
	"context"
	"net"
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

func TestIP4ToLong(t *testing.T) {
	tests := []struct {
		ip   string
		want uint32
	}{
		{
			ip:   "192.0.34.166",
			want: 3221234342,
		},
		{
			ip:   "127.0.0.1",
			want: 2130706433,
		},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			got := IP4ToLong(ip)
			if got != tt.want {
				t.Errorf("IP4ToLong() = %v, want %v", got, tt.want)
			}
			long := LongToIP4(got)
			w2 := long.String()
			if w2 != tt.ip {
				t.Errorf("LongToIP4() = %v, want %v", w2, tt.ip)
			}
		})
	}
}
