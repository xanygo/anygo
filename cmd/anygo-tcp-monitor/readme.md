# 转发并打印 TCP 请求和响应

## 1. 安装
```
go install github.com/xanygo/anygo/cmd/anygo-tcp-monitor@master
```
## 2. 使用

### 2.1 Params
```bash
anygo-tcp-monitor -help
```

```text
Usage of anygo-tcp-monitor:
  -l string
        local server listen address (default ":8200")
  -p string
        print type, s:string, b:binary, c:char (default "s")
  -r string
        remote server address,eg example.com:80
```

### 2.2 打印 HTTP 请求
```bash
anygo-tcp-monitor -r github.com:80  # 发送请求后，此控制台输出请求和响应内容
```

发送请求：
```bash
curl -v --connect-to github.com:80:127.0.0.1:8200 http://github.com/
```

### 2.3 打印 HTTPS 请求
```bash
anygo-tcp-monitor -r github.com:440  # 发送请求后，此控制台输出请求和响应内容
```

发送请求：
```
curl -v --connect-to github.com:443:127.0.0.1:8200 https://github.com/
```