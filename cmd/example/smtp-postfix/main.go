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

var to = flag.String("to", "work@localhost", "send to address")
var from = flag.String("from", "work@localhost", "send from address")
var subject = flag.String("subject", "Hello 你好，世界", "subject")
var files = flag.String("files", "", "attachment files")
var inline = flag.Bool("inline", false, "use inline image")
var hardCoded = flag.Bool("hc", false, "hard-coded smtp server info")
var num = flag.Int("n", 1, "")

// 查看邮件：https://www.emlreader.com/

func main() {
	flag.Parse()

	internal.ServiceInit()

	m := &xsmtp.Mail{
		To:      []string{*to},
		From:    *from,
		Subject: *subject,
		Content: strings.Repeat("hello world，你好 <p style='color:red'>红色</p>\n", 2),
	}
	for f := range strings.SplitSeq(*files, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		anygo.Must(m.AddAttachFile(f))
	}

	if *inline {
		m.Content = `你好 <p style='color:red'>红色</p> <img src="cid:404img"">`
		anygo.Must(m.AddInlineFile("../asset/1.jpg", "404img"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var err error

	mailIter := func(yield func(*xsmtp.Mail) bool) {
		for i := 0; i < *num; i++ {
			if !yield(m) {
				return
			}
		}
	}

	// 两种使用方式
	if *hardCoded {
		// 第一种，使用代码配置 smtp 服务器的信息
		cfg := &xsmtp.Config{
			Host:       "127.0.0.1",
			Port:       25,
			NoStartTLS: true,
			Username:   *from,
		}
		err = cfg.SendSeq(ctx, mailIter)
	} else {
		// 第二种，使用配置文件配置 smtp 服务器的信息
		// 配置在 ../service/postfix.json
		err = xsmtp.SendSeq(ctx, "postfix", mailIter)
	}

	log.Println("err=", err)
}
