//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-08

package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xhttp"
	"github.com/xanygo/anygo/xhttp/xhandler"
)

func main() {
	router := xhttp.NewRouter()

	router.GetFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/etag/index", http.StatusFound)
	})

	r1 := router.Prefix("/etag", (&xhandler.ETag{}).Next)
	r1.GetFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world"))
	})

	hc := &xhandler.Cache{
		Store: xcache.NewLRU[string, string](1024),
		Key: func(w http.ResponseWriter, r *http.Request) (string, time.Duration) {
			return r.RequestURI, time.Minute
		},
	}
	r2 := router.Prefix("/cache", hc.Next)
	r2.GetFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world"))
		log.Println("visit cache/index")
	})

	ser := &http.Server{
		Handler: router,
	}
	l, err := net.Listen("tcp4", ":8080")
	anygo.Must(err)
	log.Println("listen:", l.Addr().String())
	err = ser.Serve(l)
	log.Println("exit:", err)
}
