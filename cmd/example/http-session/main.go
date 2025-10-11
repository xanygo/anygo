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
	"github.com/xanygo/anygo/store/xsession"
	"github.com/xanygo/anygo/xhttp"
)

func main() {
	router := xhttp.NewRouter()
	sh := &xsession.HTTPHandler{
		NewStorage: func(writer http.ResponseWriter, request *http.Request) xsession.Storage {
			return &xsession.CookieStore{
				Writer:  writer,
				Request: request,
			}
		},
	}
	router.Use(sh.Next)

	router.GetFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		ss := xsession.FromContext(r.Context())
		ss.Set(r.Context(), "k1", "v1:"+r.URL.Query().Get("k1"))
		err := ss.Save(r.Context())
		_, _ = fmt.Fprintf(w, "save=%v", err)
	})

	router.GetFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		ss := xsession.FromContext(r.Context())
		_, _ = fmt.Fprintf(w, "sessionID=%q\n", xsession.IDFromContext(r.Context()))
		vs, err := ss.Get(r.Context(), "k1")
		_, _ = fmt.Fprintf(w, "k1=%v err=%v\n", vs, err)
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
