//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-12

package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/xhttp"
	"github.com/xanygo/anygo/xhttp/xhandler"
	"github.com/xanygo/anygo/xlog"
)

var listen = flag.String("l", "127.0.0.1:8080", "listen address")

func main() {
	flag.Parse()

	router := xhttp.NewRouter()
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("before", r.URL.String())
			handler.ServeHTTP(w, r)
			log.Println("after", r.URL.String())
		})
	})
	aw := &xhandler.AccessLog{
		Logger: xlog.Default(),
	}
	router.Use(aw.Next)
	router.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello " + r.RequestURI))
	}))

	router.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("demo")
	})

	router.Get("/{name}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello " + r.RequestURI + ", " + r.PathValue("name")))
	}))

	router.Get("/{id}/{name}.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello .html " + r.RequestURI + ",id=" + r.PathValue("id") + ", name=" + r.PathValue("name")))
	}))

	g := router.Prefix("/index/")

	router.Get("/{id}/{name}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello " + r.RequestURI + ",id=" + r.PathValue("id") + ", name=" + r.PathValue("name")))
	}))

	g.GetFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("index.list " + r.RequestURI))
	})
	g.GetFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("index.id " + r.RequestURI + ", id=" + r.PathValue("id")))
	})

	ser := &http.Server{
		Handler: router,
	}
	log.Println("Starting server on", *listen)

	l, err := net.Listen("tcp4", *listen)
	anygo.Must(err)
	log.Println("listen:", l.Addr().String())
	err = ser.Serve(l)
	log.Println("Server exit：", err)
}
