//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-13

package redistest

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"
)

const serverCmd = "redis-server"

func NewServer() (*Server, error) {
	srv := &Server{}
	if !srv.Enable() {
		return nil, fmt.Errorf("no %s available", serverCmd)
	}
	err := srv.Start()
	if err != nil {
		return nil, err
	}
	return srv, nil
}

type Server struct {
	rootCtx context.Context
	stop    context.CancelFunc
	addr    net.Addr
}

func (srv *Server) Enable() bool {
	p, err := exec.LookPath(serverCmd)
	return err == nil && p != ""
}

func (srv *Server) Addr() net.Addr {
	return srv.addr
}

func (srv *Server) URI() string {
	return "redis://" + srv.addr.String()
}

func (srv *Server) Start() error {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	srv.addr = l.Addr()
	addr := l.Addr().String()
	if err = l.Close(); err != nil {
		return err
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	srv.rootCtx, srv.stop = context.WithCancel(context.Background())
	args := []string{
		"--save", "",
		"--appendonly", "no",
		"--maxmemory", "256mb",
		"--maxmemory-policy", "allkeys-lru",
		"--bind", host,
		"--port", port,
	}
	cmd := exec.CommandContext(srv.rootCtx, serverCmd, args...)
	log.Println("exec:", cmd.String())
	done := make(chan bool)
	go func() {
		done <- true
		cmd.Run()
		log.Println("redis-server stopped")
	}()
	<-done
	for i := 0; i < 1000; i++ {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (srv *Server) Stop() {
	srv.stop()
}
