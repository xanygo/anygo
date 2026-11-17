package xvld

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/xanygo/anygo/internal/xstruct"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xi18n"
	"github.com/xanygo/anygo/xvld/internal"
	"github.com/xanygo/anygo/xvld/internal/buildin"
)

type (
	FieldContext = internal.FieldContext

	Rule = internal.Rule

	ValidationFunc = internal.ValidationFunc
)

var _ StructValidator = (*Instance)(nil)
var _ VarValidator = (*Instance)(nil)

type Instance struct {
	validationFuncs map[string]ValidationFunc // 自定义校验规则
	validatonErrs   map[string]*xi18n.Error   // 自定义校验规则的错误
	languages       []xi18n.Language          // 首选语言列表
	tagName         string
}

func (ins *Instance) Clone() *Instance {
	return &Instance{
		validationFuncs: maps.Clone(ins.validationFuncs),
		validatonErrs:   maps.Clone(ins.validatonErrs),
		languages:       slices.Clone(ins.languages),
	}
}

// SetLanguages 设置首选语言列表
func (ins *Instance) SetLanguages(langs ...xi18n.Language) {
	ins.languages = langs
}

func (ins *Instance) SetTagName(name string) {
	ins.tagName = name
}

func (ins *Instance) MustRegisterRule(name string, fn ValidationFunc, err *xi18n.Error) {
	ne := ins.RegisterRule(name, fn, err)
	if ne != nil {
		panic(ne)
	}
}

func (ins *Instance) RegisterRule(name string, fn ValidationFunc, err *xi18n.Error) error {
	if ins.validationFuncs == nil {
		ins.validationFuncs = make(map[string]ValidationFunc, 18)
		ins.validatonErrs = make(map[string]*xi18n.Error, 18)
	}
	if name == "" || fn == nil || err == nil {
		return fmt.Errorf("cannot RegisterRule name=%q with empty params", name)
	}
	if len(err.Messages) == 0 {
		return fmt.Errorf("cannot RegisterRule name=%q with empty err.Messages", name)
	}
	if ins.exists(name) {
		return fmt.Errorf("cannot register %q alreay exists", name)
	}
	return nil
}

func (ins *Instance) exists(name string) bool {
	if _, ok := buildin.Rules[name]; ok {
		return true
	}
	if len(ins.validationFuncs) > 0 {
		_, ok := ins.validationFuncs[name]
		return ok
	}
	return false
}

func (ins *Instance) StructContext(ctx context.Context, v any) error {
	rv := reflect.ValueOf(v)
	rv = zreflect.IndirectValue(rv)
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return internal.ErrNotStruct.WithLangParams(ins.languages, []any{fmt.Sprintf("%T", v)})
	}

	langs := ins.getLangs(ctx)

	var err error
	zreflect.RangeStructValue(rv, func(field reflect.StructField, value reflect.Value) bool {
		tag := xstruct.CachedTag(field.Tag, ins.getTagName())
		if tag.Empty() {
			return true
		}
		before, after := ins.parserRules(tag)
		newErr := ins.newErr(field.Name, langs)
		if len(before) > 0 {
			fc := internal.NewFieldContext(field, rv, rv, value, newErr)
			err = ins.validateField(fc, before)
		}

		if err == nil && len(after) > 0 {
			fc := internal.NewFieldContext(field, rv, rv, value, newErr)
			err = ins.dive(fc, after)
		}
		return err == nil
	})
	return err
}

func (ins *Instance) dive(fc *FieldContext, rules []Rule) error {
	rv := zreflect.IndirectValue(fc.Value())
	if !rv.IsValid() {
		return fc.UnsupportedType(rules[0])
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		if rv.Len() == 0 {
			return nil
		}
		for i := 0; i < rv.Len(); i++ {
			v := rv.Index(i)
			if err := ins.validateField(fc.Child(v), rules); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		if rv.Len() == 0 {
			return nil
		}
		for _, k := range rv.MapKeys() {
			v := rv.MapIndex(k)
			if err := ins.validateField(fc.Child(v), rules); err != nil {
				return err
			}
		}
		return nil
	default:
		return fc.UnsupportedType(rules[0])
	}
}

func (ins *Instance) Struct(v any) error {
	return ins.StructContext(context.Background(), v)
}

func (ins *Instance) VarContext(ctx context.Context, val any, rules ...Rule) error {
	if len(rules) == 0 {
		return nil
	}
	value := reflect.ValueOf(val)
	langs := ins.getLangs(ctx)
	newErr := func(name string, params ...any) error {
		return ins.newDetailErr(name, langs, params...)
	}
	fc := internal.NewFieldContext(reflect.StructField{}, reflect.Value{}, reflect.Value{}, value, newErr)
	return ins.validateField(fc, rules)
}

func (ins *Instance) Var(val any, rules ...Rule) error {
	return ins.VarContext(context.Background(), val, rules...)
}

func (ins *Instance) getLangs(ctx context.Context) []xi18n.Language {
	langs := xi18n.LanguagesFromContext(ctx)
	if len(langs) > 0 {
		return langs
	}
	return ins.languages
}

func (ins *Instance) newErr(fieldName string, langs []xi18n.Language) func(name string, params ...any) error {
	return func(name string, params ...any) error {
		e0 := internal.ErrBase.WithLangParams(langs, []any{fieldName}).Error()
		e1 := ins.newDetailErr(name, langs, params...)
		return fmt.Errorf(e0, e1)
	}
}

func (ins *Instance) newDetailErr(ruleName string, langs []xi18n.Language, params ...any) error {
	e, ok := ins.validatonErrs[ruleName]
	if ok {
		return e.WithLangParams(langs, params)
	}
	e, ok = buildin.Errs[ruleName]
	if ok {
		return e.WithLangParams(langs, params)
	}
	return internal.ErrDefault.WithLangParams(langs, []any{ruleName})
}

func (ins *Instance) parserRules(tag xstruct.Tag) (before []Rule, after []Rule) {
	var isDive bool
	for k, v := range tag.Iter() {
		if !isDive && (k == "dive" || k == "/") {
			isDive = true
		}
		if isDive {
			after = append(after, Rule{Name: k, Param: v})
		} else {
			before = append(before, Rule{Name: k, Param: v})
		}
	}
	return before, after
}

func (ins *Instance) getTagName() string {
	if ins.tagName != "" {
		return ins.tagName
	}
	return "xv"
}

// validateField 校验字段
func (ins *Instance) validateField(fc *FieldContext, rules []Rule) error {
	for _, rule := range rules {
		// skipmpty：空值时跳过后续所有规则
		if rule.Name == "skipempty" && zreflect.IsEmptyValue(fc.Value()) {
			break
		}
		if err := ins.one(fc, rule); err != nil {
			return err
		}
	}
	return nil
}

func (ins *Instance) one(fc *FieldContext, rule Rule) error {
	fn, ok := buildin.Rules[rule.Name]
	if ok {
		return fn(fc, rule)
	}
	if len(ins.validationFuncs) > 0 {
		fn, ok = ins.validationFuncs[rule.Name]
		if ok {
			return fn(fc, rule)
		}
	}
	return fmt.Errorf("xvld rule=%q %w", rule.Name, xerror.NotFound)
}
