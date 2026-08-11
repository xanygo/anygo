# GoGenx

并发的执行写在 `.go` 文件里的 `//go:generate` 定义的命令。

`go generate ./...` 是顺序依次执行的。而 `gogenx` 则是允许并行执行多条命令。

若定义 `//go:generate` 的文件里包含有 `//go:build` 指令，会启动子进程，以 `go generate dir/file.go` 方式执行。
其他情况，则会解析出命令，直接执行。

项目包含 2 个`go:generate` 指令的速度对比：
1. 使用 `go generate ./...`，平均耗时 0.45s
2. 使用 `gogenx .`，平均耗时 0.16s

## 安装
```bash
go install github.com/xanygo/anygo/cmd/gogenx@master
```

## 使用
````
Usage of gogenx:
  -j int
        maximum number of concurrent generate tasks (default 8)
  -q    quiet, redirect child process stdout and stderr to null
  -tags string
        build tags
````

```bash
gogenx
```