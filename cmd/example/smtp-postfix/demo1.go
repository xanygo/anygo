//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-22

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

// 查看邮件：https://www.emlreader.com/

func main() {
	var to = flag.String("to", "work@localhost", "send to address")
	var from = flag.String("from", "work@localhost", "send from address")
	var subject = flag.String("subject", "Hello 你好，世界", "subject")
	var files = flag.String("files", "", "attachment files")
	var inline = flag.Bool("inline", false, "use inline image")
	flag.Parse()

	internal.ServiceInit()

	req := &xsmtp.Mail{
		To:      []string{*to},
		From:    *from,
		Subject: *subject,
		Content: strings.Repeat("hello world，你好 <p style='color:red'>红色</p>\n", 2),
	}
	for _, f := range strings.Split(*files, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		anygo.Must(req.AddAttachFile(f))
	}

	if *inline {
		req.Content = `你好 <p style='color:red'>红色</p> <img src="cid:404img"">`
		anygo.Must(req.AddInlineFile("../asset/1.jpg", "404img"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := xsmtp.Send(ctx, "postfix", req)
	log.Println("err=", err)
}
