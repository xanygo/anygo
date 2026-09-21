package dbcodec

import (
	"reflect"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xdb/dbtype"
	"github.com/xanygo/anygo/xenc/xcodec"
)

var _ dbtype.Codec = (*Native)(nil)
var _ dbtype.HasKind = (*Native)(nil)
var _ nativeType = (*Native)(nil)

type nativeType interface {
	native()
}

// Native 数据库原生支持的类型
type Native struct{}

// native implements [nativeType].
func (r *Native) native() {
}

func (r Native) Kind() dbtype.Kind {
	return dbtype.KindNative
}

func (r Native) Name() string {
	return "native"
}

// Encode 对数据编码，
//
// 在使用的时候，会先调用 Dialect.EncodeValue，然后再对结果调用此 Encode 方法
func (r Native) Encode(a any) (any, error) {
	if a == nil {
		return nil, nil
	}
	rv := reflect.ValueOf(a)
	// 将数组转换为 slice
	if rv.Kind() == reflect.Array {
		return zreflect.ArrayToSlice(rv), nil
	}

	return a, nil
}

func (r Native) Decode(str string, obj any) error {
	// 若是基础类型，则先尝试直接解析值
	if err := xcodec.Text.UnmarshalCustom([]byte(str), obj, false); err == nil {
		return nil
	}
	return xcodec.UnmarshalFromString(xcodec.Text, str, obj)
}

func IsNative(c dbtype.Codec) bool {
	_, ok := c.(nativeType)
	return ok
}
