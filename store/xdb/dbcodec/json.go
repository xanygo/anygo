//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

import (
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/xcodec"
)

var _ dbtype.Codec = (*JSON)(nil)
var _ dbtype.HasKind = (*JSON)(nil)

// JSON 格式，可用于 slice、array、map、struct ,*struct 类型的字段,输出为 JSON 字符串
type JSON struct{}

func (j JSON) Kind() dbtype.Kind {
	return dbtype.KindJSON
}

func (j JSON) Name() string {
	return "json"
}

func (j JSON) Encode(obj any) (any, error) {
	return xcodec.EncodeToString(xcodec.JSON, obj)
}

func (j JSON) Decode(str string, obj any) error {
	if len(str) == 0 {
		return nil
	}
	return xcodec.DecodeFromString(xcodec.JSON, str, obj)
}
