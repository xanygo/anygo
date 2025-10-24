//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-02

package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/cmd/example/internal"
	"github.com/xanygo/anygo/xnet/xsmtp"
)

var files = flag.String("files", "", "attachment files")
var inline = flag.Bool("inline", false, "use inline image")
var to = flag.String("to", "", "receiver address")

func main() {
	flag.Parse()
	internal.ServiceInit()

	req := &xsmtp.Mail{
		To:      strings.Split(*to, ","),
		Subject: "hello 你好",
		Content: strings.Repeat("hello world，你好 <p style='color:red'>红色</p>\n", 2),
	}
	for _, f := range strings.Split(*files, ",") {
		log.Printf("try add file %q", f)
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		anygo.Must(req.AddAttachFile(f))
	}
	log.Println("files=", len(req.Attachment))
	if *inline {
		req.Content = "你好 <p style='color:red'>红色</p> <img src='cid:404img'>"
		anygo.Must(req.AddInlineFile("../asset/1.jpg", "404img"))
	}
	// ap := xrpc.OptHostPort("127.0.0.1:25")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := xsmtp.Send(ctx, "smtp_163", req)
	log.Println("err=", err)
}
