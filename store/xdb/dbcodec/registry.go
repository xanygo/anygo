//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

import (
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

var codecs = map[string]dbtype.Codec{}

func Register(codec dbtype.Codec) bool {
	name := codec.Name()
	if _, ok := codecs[name]; ok {
		return false
	}
	codecs[name] = codec
	return true
}

func FindByName(names ...string) dbtype.Codec {
	for _, name := range names {
		if codec, ok := codecs[name]; ok {
			return codec
		}
	}
	return nil
}

func FindByKind(kind dbtype.Kind) dbtype.Codec {
	switch kind {
	case dbtype.KindString:
		return Text{}
	case dbtype.KindBinary:
		return Binary{}
	case dbtype.KindArray, dbtype.KindJSON:
		return JSON{}
	case dbtype.KindDateTime:
		return DateTime{}
	case dbtype.KindDate:
		return Date{}
	default:
		return Native{}
	}
}

func init() {
	// 时间相关的
	Register(Date{})
	Register(DateTime{})
	Register(TimeSpan{})
	Register(Milliseconds{})
	Register(Microseconds{})
	Register(Nanoseconds{})

	// 文本格式相关的：
	Register(CSV{})
	Register(JSON{})
	Register(Text{})

	// 二进制
	Register(Binary{})

	// 数据库驱动原生支持的数据类型
	Register(Native{})
}
