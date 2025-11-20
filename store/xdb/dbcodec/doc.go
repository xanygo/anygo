//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

// Package dbcodec 用于支持 struct 字段在存取数据库过程中编解码
//
// 已内置：
//
//  1. csv: CSV 格式，可用于 []number 类型
//
//  2. json: JSON 格式，可用于 slice、array、map、struct 类型的字段
//
//  3. text:  文本格式，可用于支持 实现了 TextMarshaler 接口的类型，或者number、string、[]byte 类型的字段
//
//  4. date: 日期格式，可用于 time.Time 类型，输出格式为 "2025-10-10"
//
//  5. date_time: 日期时间格式，可用于 time.Time 类型，输出格式为 "2025-10-10 10:10:10"
//
//     使用 Register 方法可注册自定义类型的编解码器
package dbcodec
