



## 配置文件格式：
```toml
Name = "demo"
ConnectTime = 1000
ReadTimeout = 1000

[HTTP]
Host = "demo.com"
HTTPS = true
[HTTP.Header]
KA = ["a"]


[Proxy]
Protocol = "HTTP"
Use = "proxy1"

[TLS]
SkipVerify = false # 是否跳过安全验证
ServerName = "example.com"

[DownStream]
LoadBalancer = "rr"
Address  = ["127.0.0.1:80"]

[DownStream.IDC.bj]
Address  = ["127.0.0.1:80"]
```