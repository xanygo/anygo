//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-06-01

package main

import (
	"context"
	_ "embed"
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/xhttp"
	"github.com/xanygo/anygo/xhttp/xws"
)

//go:embed index.html
var indexCode []byte

var listen = flag.String("l", "127.0.0.1:8080", "listen address")

var upgrader = websocket.Upgrader{}

func main() {
	flag.Parse()

	router := xhttp.NewRouter()
	router.GetFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(indexCode)
	})
	router.GetFunc("/ws", wsHandler)

	ser := &http.Server{
		Handler: router,
	}

	log.Println("Starting server on:", *listen)
	l, err := net.Listen("tcp4", *listen)
	anygo.Must(err)
	err = ser.Serve(l)
	log.Println("Server exit:", err)
}

var hub = &xws.Hub{}
var wsRouter = xws.NewRouter()

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err == nil {
		err = hub.ServeWSUpgradeRaw(r.Context(), wsRouter, r, c)
	}
	log.Println("wsHandler, err=", err)
}

func init() {
	hub.OnConnect(func(ctx context.Context, conn xws.Conn) error {
		log.Println("OnConnect:", conn.ID(), conn.HTTPRequest().RemoteAddr)
		return nil
	})
	hub.OnDisconnect(func(ctx context.Context, conn xws.Conn) {
		log.Println("OnDisconnect:", conn.ID(), conn.HTTPRequest().RemoteAddr)
	})
	wsRouter.HandleTextFunc("echo", func(ctx context.Context, w xws.ResponseWriter, r *xws.Request) {
		w.Write(ctx, r.Message)
		log.Println("call echo, msg.Payload=", string(r.Message.Payload))
	})
	wsRouter.HandleNotFoundFunc(func(ctx context.Context, w xws.ResponseWriter, r *xws.Request) {
		log.Println("not found, messageType=", r.Message.Type, "payload=", r.Message.Payload)
	})

	wsRouter.HandleBinaryFunc(func(ctx context.Context, w xws.ResponseWriter, r *xws.Request) {
		w.Write(ctx, r.Message)
		log.Println("call HandleBinary")
	})
}
