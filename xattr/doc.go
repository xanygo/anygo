//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

// Package xattr 应用的属性信息，如应用名，应用的根目录，配置目录，日志目录，数据目录等
//
//	main.go 代码：
//
//	package main
//
// import
//
//	var c = flag.String("conf", "conf/app.yml", "app main config file")
//
//	func main() {
//		flag.Parse()
//		xattr.MustInitAppMain(*c, xcfg.Parse)  // 加载主配置文件
//
//	// ......
//	}
//
// 主配置文件可以这样放：
//
//	xxx/conf/app.yml          <-- 普通
//	xxx/conf/product/app.yml  <-- 子目录中放主配置文件
//
//	xxx/conf_product/app.yml  <-- 不同环境的配置文件以 conf_xxx 命名
//	xxx/conf_dev/app.yml
package xattr
