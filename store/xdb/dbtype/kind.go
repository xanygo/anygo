//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package dbtype

type Kind string

func (k Kind) IsValid() bool {
	return allKinds[k]
}

func (k Kind) Is(name string) bool {
	return k == Kind(name)
}

func (k Kind) String() string {
	return string(k)
}

const (
	KindInvalid Kind = "invalid"

	KindNative   Kind = "native"    // 原生类型，数据库驱动已支持该类型，不需要额外的 Codec
	KindAutoJSON Kind = "auto_json" // 需要数据库方言来不判断类型，若方言判断不出来，则默认使用 json 编解码

	KindString Kind = "string"

	KindInt   Kind = "int"
	KindInt8  Kind = "int8"
	KindInt16 Kind = "int16"
	KindInt32 Kind = "int32"
	KindInt64 Kind = "int64"

	KindUint   Kind = "uint"
	KindUint8  Kind = "uint8"
	KindUint16 Kind = "uint16"
	KindUint32 Kind = "uint32"
	KindUint64 Kind = "uint64"

	KindBoolean Kind = "boolean"

	KindFloat32 Kind = "float32"
	KindFloat64 Kind = "float64" // 8 字节（双精度）

	KindBinary Kind = "binary"
	KindArray  Kind = "array" // 数组类型
	KindJSON   Kind = "json"

	KindDate     Kind = "date"      // 仅日期 '2000-01-01'
	KindDateTime Kind = "date_time" // 日期和时间 '2000-01-01 00:00:00'
)

var allKinds = map[Kind]bool{
	KindString: true,

	KindInt:   true,
	KindInt8:  true,
	KindInt16: true,
	KindInt32: true,
	KindInt64: true,

	KindUint:   true,
	KindUint8:  true,
	KindUint16: true,
	KindUint32: true,
	KindUint64: true,

	KindBoolean: true,

	KindFloat32: true,
	KindFloat64: true,

	KindBinary: true,
	KindJSON:   true,
	KindArray:  true,

	KindNative:   true,
	KindAutoJSON: true,

	KindDate:     true,
	KindDateTime: true,
}
