//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

// Package dbcodec 用于支持 struct 字段在存取数据库过程中编解码
//
// 已内置：csv,json,text,date,date_time,timespan,milliseconds,microseconds,nanoseconds
//
//	使用 Register 方法可注册自定义类型的编解码器
//
// 如：
//
//	type User struct {
//		 RegisterTime time.Time `db:"reg_time,codec=timespan"`
//	}
package dbcodec
