//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xlog_test

import (
	"context"
	"os"

	"github.com/xanygo/anygo/xlog"
)

func ExampleInfo() {
	// 使用 Default Logger ( 输出到 stderr ) 打印一条 Info 级别的日志
	xlog.Info(context.Background(), "hello world")
}

func ExampleAddAttr() {
	ctx := xlog.NewContext(context.Background())

	xlog.SetDefault(xlog.NewSimple(os.Stdout))
	// 让 ctx 携带一个日志字段
	xlog.AddAttr(ctx, xlog.String("ClientIP", "127.0.0.1"))

	// 打印日志到 stderr
	xlog.Info(ctx, "hello world")

	// 日志示例（实际输出为 1 行）：

	// {"time":"2024-08-21 23:09:00.671","level":"INFO",
	// "source":{"function":"github.com/xanygo/anygo/xlog_test.ExampleAddAttr","file":"anygo/xlog/example_test.go","line":26},
	// "msg":"hello world","ClientIP":"127.0.0.1"}
}

func ExampleMultiLogger() {
	logger1 := xlog.NewSimple(os.Stderr)
	logger2 := xlog.NewSimple(os.Stdout)

	logger3 := xlog.MultiLogger(logger1, logger2)

	// 日志会同时输出到 stderr 和 stdout
	logger3.Info(context.Background(), "hello world")
}
