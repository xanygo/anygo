//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-01

package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/xnet/xjsonrpc2"
)

var listen = flag.String("l", "127.0.0.1:8080", "listen address")

func main() {
	flag.Parse()
	log.Println("Starting server on", *listen)
	l, err := net.Listen("tcp4", *listen)
	anygo.Must(err)

	router := xjsonrpc2.NewRouter()
	router.RegisterUnary("ping", func(ctx context.Context, req *xjsonrpc2.Request) (result any, err error) {
		var payload string
		err = req.DecodeParams(&payload)
		if err != nil {
			return nil, err
		}
		return "Ok:" + payload, nil
	})
	ser := &http.Server{
		Handler: router,
	}
	err = ser.Serve(l)
	log.Fatalf("server exited with error: %v", err)
}
