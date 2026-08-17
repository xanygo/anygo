//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package encoder

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbschema"
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/xerror"
)

type Encoder[T any] struct {
	Action       Action
	Dialect      dbtype.Dialect // 方言，必填
	OnlyFields   []string       // 当不为空时，输出结果的 key 只限定此列表中的
	IgnoreFields []string       // 当不为空时，输出结果的 key 限定不能是此列表中的
}

func (e Encoder[T]) WithAction(action Action) Encoder[T] {
	return Encoder[T]{
		Action:       action,
		Dialect:      e.Dialect,
		OnlyFields:   e.OnlyFields,
		IgnoreFields: e.IgnoreFields,
	}
}

func (e Encoder[T]) Encode(data T) (map[string]any, error) {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil, fmt.Errorf("invalid value: %v", v)
	}

	// 支持指针类型
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, fmt.Errorf("nil pointer: %#v", data)
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		schema, err := dbschema.Schema(e.Dialect, data)
		if err != nil {
			return nil, err
		}
		return e.encodeStruct(v, schema)
	case reflect.Map:
		return e.encodeMap(v)
	default:
		return nil, fmt.Errorf("unsupported type %T", data)
	}
}

func (e Encoder[T]) EncodeBatch(items ...T) ([]map[string]any, error) {
	if len(items) == 0 {
		return nil, errors.New("no value to encode")
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		data, err := e.Encode(item)
		if err != nil {
			return nil, err
		}
		result = append(result, data)
	}
	return result, nil
}

// Diff 得到 newValue 和 oldValue 中值不一样的字段和值
func (e Encoder[T]) Diff(newValue T, oldValue T) (map[string]any, error) {
	data1, err1 := e.Encode(newValue)
	if err1 != nil {
		return nil, err1
	}
	// 使用 ActionSelect，确保 oldValue 的 Created 和 Updated 等字段不会被自动更新并编码
	data2, err2 := e.WithAction(ActionSelect).Encode(oldValue)
	if err2 != nil {
		return nil, err2
	}
	diff := make(map[string]any, len(data1))
	for k1, v1 := range data1 {
		v2, ok := data2[k1]
		if !ok || !reflect.DeepEqual(v1, v2) {
			diff[k1] = v1
		}
	}
	return diff, nil
}

// encodeStruct 处理 struct
func (e Encoder[T]) encodeStruct(v reflect.Value, schema *dbtype.TableSchema) (map[string]any, error) {
	result := make(map[string]any, len(e.OnlyFields))
	err := e.withStruct(v, func(name string, tag xstruct.Tag, field reflect.StructField, value reflect.Value) error {
		fs, err := schema.ColumnByName(name)
		if err != nil {
			return err
		}
		if e.Action == ActionUpdate && fs.AutoIncrement {
			// 更新的时候，将自增长类型的字段忽略掉
			// 如 sql server，在更新的字段列表中，包含【自增长】类型的字段，
			// 将报错：mssql: Cannot update identity column 'id'
			return nil
		}
		encodedVal, err := e.encodeStructFieldValue(fs, value.Interface())
		if err != nil {
			if errors.Is(err, xerror.SkipOne) {
				return nil
			}
			return fmt.Errorf("encode field %q: %w", field.Name, err)
		}
		result[name] = encodedVal
		return nil
	})
	return result, err
}

var sliceEmpty = map[string]bool{}

func (e Encoder[T]) sliceToMapTrue(s []string) map[string]bool {
	if len(s) == 0 {
		return sliceEmpty
	}
	return xslice.ToMap(e.OnlyFields, true)
}

func (e Encoder[T]) withStruct(v reflect.Value, fn func(name string, tag xstruct.Tag, field reflect.StructField, value reflect.Value) error) error {
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("unsupported type %s, expect struct", v.Kind().String())
	}
	keys := make(map[string]struct{}, len(e.OnlyFields))
	fieldsLimit := e.sliceToMapTrue(e.OnlyFields)
	fieldsIgnore := e.sliceToMapTrue(e.IgnoreFields)
	tn := dbschema.TagName()
	err := zreflect.RangeStructFields(v.Type(), func(field reflect.StructField) error {
		// embed 类型的，详见 testUser3、testUser4
		if field.Anonymous {
			fv := v.FieldByName(field.Name)
			switch fv.Kind() {
			case reflect.Struct:
				return e.withStruct(fv, fn)
			case reflect.Pointer:
				return e.withStruct(fv.Elem(), fn)
			default:
				panic(fmt.Sprintf("what Anonymous %s: %v", fv.Kind(), fv))
			}
		}
		if !field.IsExported() {
			return nil
		}
		fv := v.FieldByName(field.Name)
		if !fv.CanInterface() {
			return nil
		}

		tag := xstruct.ParserTagCached(field.Tag, tn)
		name := tag.Name()
		if name == "-" || name == "" {
			return nil
		}
		if _, has := keys[name]; has {
			return fmt.Errorf("duplicate field %q", name)
		}

		if len(fieldsLimit) > 0 && !fieldsLimit[name] {
			return nil
		}
		if len(fieldsIgnore) > 0 && fieldsIgnore[name] {
			return nil
		}
		if dbschema.TagHasAutoInc(tag) && fv.IsZero() {
			// 当是自增长类型、并且是零值，则跳过该字段
			return nil
		}
		if err := fn(name, tag, field, fv); err != nil {
			return err
		}
		keys[name] = struct{}{}
		return nil
	})
	return err
}

// encodeMap 处理 map[string]any
func (e Encoder[T]) encodeMap(v reflect.Value) (map[string]any, error) {
	fieldsLimit := e.sliceToMapTrue(e.OnlyFields)
	fieldsIgnore := e.sliceToMapTrue(e.IgnoreFields)
	result := make(map[string]any)
	for _, k := range v.MapKeys() {
		val := v.MapIndex(k).Interface()
		if k.Kind() != reflect.String {
			return nil, fmt.Errorf("key %#v is not a string", val)
		}
		key := k.String()
		if len(fieldsLimit) > 0 && !fieldsLimit[key] {
			continue
		}
		if len(fieldsIgnore) > 0 && fieldsIgnore[key] {
			continue
		}
		result[key] = val
	}
	return result, nil
}

// Fields 获取 data 的字段列表（未过滤的）。
//
//  1. 当类型是 struct 或者 *struct 的时候，返回所有有效的 db tag 的字段
//  2. 当类型是 map[string]any 时，返回 map keys, nil
//  3. 其他类型，返回 error
func (e Encoder[T]) Fields(data T) ([]string, error) {
	sc, _ := dbschema.Schema(e.Dialect, data)
	if sc != nil {
		return sc.ColumnNames, nil
	}

	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil, fmt.Errorf("encoder.Fields with invalid value: %#v", v)
	}

	// 支持指针类型
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, fmt.Errorf("encoder.Fields with nil pointer: %#v", data)
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Map {
		return nil, fmt.Errorf("encoder.Fields with invalid type: %T", data)
	}

	result := make([]string, 0, v.Len())
	for _, k := range v.MapKeys() {
		val := v.MapIndex(k).Interface()
		if k.Kind() != reflect.String {
			return nil, fmt.Errorf("key %#v is not a string", val)
		}
		result = append(result, k.String())
	}
	return result, nil
}

// encodeStructFieldValue 对单个字段根据类型和 serializer 转换
func (e Encoder[T]) encodeStructFieldValue(schema dbtype.ColumnSchema, val any) (any, error) {
	if schema.Auto != "" {
		if e.Action.IsInsert() {
			if nv, ok := insertAutoFns.do(schema, val); ok {
				val = nv
			} else if nv, ok = updateAutoFns.do(schema, val); ok {
				val = nv
			}
		} else if e.Action.IsUpdate() {
			if nv, ok := updateAutoFns.do(schema, val); ok {
				val = nv
			}
		}
	}

	rv := reflect.ValueOf(val)
	if !rv.IsValid() {
		return nil, fmt.Errorf("invalid value: %v", val)
	}
	// 依据 schema.NotNull 对 *nil ptr 的处理规则：
	// 若是 insert：NotNull==true，将 nil 转换为空值
	// 若是 update：NotNull==true，值为 nil 则忽略

	// 处理指针
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			if e.Action.IsInsertOrUpdate() && schema.NotNull {
				return nil, xerror.SkipOne
			}
			rv = reflect.New(rv.Type().Elem()).Elem()
			val = rv.Interface()
		} else {
			rv = rv.Elem()
			val = rv.Interface()
		}
	}

	if e.Action.IsInsertOrUpdate() && schema.NotNull && rv.Kind() == reflect.Slice && rv.IsNil() {
		return nil, xerror.SkipOne
	}

	// 类型的判断处理应该有 schema parser 处理好，传入正确的 Codec 即可
	if schema.Codec != nil {
		return schema.Codec.Encode(val)
	}
	return val, nil
}

func (e Encoder[T]) PKNameAndValues(obj T) (map[string]any, error) {
	cols, values, err := e.PrimaryKeys(obj)
	if err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, dbtype.ErrNoPK
	}
	result := make(map[string]any, len(cols))
	for i, col := range cols {
		value, err1 := e.encodeStructFieldValue(col, values[i].Interface())
		if err1 != nil {
			return nil, err1
		}
		result[col.Name] = value
	}
	return result, nil
}

func (e Encoder[T]) PrimaryKeys(obj T) (columns []dbtype.ColumnSchema, values []reflect.Value, err error) {
	v := reflect.ValueOf(obj)
	if !v.IsValid() {
		return nil, nil, fmt.Errorf("invalid value: %v", v)
	}

	// 支持指针类型
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, nil, fmt.Errorf("nil pointer: %#v", obj)
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("invalid value: %#v", obj)
	}

	schema, err := dbschema.Schema(e.Dialect, obj)
	if err != nil {
		return nil, nil, err
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !v.Field(i).CanInterface() {
			continue
		}

		tag := xstruct.ParserTag(field.Tag.Get(dbschema.TagName()))
		name := tag.Name()
		if name == "-" || name == "" {
			continue
		}
		col, err := schema.ColumnByName(name)
		if err != nil {
			return nil, nil, err
		}
		if !col.IsPrimaryKey {
			continue
		}
		// columns 和 values 同时 append
		columns = append(columns, col)
		values = append(values, v.Field(i))
	}

	return columns, values, nil
}
