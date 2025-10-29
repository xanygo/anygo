
## 配置文件格式：
```toml
Name = "demo"          # Name 需要和文件名保持一致: demo.toml
ConnectTime = 1000
ReadTimeout = 1000
HandshakeTimeout = 1000

Protocol="HTTP"     # 交互协议

UseProxy = "proxy1" # 可选，使用指定的代理

[HTTP]
Host = "demo.com"
HTTPS = true
[HTTP.Header]
KA = ["a"]

# [Proxy]            # 此服务是代理时配置 
# Protocol = "HTTP"  # 代理的协议，目前支持 HTTP、HTTPS

# 网络连接池参数，可选
[ConnPool]
Name = "Short"        # 默认为 Short，可选为 Long-长连接
# MaxOpen = 0         # 单个下游最大打开数量,<= 0 为不限制
# MaxIdle =0          # 单个下游最大空闲数，应 <= MaxOpen,<=0 为不允许存在 Idle 元素
# MaxLifeTime =0      # 单个下游最大使用时长,单位毫秒，超过后将被销毁, <=0 为不限制
# MaxIdleTime = 0     # 单个下游最大空闲等待时间,单位毫秒，超过后将被销毁, <=0 为不限制
# MaxPoolIdleTime =0  # 单位毫秒，当超过此时长未被使用后,关闭并清理对应的 Pool,<=0 时使用默认值 10 minute   

# redis 协议的下游专属，可选
[Redis]
Username = "user"
Password = "psw"

# [TLS]
# Disable = false             # 可选，是否不使用 TLS
# SkipVerify = false          # 可选，是否跳过安全验证
# ServerName = "example.com"  # 可选，用于校验证书中的服务器名称
# CAFile = ""                 # 可选，根证书（CA），用于信任自签名证书,ca.crt 的内容
# CertFile =""                # 可选，客户端证书,  client.crt 的内容
# KeyFile =""                 # 可选，客户端证私钥， client.key 的内容

[DownStream]
LoadBalancer = "rr"
Address  = ["127.0.0.1:80"]

[DownStream.IDC.bj]
Address  = ["127.0.0.1:80"]
```