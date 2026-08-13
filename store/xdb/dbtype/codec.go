//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package dbtype

import "fmt"

type Decoder interface {
	Name() string

	// Decode 解码
	Decode(b string, a any) error
}

type Encoder interface {
	Name() string

	// Encode 编码为基础类型
	Encode(a any) (any, error)
}

type Codec interface {
	Encoder
	Decoder
}

func Decode(d Decoder, str string, obj any) error {
	err := d.Decode(str, obj)
	if err == nil {
		return nil
	}
	return fmt.Errorf("%q dbcodec: %w", d.Name(), err)
}
