//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

type Codec interface {
	Name() string

	// Encode 编码为基础类型
	Encode(a any) (any, error)

	// Decode 解码
	Decode(b string, a any) error

	// Kind 数据库中存储的数据类型
	Kind() Kind
}
