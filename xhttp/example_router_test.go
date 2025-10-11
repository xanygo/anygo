//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-14

package xhttp_test

import (
	"fmt"
	"net/http"

	"github.com/xanygo/anygo/xhttp"
)

func ExampleRouter_Prefix() {
	r := xhttp.NewRouter()
	r.GetFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("index"))
	})

	g1 := r.Prefix("/user/")

	// 完整地址： /user/{id}/detail,  可接收如 GET /user/123/detail
	g1.GetFunc("/{id}/detail", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("user detail"))
	})

	// 完整地址：/user/{id}, 可接收如 DELETE /user/123
	g1.DeleteFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("delete user"))
	})
}

func ExampleRouter_Use() {
	r := xhttp.NewRouter()
	r.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("call before")

			handler.ServeHTTP(w, r)

			fmt.Println("call after")
		})
	})
}
