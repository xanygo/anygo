//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-12

package xerror

import (
	"net"
	"testing"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestIsNetError(t *testing.T) {
	l, err1 := net.Listen("tcp", "127.0.0.1:0")
	xt.NoError(t, err1)
	defer l.Close()
	conn1, err2 := net.DialTimeout("tcp", l.Addr().String(), time.Second)
	xt.NoError(t, err2)
	_ = conn1.Close()

	_, err3 := conn1.Write([]byte("hello"))
	xt.Error(t, err3)
	xt.True(t, IsClientNetError(err3))
}
