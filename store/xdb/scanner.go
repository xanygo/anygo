//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-06

package xdb

import (
	"database/sql"
	"fmt"
	"iter"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
	"github.com/xanygo/anygo/store/xdb/dbschema"
)

// ScanRows 读取并解析数据为指定的类型，T 类型可以是 struct、*struct、map[string]any 这三种类型
//
// 如：
//
//	 type User struct{
//		   ID int
//		   Name string
//		   ArticleIDs []int  `db:"article_ids"`          // 默认 codec=json
//		   LinkIDs []int    `db:"link_ids,codec:json"`   // 数据库中存储为 "[1,2,3]"
//		   ID2 []int    `db:"id,codec:csv"`              // 数据库中存储为 "1,2,3"
//		   Enable bool                                   // 数据库中存储为 0 或者 1
//	 }
//
// 可以使用 db 标签来设置相关属性，具体格式为:`db:"数据库字段名,其他 kv 属性"`
// 其他 kv 属性 必须使用 k:v 格式，多个属性之间使用英文逗号连接。
// 若标签中字段名为空，则直接使用 struct 中的字段名。若标签名为 "-",则跳过该字段
//
// 目前支持 codec 属性用于设置序列化、反序列化的方式，codec 可选值：json、csv。
// 如 codec=json, 数据库存储字段值类型应该 string、text、blob 等文本或者二进制类型，
// 字段在 struct 中应是非基础类型（为 Slice、Map、Pointer）。
// 比如：LinkIDs []int    `db:"link_ids,codec:json"` ，在数据库中 查询出来值为 select '[1,2,3]' as link_ids 。
//
// 对于 csv 格式，即为使用英文逗号连接的字段，
// 比如 ID2 []int    `db:"id,codec:csv"` ，在数据库中 查询出来值为 select '1,2,3' as link_ids 。
func ScanRows[T any](rows *sql.Rows) ([]T, error) {
	return ScanRowsLimit[T](rows, -1)
}

func ScanRowsFirst[T any](rows *sql.Rows) (v T, ok bool, err error) {
	items, err := ScanRowsLimit[T](rows, -1)
	if err != nil {
		return v, false, err
	}
	if len(items) == 1 {
		return items[0], true, nil
	}
	return v, false, nil
}

func ScanRowsLimit[T any](rows *sql.Rows, limit int) ([]T, error) {
	defer rows.Close()
	var result []T
	for item, err := range ScanRowsIter[T](rows) {
		if err != nil {
			return result, err
		}
		result = append(result, item)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result, nil
}

// ScanRowsIter 依次读取并解析 Rows，需要自行调用 rows，Close()
func ScanRowsIter[T any](rows *sql.Rows) iter.Seq2[T, error] {
	var zero T
	return func(yield func(T, error) bool) {
		cols, err := rows.Columns()
		if err != nil {
			yield(zero, err)
			return
		}
		rv := reflect.ValueOf(&zero).Elem()
		rt := rv.Type()

		for rows.Next() {
			var item T
			switch rt.Kind() {
			case reflect.Map:
				item, err = scanRowsAsMap[T](rows, cols)
			default:
				item, err = scanRowsAsStruct[T](rows, cols)
			}
			if err != nil {
				yield(zero, err)
				return
			}
			if !yield(item, nil) {
				return
			}
		}
		if err = rows.Err(); err != nil {
			yield(zero, err)
		}
	}
}

func scanRowsAsMap[T any](rows *sql.Rows, cols []string) (T, error) {
	var zero T

	rt := reflect.TypeOf(zero)
	if rt.Kind() != reflect.Map ||
		rt.Key().Kind() != reflect.String ||
		rt.Elem().Kind() != reflect.Interface {
		return zero, fmt.Errorf("type %T is not map[string]any", zero)
	}

	values := make([]any, len(cols))
	valuePtrs := make([]any, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return zero, err
	}

	m := reflect.MakeMap(rt)
	for i, col := range cols {
		val := values[i]
		if b, ok := val.([]byte); ok {
			val = string(b)
		}
		m.SetMapIndex(reflect.ValueOf(col), reflect.ValueOf(val))
	}

	return m.Interface().(T), nil
}

func scanRowsAsStruct[T any](rows *sql.Rows, cols []string) (T, error) {
	var v T
	rv := reflect.ValueOf(&v).Elem()
	rt := rv.Type()

	var structVal reflect.Value
	switch rt.Kind() {
	case reflect.Pointer:
		structVal = reflect.New(rt.Elem()).Elem()
	case reflect.Struct:
		structVal = rv
	default:
		return v, fmt.Errorf("scan type %T is not struct or *struct", v)
	}

	scanTargets := make([]any, len(cols))
	columnToField := make(map[string]reflect.Value) // 用于存储  dbFieldName -> structFieldName 的关系
	tags := make(map[string]xstruct.Tag)

	tn := dbschema.TagName()

	var doScanField func(tv reflect.Value) error

	doScanField = func(tv reflect.Value) error {
		err := zreflect.RangeStructFields(tv.Type(), func(field reflect.StructField) error {
			if field.Anonymous {
				fv := tv.FieldByName(field.Name)
				switch field.Type.Kind() {
				case reflect.Struct:
					return doScanField(fv)
				case reflect.Ptr:
					if fv.IsNil() {
						if fv.CanSet() {
							fv.Set(reflect.New(field.Type.Elem()))
						} else {
							return fmt.Errorf("field %q Cannot Set", field.Name)
						}
					}
					return doScanField(fv.Elem())
				default:
					panic(fmt.Sprintf("what Anonymous kind %v, filed=%q", field.Type.Kind(), field.Name))
				}
			}
			if !field.IsExported() {
				return nil
			}
			tag := xstruct.ParserTagCached(field.Tag, tn)
			name := tag.Name()
			if name == "" || name == "-" {
				// 此字段不需要解析
				return nil
			}
			name = strings.ToLower(name)
			columnToField[name] = tv.FieldByName(field.Name)
			tags[name] = tag

			return nil
		})
		return err
	}

	if err := doScanField(structVal); err != nil {
		var zero T
		return zero, err
	}

	serializerFields := make(map[string]int)
	for idx, col := range cols {
		name := strings.ToLower(col)
		fieldValue, ok := columnToField[name]
		if !ok {
			var dummy any
			scanTargets[idx] = &dummy
			continue
		}

		if !isComplexKind(fieldValue.Kind()) {
			scanTargets[idx] = fieldValue.Addr().Interface()
			continue
		}

		var s sql.NullString
		scanTargets[idx] = &s
		serializerFields[name] = idx
	}

	if err := rows.Scan(scanTargets...); err != nil {
		return v, err
	}

	if len(serializerFields) > 0 {
		for name, idx := range serializerFields {
			tag := tags[name]
			codec := dbschema.TagCodecName(tag)
			// 从 scanTargets 里取出 sql.NullString
			sPtr := scanTargets[idx].(*sql.NullString)
			fieldValue := columnToField[name]
			if err := unmarshallingField(fieldValue, codec, sPtr); err != nil {
				return v, fmt.Errorf("unmarshalling field %q: %w", name, err)
			}
		}
	}

	if rt.Kind() == reflect.Pointer {
		rv.Set(reflect.ValueOf(structVal.Addr().Interface()))
	} else {
		rv.Set(structVal)
	}

	return v, nil
}

func unmarshallingField(field reflect.Value, codec string, sPtr *sql.NullString) error {
	if !sPtr.Valid || sPtr.String == "" {
		// NULL 或空字符串，跳过或置为零值
		return nil
	}

	decoder, err := dbcodec.Find(codec)
	if err != nil {
		return err
	}
	ptr := reflect.New(field.Type())
	if err = decoder.Decode(sPtr.String, ptr.Interface()); err != nil {
		return err
	}
	field.Set(ptr.Elem())
	return nil
}

func isComplexKind(k reflect.Kind) bool {
	return k == reflect.Struct || k == reflect.Slice || k == reflect.Array || k == reflect.Map || k == reflect.Pointer
}
