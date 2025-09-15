
# 1. HTTP Proxy 协议/报文
## 1. 建立隧道
```
CONNECT www.example.com:443 HTTP/1.1\r\n
Host: www.proxy.com:1080\r\n
Proxy-Connection: Keep-Alive\r\n
\r\n
```
注：上述 `www.example.com:443` 是被代理的 url 地址中的 Host+Port

## 2. 代理应答
```
HTTP/1.1 200 Connection Established\r\n
Proxy-Agent: ProxyServer/1.0\r\n
\r\n
```

##  3. TLS 握手
```
<16 03 01 02 00 01 00 01 fc 03 03 ...>
```
注：若被代理的 url 不是 HTTPS 地址，则跳过此步骤

## 4. 传输加密的 HTTP 报文
解密后实际内容可能是: 
```
GET /index.html HTTP/1.1\r\n
Host: www.example.com\r\n
User-Agent: curl/7.88.0\r\n
Accept: */*\r\n
\r\n
```
