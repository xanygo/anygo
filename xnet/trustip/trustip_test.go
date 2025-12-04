//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-04

package trustip

import (
	"net"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestList(t *testing.T) {
	xt.NotEmpty(t, List())
	xt.True(t, IsTrusted(net.ParseIP("127.0.0.1")))

	xt.False(t, IsTrusted(net.ParseIP("8.8.8.8")))

	xt.NoError(t, Set(nil))
	xt.False(t, IsTrusted(net.ParseIP("127.0.0.1")))
}
