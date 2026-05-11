//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

// Package xnet 支持拦截器 ( Interceptor ) 和 对冲请求/备份请求 ( Hedging Request / backup request) 的网络 API 。
//
// 框架内基础行为统一:
//
//	① 所有创建网络连接，都统一使用 DialContext 方法
//	② 所有的域名解析逻辑，统一使用 LookupIP 方法
//
// 如此，使 xrpc (通用的 RPC 客户端)具备非常完整的 Interceptor 的能力，可以对网络交互中的底层行为进行观察和拦截，实现一些高级功能。
// 比如，使用 xrpc 时，只需要传入 xrpc.OptBlockPrivateIPs() 这个 xrpc.Option ，就能实现安全功能。
// 若 rpc client 请求的下游地址是有用户传入的，通过 OptBlockPrivateIPs 来避免传入的地址是内网地址，避免查询返回内网数据。
// 比如，用户传入 http://127.0.0.1:8080/api 或者 http://localhost:8080/api
// 或者 https://some-domain/api (some-domain 会解析为内网 IP 地址) 等地址，通过接口来探测内网资源文件。
//
//	xrpc.OptBlockPrivateIPs() 底层就是基于 ResolverInterceptor 实现，若最终目标 IP 地址是内网或本地环回地址，RPC Client 直接拒绝。
package xnet
