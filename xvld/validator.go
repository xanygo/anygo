//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-15

package xvld

import (
	"context"

	"github.com/xanygo/anygo/internal/zreflect"
)

// AutoChecker 当配置解析完成后，用于自动校验，
// 这个方法是在 validator 校验完成之后才执行的
type AutoChecker interface {
	AutoCheck() error
}

type StructValidator interface {
	Struct(val any) error
	StructContext(ctx context.Context, val any) error
}

type VarValidator interface {
	Var(val any, rules ...Rule) error
	VarContext(ctx context.Context, val any, rules ...Rule) error
}

// Default 默认的 Validator
//
//	type Address struct {
//		Street string `xv:"required"`
//		City   string `xv:"required"`
//		Planet string `xv:"required"`
//		Phone  string `xv:"required"`
//	}
var Default = &Instance{}

func StructContextWith(ctx context.Context, v StructValidator, val any) error {
	if v == nil {
		v = Default
	}
	if err := v.StructContext(ctx, val); err != nil {
		return err
	}
	if ac, ok := val.(AutoChecker); ok {
		return ac.AutoCheck()
	}
	return nil
}

func StructWith(v StructValidator, val any) error {
	return StructContextWith(context.Background(), v, val)
}

func StructContext(ctx context.Context, val any) error {
	return StructContextWith(ctx, Default, val)
}

func Struct(val any) error {
	return StructContext(context.Background(), val)
}

func VarContextWith(ctx context.Context, v VarValidator, val any, rules ...Rule) error {
	if v == nil {
		v = Default
	}
	if err := v.VarContext(ctx, val, rules...); err != nil {
		return err
	}
	if ac, ok := val.(AutoChecker); ok {
		return ac.AutoCheck()
	}
	return nil
}

// VarContext 使用指定的规则校验单个值
func VarContext(ctx context.Context, val any, rules ...Rule) error {
	return VarContextWith(ctx, Default, val, rules...)
}

func Var(val any, rules ...Rule) error {
	return VarContextWith(context.Background(), Default, val, rules...)
}

// ValidateContextWith 使用传入的 Validator ，若 val 是 struct 则校验，否则值校验 AutoChecker
func ValidateContextWith(ctx context.Context, v StructValidator, val any) error {
	if zreflect.IsStruct(val) {
		return StructContextWith(ctx, v, val)
	}
	if ac, ok := val.(AutoChecker); ok {
		return ac.AutoCheck()
	}
	return nil
}

func ValidateWith(v StructValidator, val any) error {
	return ValidateContextWith(context.Background(), v, val)
}

func ValidateContext(ctx context.Context, val any) error {
	return ValidateContextWith(context.Background(), Default, val)
}

func Validate(val any) error {
	return ValidateWith(Default, val)
}
