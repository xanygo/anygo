// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/16

package xrps_test

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/xanygo/anygo/xnet/xrps"
	"github.com/xanygo/anygo/xt"
)

func TestAnyServer(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	xt.NoError(t, err)
	xt.NotNil(t, l)
	defer l.Close()

	ser := &xrps.AnyServer{
		Handler: xrps.HandleFunc(echoHandler),
	}
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ser.Serve(l)
	})
	conn, err := net.DialTimeout("tcp", l.Addr().String(), 100*time.Millisecond)
	xt.NoError(t, err)
	rd := bufio.NewReader(conn)
	for i := range 10 {
		t.Run(fmt.Sprintf("loop=%d", i), func(t *testing.T) {
			_, err = conn.Write([]byte("hello\n"))
			xt.NoError(t, err)
			line, _, err := rd.ReadLine()
			xt.NoError(t, err)
			xt.Equal(t, `resp:"hello"`, string(line))
		})
	}
	xt.NoError(t, conn.Close())
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	xt.NoError(t, ser.Shutdown(ctx))
	xt.NoError(t, l.Close())
	wg.Wait()
}

func echoHandler(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	rd := bufio.NewReader(conn)
	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			return
		}
		resp := fmt.Sprintf("resp:%q\n", line)
		_, err = conn.Write([]byte(resp))
		if err != nil {
			return
		}
	}
}
