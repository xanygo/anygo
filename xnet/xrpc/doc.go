//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-29

// Package xrpc 通用的 RPC Client
//
// 目前已实现 HTTP 协议(xhttp/xhttpc)、Redis Resp3 协议(store/xredis)、SMTP(xnet/xsmtp)
//
//		1.支持负载均衡策略：随机(Random)、轮询(RoundRobin,默认的)
//		2.支持连接池，默认短连接(Short)、可选使用长连接(Long)
//		3.将协议握手逻辑独立，若协议有注册，则在拨号完成后即握手
//		4.支持 TLS，支持使用自签名证书
//		5.支持多种名字服务，如静态的 IP+Port、域名+Port、主机列表来自文件
//	 6.域名提前解析，并异步刷新
package xrpc
