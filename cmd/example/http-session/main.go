//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/store/session"
	"github.com/xanygo/anygo/xhttp"
)

func main() {
	router := xhttp.NewRouter()
	sessionHandler := (&session.CookieStoreHandler{}).Trans()
	router.Use(sessionHandler.Next)

	router.GetFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		ss := sessionHandler.Session(r)
		ss.Set("k1", "v1:"+r.URL.Query().Get("k1"))
		err := ss.Save(r.Context())
		_, _ = fmt.Fprintf(w, "save=%v", err)
	})

	router.GetFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		ss := sessionHandler.Session(r)
		_, _ = fmt.Fprintf(w, "k1=%v", ss.Get("k1"))
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
