//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-21

// Package xkv 一个通用的 KV 存储的定义和实现
//
// 内置存储实现：
//
//   - FileStore 文件系统存储，key-Value 是 string 类型
//   - NopStore 黑洞，key 是 string，Value 是 任意类型
//   - MemoryStore 内存存储，key-Value 是 string 类型
//   - RedisStore  使用 redis 存储 ，key-Value 是 string 类型
//   - DatabaseStore 使用数据库存储(支持 sqlite、pg、mysql 等)，key-Value 是 string 类型
//
// 特殊的：
//   - Transformer: 可以将上面 key-Value 是 string 类型（ StringStorage ）的 Storage实现，转换为支持泛型的类型。
//   - Monitor: 可用于包裹上述各种 Storage 类型，实现 SLA 的观察统计
package xkv
