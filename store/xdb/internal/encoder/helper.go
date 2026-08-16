package encoder

import (
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

// EncodeInsert 将结构体或 map 转成 map[string]any，用于 SQL insert
func EncodeInsert[T any](fy dbtype.Dialect, data T, cols ...string) (map[string]any, error) {
	return Encoder[T]{Dialect: fy, Action: ActionInsert, OnlyFields: cols}.Encode(data)
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
