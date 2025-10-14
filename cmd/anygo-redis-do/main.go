//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-13

package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xredis"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

var uri = flag.String("uri", "redis://127.0.0.1:6379", "redis URI")
var cmds = flag.String("c", "ping", "commands")

func main() {
	flag.Parse()
	if *uri == "" {
		log.Fatalln("uri flag is required")
	}
	_, client, err := xredis.NewClientByURI("demo", *uri)
	if err != nil {
		log.Fatalln("NewClientByURI:", err)
	}
	lines := strings.Split(*cmds, ";")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		arr := strings.Fields(line)
		log.Printf("command : %q\n", arr)
		req := resp3.NewRequest(resp3.DataTypeAny, xslice.ToAnys(arr)...)
		resp := client.Do(ctx, req)
		result, err := resp.Result()
		if err == nil {
			log.Printf("result  : %#v\n", result)
		} else {
			log.Fatalln("err=", err)
		}
	}
}
