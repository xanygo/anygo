//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

import (
	"github.com/xanygo/anygo/xdb/dbtype"
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

var kind2Codec = map[dbtype.Kind]dbtype.Codec{
	dbtype.KindString:       &Text{},
	dbtype.KindBinary:       &Binary{},
	dbtype.KindArray:        &JSON{},
	dbtype.KindJSON:         &JSON{},
	dbtype.KindDateTime:     &DateTime{},
	dbtype.KindDate:         &Date{},
	dbtype.KindTimespan:     &TimeSpan{},
	dbtype.KindMilliseconds: &Milliseconds{},
	dbtype.KindMicroseconds: &Microseconds{},
	dbtype.KindUUID:         &UUID{},
}

var instNative = &Native{}

func FindByKind(kind dbtype.Kind) dbtype.Codec {
	if c, ok := kind2Codec[kind]; ok {
		return c
	}
	return instNative
}

func init() {
	// 时间相关的
	Register(&Date{})
	Register(&DateTime{})
	Register(&TimeSpan{})
	Register(&Milliseconds{})
	Register(&Microseconds{})
	Register(&Nanoseconds{})

	// 文本格式相关的：
	Register(&CSV{})
	Register(&JSON{})
	Register(&Text{})

	// 二进制
	Register(&Binary{})

	Register(&UUID{})

	// 数据库驱动原生支持的数据类型
	Register(&Native{})
}
