//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

import (
	"fmt"
)

var codecs = map[string]Codec{}

func Register(codec Codec) {
	codecs[codec.Name()] = codec
}

func Find(name string) (Codec, error) {
	codec, ok := codecs[name]
	if ok {
		return codec, nil
	}
	return nil, fmt.Errorf("codec %q not found", name)
}

func init() {
	// 时间相关的
	Register(Date{})
	Register(DateTime{})
	Register(TimeSpan{})

	// 文本格式相关的：
	Register(CSV{})
	Register(JSON{})
	Register(Text{})
}
