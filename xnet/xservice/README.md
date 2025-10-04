


## 配置文件格式：
```toml
Name = "demo"
ConnectTime = 1000
ReadTimeout = 1000
HandshakeTimeout = 1000

Protocol="resp3" # 交互协议

UseProxy = "proxy1" # 可选，使用指定的代理

[HTTP]
Host = "demo.com"
HTTPS = true
[HTTP.Header]
KA = ["a"]


[Proxy]
Protocol = "HTTP"


# 网络连接池参数，可选
[ConnPool]


# redis 协议的下游专属，可选
[Redis]
Username = "user"
Password = "psw"

[TLS]
SkipVerify = false # 是否跳过安全验证
ServerName = "example.com"

[DownStream]
LoadBalancer = "rr"
Address  = ["127.0.0.1:80"]

[DownStream.IDC.bj]
Address  = ["127.0.0.1:80"]
```