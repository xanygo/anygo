
## 配置文件格式 ( toml ) 示例：
```toml
Name = "demo"             # Name 需要和文件名保持一致: demo.toml
ConnectTime = 1000        # 创建网络连接的超时时间，可选，单位 ms，默认值为 5 秒
ReadTimeout = 1000        # socket Read 超时时间，可选，单位 ms，默认值为 10 秒
WriteTimeout = 1000       # socket Write 超时时间，可选，单位 ms，默认值为 5 秒
HandshakeTimeout = 1000   # 握手超时时间

Protocol="HTTP"     # 交互协议,可选

UseProxy = "proxy1" # 可选，使用指定的代理

# HTTP 协议专属配置，可选
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

# 用于在连接创建完成后，业务正式使用前，执行会话开启的逻辑，可选
[SessionInit]
Name = "HTTP-Upgrade"
[SessionInit.Params]  # HTTP-Upgrade 专属配置
Method = "POST"
URI = "/api/v1/stream"
Protocol = "JSON-RPC2"

# 下游地址列表，必填
[DownStream]
# 负载均衡策略，可选。可选值： RoundRobin（依次轮询，默认，可简写为 rr）、Random (随机)
LoadBalancer = "rr"

# 默认地址列表，若对应 IDC 有地址则优先使用 IDC 的地址，否则才被使用
# Address 可以配置 IP+Port, 域名+Port, 包含IP/域名+Port列表的文件地址等，比如：
# IP+Port     ： "127.0.0.1:80"
# 域名+Port ① ： "api.example.com:443"     负载均衡的时候直接使用，创建连接时，将域名解析为 IP 列表，并选择一个 ip 使用
# 域名+Port ② ： "dns@api.example.com:443" 直接解析为 IP 地址列表(会定期刷新)，负载均衡时从 IP 列表中选择一个 IP 地址使用
# IP/域名+Port列表的文件地址:  "file@server_list.ns"   server_list.ns 是文件地址
# unix socket 文件地址： "unix@fielpath.sock"
# stdio ： 'stdio@{"Path":"echo","Args":["arg1","Args可选"],"Dir":"工作目录，可选"}' 利用子进程的 stdin 和 stdout 通信
Address  = ["127.0.0.1:80"]    

# 分 IDC 的地址列表，可选
[DownStream.IDC.bj]
Address  = ["127.0.0.1:80"]
```