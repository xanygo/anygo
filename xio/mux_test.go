//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-08

package xio_test

import (
	"fmt"
	"net"
	"os"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/xanygo/anygo/xio"
	"github.com/xanygo/anygo/xt"
)

func TestNewMux(t *testing.T) {
	time.AfterFunc(10*time.Second, func() {
		f, _ := os.Create("goroutine.prof")
		pprof.Lookup("goroutine").WriteTo(f, 2) // 2 = 包含 stack
		f.Close()
	})

	client, server := net.Pipe()

	muxClient := xio.NewMux(true, client)
	muxServer := xio.NewMux(false, server)

	handleServerStream := func(s *xio.MuxStream[net.Conn]) {
		t.Logf("handleServerStream sid=%d", s.ID())
		defer s.Close()
		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil {
			t.Logf("handleServerStream read error: %v", err)
			return
		}
		got := string(buf[:n])
		want := fmt.Sprintf("hello %d", s.ID())
		xt.Equal(t, want, got)
		t.Logf("stream=%d Read %q", s.ID(), got)

		resp := "echo: " + got
		n, err = s.Write([]byte(resp))
		xt.NoError(t, err)
		xt.Equal(t, len(resp), n)
	}

	var wg1 sync.WaitGroup
	// server side: accept remote streams
	wg1.Go(func() {
		var wg3 sync.WaitGroup
		for range 2 {
			s, err := muxServer.Accept()
			if err != nil {
				t.Logf("muxServer.Accept, err=%v", err)
				return
			}
			t.Logf("muxServer.Accept sid=%d", s.ID())
			wg3.Go(func() {
				handleServerStream(s)
			})
		}
		wg3.Wait()
	})

	// client side: open two streams
	clientStream1, _ := muxClient.Open()
	xt.Equal(t, 2, clientStream1.ID())

	clientStream2, _ := muxClient.Open()
	xt.Equal(t, 4, clientStream2.ID())

	var wg2 sync.WaitGroup
	wg2.Go(func() {
		str := fmt.Sprintf("hello %d", clientStream1.ID())
		n, err := clientStream1.Write([]byte(str))
		xt.NoError(t, err)
		xt.Equal(t, len(str), n)
	})

	wg2.Go(func() {
		str := fmt.Sprintf("hello %d", clientStream2.ID())
		n, err := clientStream2.Write([]byte(str))
		xt.NoError(t, err)
		xt.Equal(t, len(str), n)

		// read echo
		buf := make([]byte, 1024)
		n, err = clientStream2.Read(buf)
		xt.NoError(t, err)
		got := string(buf[:n])
		want := "echo: " + str
		xt.Equal(t, want, got)
	})

	wg2.Wait()
	wg1.Wait()

	t.Log("call muxClient.Close()")
	xt.NoError(t, muxClient.Close())

	t.Log("call muxServer.Close()")
	xt.NoError(t, muxServer.Close())
	clientStream1.Close()
	clientStream2.Close()

	client.Close()
	server.Close()
}
