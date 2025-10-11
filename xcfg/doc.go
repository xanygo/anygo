//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

// Package xcfg 通用的配置解析组件，已默认内置支持 .json 和 .xml 格式的配置文件
//
// # 1. 支持的文件格式
//
// 默认内置支持 .json 和 .xml 格式, 支持使用 WithParser 和 MustWithParser 注册新的解析器。
//
// # 2. 配置文件路径
//
// 在使用 Parse 方法解析配置文件或者使用 Exists 判断文件是否存在时，传入的配置文件路径( path 参数)
// 可以是绝对路径，也可以是相对于配置文件根目录 ( 读取自 xattr.ConfDir() ） 的相对路径。
// 同时，文件后缀是可选的，当查找文件不存在时，会添加上支持的后缀依次去判断，
// 如 Exists("app.toml") 会补充为完整路径 {ConfDir}/app.json、{ConfDir}/app.xml 等去判断。
//
// # 3. Hooks
//
// Hook 功能会在解析配置内容时自动运行。目前内置如下几个功能：
//
// 1. 读取环境变量的值
//
// 在配置内容中使用 {env.变量名} 或者 {env.变量名|默认值}。 变量名应该是有效的环境变量名称。
// 环境变量的值会使用 os.Getenv( 变量名 ) 方法读取，若环境变量不存在，会返回空字符串。
// 如 user.json：
//
//	{"name":"{env.userName}","age":{env.age|18}}
//
// 若有环境变量值( export userName=hello ),在解析前，配置内容会被替换为：
//
//	{"name":"hello","age":18}
//
// 2. 读取 xattr 属性的值
//
// 在配置内容中使用 {xattr.属性名}，可读取到 xattr 的属性值，支持的属性名仅限如下：
//
//	RootDir : 应用根目录地址，如 /home/work/myapp
//	IDC     :  应用所属机房, 如 online
//	DataDir :  应用数据目录地址，如 /home/work/myapp/data
//	ConfDir :  应用配置目录地址, 如 /home/work/myapp/conf
//	TempDir :  应用临时文件目录地址, 如 /home/work/myapp/temp
//	LogDir  :  应用日志文件目录地址, 如 /home/work/myapp/log
//	RunMode :  应用运行模式,如 product
//
// 3. Go Template 语法
//
// 该功能默认未启用,需要在配置文件的头部添加如何内容以启用该功能：
//
//	# hook.template  Enable=true
//
// 以下是一个例子 (  server.toml )：
//
//	# hook.template  Enable=true
//
//	 Port = {env.Port1}
//	 {{ include "sub1/*.toml" }}
//	 {{ include "sub2/z.toml" }}
//
// template 语法详见 https://pkg.go.dev/text/template 。
// 默认内置如下自定义函数：
//
//	include  : 包含指定的文件，如 include "sub2/z.toml"，include "sub1/*.toml"
//	env      : 读取环境变量的值，如 env "userName"
//	contains : 即 strings.Contains，如 contains "hello" "h"
//	prefix   : 即 strings.HasPrefix，如 contains "hello" "h"
//	suffix   : 即 strings.HasSuffix，如 contains "hello" "o"
//
// # 4. 数据校验( Validator )
//
// 使用 Parser、ParseBytes 解析配置后，得的的结果可能是非预期的，如必填字段为空或者格式错误。
//
// 如典型的解析配置然后校验的逻辑为：
//
//	 if err := xcfg.Parse("abc.json", &info); err != nil {
//		return err
//	 }
//	 // 下面是校验逻辑
//	 if info.A == "" {
//		return fmt.Errorf("required field A is empty")
//	 }
//
// 若 info 的 struct 对象实现了 AutoChecker 接口，则在调用 Parse 解析配置内容之后，会自动调用校验逻辑。
//
//	type AutoChecker interface {
//		AutoCheck() error
//	}
//
// 即让 info 的 struct 定义为：
//
//	type Info struct {
//		A string
//	}
//
//	 func (in *Info)AutoCheck() error{
//	 	if in==nil || in.A==""{
//	 		return fmt.Errorf("invalid info")
//	 	}
//	 	return nil
//	 }
//
// 除此之外，还可以使用 Validator 实现自动校验，使用的前提是给 DefaultValidator 赋值。
// 如可以使用 github.com/go-playground/validator/v10 来初始化 DefaultValidator。
//
//	type Info struct {
//		A string `validate:"required"` // A 字段必填，不能为空
//	}
//
// 具体顺序为：
//
//  1. Validator
//  2. AutoChecker
package xcfg
