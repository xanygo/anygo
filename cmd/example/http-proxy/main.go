//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-13

package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/xanygo/anygo/xnet/xproxy"
)

var proxy = flag.String("proxy", "http://172.31.0.1:10000", "proxy server address")
var get = flag.String("get", "https://ifconfig.me/all", "request get")

func main() {
	flag.Parse()
	log.Println("proxy server:", *proxy)
	pd, err := xproxy.NewDialer(*proxy)
	if err != nil {
		log.Fatalln("create proxy dialer:", err)
	}
	c := &http.Client{
		Transport: &http.Transport{
			DialContext: pd.DialContext,
		},
	}
	log.Println("request get:", *get)
	resp, err := c.Get(*get)
	if err != nil {
		log.Fatalln("Get failed", err)
	}
	defer resp.Body.Close()
	bf, _ := httputil.DumpResponse(resp, true)
	log.Println("DumpResponse:\n", string(bf))
}
