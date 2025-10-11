//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-17

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/internal/cmd/monitor"
	"github.com/xanygo/anygo/xnet"
)

var localAddress = flag.String("l", ":8200", "local server listen address")
var remoteAddress = flag.String("r", "", "remote server address,eg example.com:80")
var printType = flag.String("p", "s", `print types, use comma ',' to separate multiple values (e.g., "b64,np")
s   : string (default)
b   : binary
c   : char
x   : base 16, with lower-case letters for a-f
X   : base 16, with upper-case letters for A-F
U   : unicode format
b64 : with base64 std encoding
q   : quoted character
qn  : quoted string with extra \n retained`)
var outDir = flag.String("o", "", "output directory")
var noRead = flag.Bool("nr", false, "don't print read data to stdout")

func main() {
	flag.Parse()
	if *localAddress == "" {
		log.Fatal("local address is empty")
	}
	if *remoteAddress == "" {
		log.Fatal("remote address is empty")
	}
	listener, err := net.Listen("tcp", *localAddress)
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("Listen on", listener.Addr().String())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}
		go onConnect(conn)
	}
}

var cid atomic.Int64

func onConnect(conn net.Conn) {
	start := time.Now()
	id := cid.Add(1)
	lg := log.New(os.Stderr, fmt.Sprintf("[%d] ", id), log.Ltime)
	lg.Println("onConnect", conn.LocalAddr().String())

	defer conn.Close()

	remote, err := net.DialTimeout("tcp", *remoteAddress, 3*time.Second)
	if err != nil {
		lg.Println("connect remote fail", err.Error())
		return
	}
	lg.Println("connected", conn.LocalAddr().String(), "--->", remote.RemoteAddr().String())
	defer remote.Close()

	mit := &monitor.ConnMonitor{
		ID:        id,
		Logger:    lg,
		PrintType: *printType,
		OutputDir: *outDir,
		NoRead:    *noRead,
		Time:      time.Now(),
	}
	remote = xnet.NewConn(remote, mit.Interceptor())
	done := make(chan error, 2)
	go func() {
		_, err1 := io.Copy(remote, conn)
		done <- err1
		lg.Println("io.Copy1", err1)
	}()
	go func() {
		_, err2 := io.Copy(conn, remote)
		done <- err2
		lg.Println("io.Copy2", err2)
	}()
	<-done
	cost := time.Since(start)
	lg.Println("closed", remote.RemoteAddr().String(), "cost=", cost.String())
}
