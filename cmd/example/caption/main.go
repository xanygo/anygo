//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-18

package main

import (
	"net/http"

	"github.com/xanygo/anygo/ximage/caption"
)

func c1(w http.ResponseWriter, r *http.Request) {
	cp := caption.NewRandom(4)
	cp.ServeHTTP(w, r)
}

func c2(w http.ResponseWriter, r *http.Request) {
	cp := caption.NewRandomDigits(4)
	cp.ServeHTTP(w, r)
}

func c3(w http.ResponseWriter, r *http.Request) {
	cp := caption.NewArithmetic()
	cp.ServeHTTP(w, r)
}

func c4(w http.ResponseWriter, r *http.Request) {
	cp := caption.NewArithmetic()
	cp.SetSize(50, 20)
	cp.ServeHTTP(w, r)
}

func index(w http.ResponseWriter, r *http.Request) {
	code := `<html>
<head>
<title>anygo Caption</title>
</head>
<body>
<p><img src='/c1'></p>
<p><img src='/c2'></p>
<p><img src='/c3'></p>
<p><img src='/c4'></p>
</body>
</html>`
	w.Write([]byte(code))
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/c1", c1)
	http.HandleFunc("/c2", c2)
	http.HandleFunc("/c3", c3)
	http.HandleFunc("/c4", c4)

	http.ListenAndServe("127.0.0.1:8080", nil)
}
