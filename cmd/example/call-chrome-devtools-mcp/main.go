//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-03

package main

//  https://github.com/ChromeDevTools/chrome-devtools-mcp

import (
	"context"
	"encoding/json"
	"flag"
	"time"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/cli/xcolor"
	"github.com/xanygo/anygo/xnet/xjsonrpc2"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
)

var url = flag.String("u", "https://oschina.net/", "url")

const service = "chrome-devtools-mcp"

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	{
		err := xservice.LoadFile(ctx, "./chrome-devtools-mcp.json")
		anygo.Must(err)
		xrpc.RegisterIT((&xrpc.Logger{}).Interceptor())
	}

	req1 := &xjsonrpc2.ClientRequest[any]{
		ID:     xjsonrpc2.Int64ID(1),
		Method: "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]string{
				"name":    "go-mcp-client",
				"version": "1.0",
			},
		},
	}
	resp1 := &xjsonrpc2.ClientResponse[json.RawMessage]{}
	anygo.Must(xrpc.Invoke(ctx, service, req1, resp1))
	xcolor.Green("\nresp1=%s\n", resp1.Result)

	req2 := &xjsonrpc2.ClientRequest[any]{
		Method: "initialized",
	}
	anygo.Must(xrpc.Invoke(ctx, service, req2, xrpc.NoResponse()))

	req3 := &xjsonrpc2.ClientRequest[any]{
		ID:     xjsonrpc2.Int64ID(2),
		Method: "tools/call",
		Params: map[string]any{
			"name": "navigate_page",
			"arguments": map[string]any{
				"type": "url",
				"url":  *url,
			},
		},
	}
	resp3 := &xjsonrpc2.ClientResponse[json.RawMessage]{}
	anygo.Must(xrpc.Invoke(ctx, service, req3, resp3))
	xcolor.Green("\nresp3=%s\n", string(resp3.Result))

	req4 := &xjsonrpc2.ClientRequest[any]{
		ID:     xjsonrpc2.Int64ID(3),
		Method: "tools/call",
		Params: map[string]any{
			"name": "evaluate_script",
			"arguments": map[string]any{
				"function": "() => document.title",
			},
		},
	}
	resp4 := &xjsonrpc2.ClientResponse[*parsed]{}
	anygo.Must(xrpc.Invoke(ctx, service, req4, resp4))

	xcolor.Green("\nresp4=%s\n", resp4.Result)
}

type parsed struct {
	Content []struct {
		Type string `json:"type,omitempty"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

func (p *parsed) String() string {
	bf, _ := json.Marshal(p)
	return string(bf)
}
