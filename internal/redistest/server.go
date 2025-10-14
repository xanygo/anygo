//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-13

package redistest

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

const (
	serverCmd   = "redis-server"
	testURIName = "anygo-ut-redis"
	utRdsCmd    = "anygo-redis-do"
)

func NewServer() (*Server, error) {
	srv := &Server{}
	return srv, srv.Start()
}

type Server struct {
	rootCtx context.Context
	stop    context.CancelFunc
	addr    net.Addr
}

func (srv *Server) Addr() net.Addr {
	return srv.addr
}

func (srv *Server) URI() string {
	if oe := srv.envURI(); oe != "" {
		return oe
	}
	return "redis://" + srv.addr.String()
}

func (srv *Server) envURI() string {
	return os.Getenv(testURIName)
}

func (srv *Server) Start() error {
	if oe := srv.envURI(); oe != "" {
		return srv.startByURI()
	}
	return srv.startCmd()
}

func (srv *Server) startByURI() error {
	uri := srv.envURI()
	if uri == "" {
		return errors.New("no redis server URI")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if _, err := exec.LookPath(utRdsCmd); err != nil {
		installCmd := exec.CommandContext(ctx, "go", "install", "github.com/xanygo/anygo/cmd/anygo-redis-do")
		bf, err := installCmd.CombinedOutput()
		log.Println("exec:", installCmd.String(), "\noutput:", string(bf), "\nerr:", err)
		if err != nil {
			return err
		}
	}
	cmd := exec.CommandContext(ctx, utRdsCmd, "-uri", uri, "-c", "FLUSHALL sync;dbsize")
	bf, err := cmd.CombinedOutput()
	log.Println("exec:", cmd.String(), "\noutput:", string(bf), "\nerr:", err)
	return err
}

func (srv *Server) startCmd() error {
	_, err := exec.LookPath(serverCmd)
	if err != nil {
		return err
	}

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
	if srv.stop != nil {
		srv.stop()
	}
}
