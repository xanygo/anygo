package dbcodec

import (
	"reflect"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/xcodec"
)

var _ dbtype.Codec = (*Native)(nil)
var _ dbtype.HasKind = (*Native)(nil)

// Native 数据库原生支持的类型
type Native struct{}

func (r Native) Kind() dbtype.Kind {
	return dbtype.KindNative
}

func (r Native) Name() string {
	return "native"
}

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
	return xcodec.DecodeFromString(xcodec.Text, str, obj)
}
