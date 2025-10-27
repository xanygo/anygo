//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-27

package main

import (
	"context"
	"flag"
	"log"
	"net/http/httputil"

	"github.com/xanygo/anygo/xhttp/xhttpc"
)

var url = flag.String("url", "https://ifconfig.me/ip", "fetch url")

func main() {
	flag.Parse()

	log.Println("Fetch:", *url)

	client := &xhttpc.Client{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := client.Get(ctx, *url)
	if err != nil {
		log.Fatalln(err)
	}
	bf, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(string(bf))
}
