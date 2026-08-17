package encoder

import (
	"github.com/xanygo/anygo/store/xdb/dbschema"
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

// EncodeInsert 将结构体转成 map[string]any，用于 SQL insert
func EncodeInsert[T any](fy dbtype.Dialect, data T, cols ...string) (map[string]any, error) {
	schema, err := dbschema.Schema(fy, data)
	if err != nil {
		return nil, err
	}
	return Encoder[T]{Dialect: fy, Action: ActionInsert, OnlyFields: cols, Schema: schema}.Encode(data)
}

// func EncodeBatch[T any](fy dbtype.Dialect, items []T, fields ...string) ([]map[string]any, error) {
//	if len(items) == 0 {
//		return nil, errors.New("no value to encode")
//	}
//	result := make([]map[string]any, 0, len(items))
//	for _, item := range items {
//		data, err := Encode(fy, item, fields...)
//		if err != nil {
//			return nil, err
//		}
//		result = append(result, data)
//	}
//	return result, nil
// }
