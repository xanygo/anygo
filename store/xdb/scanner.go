//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-06

package xdb

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xcodec"
)

var tagName = xsync.OnceInit[string]{
	New: func() string {
		return "db"
	},
}

func TagName() string {
	return tagName.Load()
}

func SetTagName(name string) {
	if name == "" {
		panic("empty tag name")
	}
	tagName.Store(name)
}

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
	defer rows.Close()

	var result []T

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var zero T
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
			return nil, err
		}
		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
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

	rtStruct := structVal.Type()
	scanTargets := make([]any, len(cols))
	columnToField := make(map[string]int, rtStruct.NumField()) // 用于存储  dbFieldName -> structField Index 的关系
	tags := make(map[string]xstruct.Tag)

	tn := TagName()
	for i := 0; i < rtStruct.NumField(); i++ {
		field := rtStruct.Field(i)
		tag := xstruct.ParserTag(field.Tag.Get(tn))
		name := tag.Name()
		if name == "-" {
			// 此字段不需要解析
			continue
		}

		if name == "" {
			name = field.Name
		}
		name = strings.ToLower(name)
		columnToField[name] = i
		tags[name] = tag

		if tag.Name() == "" {
			// 将驼峰命名转换为使用 _ 命名，如 UserID -> user_id
			snakeName := xstr.ToSnakeCase(field.Name)
			if snakeName != name {
				columnToField[snakeName] = i
			}
		}
	}

	serializerFields := make(map[string]int)
	for idx, col := range cols {
		name := strings.ToLower(col)
		fieldIndex, ok := columnToField[name]
		if !ok {
			var dummy any
			scanTargets[idx] = &dummy
			continue
		}

		field := structVal.Field(fieldIndex)
		if !isComplexKind(field.Kind()) {
			scanTargets[idx] = structVal.Field(fieldIndex).Addr().Interface()
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
			codec := tag.Value("codec")
			// 从 scanTargets 里取出 sql.NullString
			sPtr := scanTargets[idx].(*sql.NullString)
			field := structVal.Field(columnToField[name])
			if err := unmarshallingField(field, codec, sPtr); err != nil {
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

	rawStr := sPtr.String
	switch codec {
	case "json", "":
		ptr := reflect.New(field.Type())
		if len(rawStr) > 0 {
			if err := xcodec.DecodeFromString(xcodec.JSON, rawStr, ptr.Interface()); err != nil {
				return err
			}
		}
		field.Set(ptr.Elem())
		return nil
	case "csv":
		parts := strings.Split(rawStr, ",")
		slice := reflect.MakeSlice(field.Type(), 0, len(parts))
		elemType := field.Type().Elem()
		for idx, p := range parts {
			elem := reflect.New(elemType).Elem()
			if err := setFromString(elem, strings.TrimSpace(p)); err != nil {
				return fmt.Errorf("decode[%d](%q):%w", idx, p, err)
			}
			slice = reflect.Append(slice, elem)
		}
		field.Set(slice)
		return nil
	default:
		if decoder, ok := codecs[codec]; ok {
			ptr := reflect.New(field.Type())
			if err := xcodec.DecodeFromString(decoder, rawStr, ptr.Interface()); err != nil {
				return err
			}
			field.Set(ptr.Elem())
			return nil
		}
		return fmt.Errorf("unsupported codec %q", codec)
	}
}

var codecs = map[string]xcodec.Codec{}

func isComplexKind(k reflect.Kind) bool {
	return k == reflect.Struct || k == reflect.Slice || k == reflect.Array || k == reflect.Map || k == reflect.Pointer
}

func isBasicKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

func setFromString(v reflect.Value, s string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var x int64
		_, err := fmt.Sscan(s, &x)
		v.SetInt(x)
		return err
	case reflect.Float32, reflect.Float64:
		var f float64
		_, err := fmt.Sscan(s, &f)
		v.SetFloat(f)
		return err
	default:
	}
	return nil
}
