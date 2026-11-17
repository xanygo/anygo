package internal

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/internal/zreflect"
)

var ErrParamEmpty = errors.New("param is empty")

type ValidationFunc func(fl *FieldContext, rule Rule) error

type NewErrFunc func(name string, params ...any) error

type Rule struct {
	Name  string
	Param string
}

func (rule Rule) String() string {
	if rule.Param == "" {
		return rule.Name
	}
	return fmt.Sprintf("%s=%q", rule.Name, rule.Param)
}

func NewFieldContext(field reflect.StructField, top, parent, current reflect.Value, newErr NewErrFunc) *FieldContext {
	return &FieldContext{
		fieldName:   field.Name,
		value:       zreflect.IndirectValue(current),
		parent:      zreflect.IndirectValue(parent),
		top:         zreflect.IndirectValue(top),
		structField: field,
		newErr:      newErr,
	}
}

// FieldContext 字段校验上下文
type FieldContext struct {
	// 当前字段名称
	fieldName string

	// 当前待校验字段
	value reflect.Value

	// 当前字段所属结构体
	parent reflect.Value

	// 最外层结构体
	top reflect.Value

	// 字段定义
	structField reflect.StructField

	newErr func(name string, params ...any) error
}

// FieldName 获取字段名称
func (fc *FieldContext) FieldName() string {
	return fc.fieldName
}

// Value 获取当前待校验字段的值
func (fc *FieldContext) Value() reflect.Value {
	return fc.value
}

// Parent 获取当前字段所属结构体
func (fc *FieldContext) Parent() reflect.Value {
	return fc.parent
}

// Top 获取最外层结构体
func (fc *FieldContext) Top() reflect.Value {
	return fc.top
}

// StructField 获取字段定义
func (fc *FieldContext) StructField() reflect.StructField {
	return fc.structField
}

// ExtractField 获取指定字段的值
func (fc *FieldContext) ExtractField(name string) reflect.Value {
	parent := fc.parent
	if !parent.IsValid() || parent.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	return parent.FieldByName(name)
}

// Error 创建校验错误
// 传参数约定：预期值在前，实际值在后
func (fc *FieldContext) Error(name string, params ...any) error {
	return fc.newErr(name, params...)
}

const (
	KeyInvalidParam  = "InvalidParam"
	KeyUnsupportType = "UnsupportType"
)

// InvalidParam 校验规则定义错误,如 定义 max=abc (预期应该是数字)
func (fc *FieldContext) InvalidParam(rule Rule, err error) error {
	return fc.Error(KeyInvalidParam, rule.String())
}

// UnsupportedType 数据类型不支持该校验规则的错误
func (fc *FieldContext) UnsupportedType(rule Rule) error {
	return fc.Error(KeyUnsupportType, fc.Value().Kind(), rule.Name)
}

func (fc *FieldContext) Child(v reflect.Value) *FieldContext {
	return &FieldContext{
		fieldName:   fc.fieldName,
		value:       zreflect.IndirectValue(v),
		top:         fc.top,
		parent:      fc.value,
		structField: fc.structField,
		newErr:      fc.newErr,
	}
}
