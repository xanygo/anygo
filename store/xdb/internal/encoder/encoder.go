//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package encoder

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/xerror"
)

// Encoder 将数据 T 编码为 可以直接用于 sql 语句的值
// T 应该是 struct 或者 *struct
type Encoder[T any] struct {
	Schema       *dbtype.TableSchema // 必填
	Action       Action              // 必填
	Dialect      dbtype.Dialect      // 方言，必填
	OnlyFields   []string            // 当不为空时，输出结果的 key 只限定此列表中的
	IgnoreFields []string            // 当不为空时，输出结果的 key 限定不能是此列表中的
}

func (e Encoder[T]) WithAction(action Action) Encoder[T] {
	return Encoder[T]{
		Schema:       e.Schema,
		Action:       action,
		Dialect:      e.Dialect,
		OnlyFields:   e.OnlyFields,
		IgnoreFields: e.IgnoreFields,
	}
}

// Encode 将 struct 或者 *struct 编码为数据
func (e Encoder[T]) Encode(data T) (map[string]any, error) {
	v, err := e.reflectValue(data)
	if err != nil {
		return nil, err
	}
	return e.encodeStruct(v)
}

func (e Encoder[T]) reflectValue(data T) (reflect.Value, error) {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return reflect.Value{}, fmt.Errorf("invalid value: %v", v)
	}

	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}, fmt.Errorf("nil pointer: %#v", data)
		}
		v = v.Elem()
	}
	return v, nil
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
func (e Encoder[T]) encodeStruct(v reflect.Value) (map[string]any, error) {
	result := make(map[string]any, len(e.OnlyFields))
	err := e.rangeStructFieldsWithFilter(v, func(fieldSchema dbtype.ColumnSchema, value reflect.Value) error {
		if fieldSchema.AutoIncrement {
			if e.Action.IsInsert() && value.IsZero() {
				// 当是自增长类型、并且是零值，则跳过该字段
				return nil
			}

			if e.Action.IsUpdate() {
				// 更新的时候，将自增长类型的字段忽略掉
				// 如 sql server，在更新的字段列表中，包含【自增长】类型的字段，
				// 将报错：mssql: Cannot update identity column 'id'
				return nil
			}
		}
		name := fieldSchema.Name
		encodedVal, err := e.encodeStructFieldValue(fieldSchema, value.Interface())
		if err != nil {
			if errors.Is(err, xerror.SkipOne) {
				return nil
			}
			return fmt.Errorf("encode field %q: %w", name, err)
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
	return xslice.ToMap(s, true)
}

// rangeStructFields 遍历 value 里所有的字段，同时剔除过滤字段
func (e Encoder[T]) rangeStructFieldsWithFilter(v reflect.Value, fn func(fieldSchema dbtype.ColumnSchema, value reflect.Value) error) error {
	fieldsLimit := e.sliceToMapTrue(e.OnlyFields)
	fieldsIgnore := e.sliceToMapTrue(e.IgnoreFields)

	return e.rangeStructFields(v, func(fieldSchema dbtype.ColumnSchema, value reflect.Value) error {
		name := fieldSchema.Name
		if len(fieldsLimit) > 0 && !fieldsLimit[name] {
			return nil
		}
		if len(fieldsIgnore) > 0 && fieldsIgnore[name] {
			return nil
		}
		return fn(fieldSchema, value)
	})
}

// rangeStructFields 遍历 value 里所有的字段
func (e Encoder[T]) rangeStructFields(v reflect.Value, fn func(fieldSchema dbtype.ColumnSchema, value reflect.Value) error) error {
	// 不需要检查 v 的类型是 struct, RangeStructFields 会检查
	return zreflect.RangeStructFields(v.Type(), func(field reflect.StructField) error {
		// embed 类型的，详见 testUser3、testUser4
		if field.Anonymous {
			fieldValue := v.FieldByName(field.Name)
			switch fieldValue.Kind() {
			case reflect.Struct:
				return e.rangeStructFields(fieldValue, fn)
			case reflect.Pointer:
				return e.rangeStructFields(fieldValue.Elem(), fn)
			default:
				// 理论上不会出现
				return fmt.Errorf("invalid Anonymous %s: %v", fieldValue.Kind(), fieldValue)
			}
		}

		if !field.IsExported() {
			return nil
		}
		fieldValue := v.FieldByName(field.Name)
		if !fieldValue.CanInterface() {
			return nil
		}

		tag := xstruct.ParserTagCached(field.Tag, e.Schema.TagName)
		name := tag.Name()
		if name == "-" || name == "" {
			return nil
		}

		fieldSchema, err := e.Schema.ColumnByName(name)
		if err != nil {
			return err
		}
		return fn(fieldSchema, fieldValue)
	})
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
	for rv.Kind() == reflect.Pointer {
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
	v, err := e.reflectValue(obj)
	if err != nil {
		return nil, nil, err
	}
	err = e.rangeStructFields(v, func(fieldSchema dbtype.ColumnSchema, value reflect.Value) error {
		if fieldSchema.IsPrimaryKey {
			columns = append(columns, fieldSchema)
			values = append(values, value)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return columns, values, nil
}

// EncodeArgs 对 where 的参数编码
func (e Encoder[T]) EncodeArgs(args ...any) ([]any, error) {
	if len(args) == 0 {
		return nil, nil
	}
	result := make([]any, len(args))
	for i, arg := range args {
		if na, ok := arg.(sql.NamedArg); ok {
			v, err := e.Dialect.EncodeValue(na.Value)
			if err != nil {
				return nil, fmt.Errorf("encode args %#v: %w", arg, err)
			}
			na.Value = v
			result[i] = na
			continue
		}
		val, err := e.Dialect.EncodeValue(arg)
		if err != nil {
			return nil, fmt.Errorf("encode args %#v: %w", arg, err)
		}
		result[i] = val
	}
	return result, nil
}
