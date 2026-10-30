// Package trustheader 可信的 HTTP Header 字段管理
//
// 有些逻辑，需要读取 HTTP 的 Header，而大多数 Header 都是不可信的，只有网关设置写入的才是可信的。
// 对于这一类网关设置的字段，可以在程序启动阶段，设置为可信的。
//
// 如 xhttp.ClientIP 和 xhttp.ClientScheme ，会使用 X-Forwarded-Proto、Cf-Visitor、Cf-Connecting-Ip、
// X-Real-IP、X-Forwarded-For、X-Forwarded-Port 等 HTTP Header 字段。当字段是可信的时候，会优先使用。
//
// 注意： 可信字段，应根据程序部署情况，配置在配置文件中，在程序启动时设置导入。如果上游网关是 cloudflare,可以将
// Cf-Visitor、Cf-Connecting-Ip 配置为可信字段，否则不应配置。
//
// # 1. 配置和初始化：
//
// 在 conf/app.yml 中添加配置段：
//
//	# 可信的 HTTP Header (按照实际情况配置)
//	TrustHeader:
//	  - X-Forwarded-Proto
//	  - X-Forwarded-Port
//	  - X-Real-Ip
//	  - X-Forwarded-For
//	  - Cf-Visitor
//	  - Cf-Connecting-Ip
//
// 然后在程序启动阶段添加如下代码加载配置：
//
//	trustheader.Add(xattr.GetOtherDefault[[]string]("TrustHeader", nil)...)
//
// # 2. nginx 配置参考：
//
//	location / {
//		proxy_pass http://127.0.0.1:8080/;
//		proxy_set_header Host $host;
//		proxy_set_header X-Real-IP $remote_addr;
//		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
//
//		# 传递客户端访问时的协议
//		proxy_set_header X-Forwarded-Proto $scheme;
//
//		# 传递客户端访问的 HTTPS 端口
//		proxy_set_header X-Forwarded-Port $server_port;
//	}
package trustheader
