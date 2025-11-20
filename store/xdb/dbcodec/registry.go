//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

import (
	"fmt"

	"github.com/xanygo/anygo/store/xdb/dbtype"
)

var codecs = map[string]dbtype.Codec{}

func Register(codec dbtype.Codec) {
	codecs[codec.Name()] = codec
}

func Find(names ...string) (dbtype.Codec, error) {
	for _, name := range names {
		if codec, ok := codecs[name]; ok {
			return codec, nil
		}
	}
	return nil, fmt.Errorf("codec %q not found", names)
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
