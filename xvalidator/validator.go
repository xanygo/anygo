//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-15

package xvalidator

// AutoChecker 当配置解析完成后，用于自动校验，
// 这个方法是在 validator 校验完成之后才执行的
type AutoChecker interface {
	AutoCheck() error
}

// Validator 自动规则校验器
type Validator interface {
	Validate(val any) error
}

// Default 默认的 Validator，为 nil，在使用前可以基于
// github.com/go-playground/validator/v10 初始化。
//
// 初始化之后，可以采用如下设置，以让所有字段都是必填的：
//
//	type Address struct {
//		Street string `validator:"required"`
//		City   string `validator:"required"`
//		Planet string `validator:"required"`
//		Phone  string `validator:"required"`
//	}
var Default Validator = &empty{}

var _ Validator = (*empty)(nil)

type empty struct{}

func (a empty) Validate(val any) error {
	return nil
}

func Validate(val any) error {
	if err := Default.Validate(val); err != nil {
		return err
	}
	if ac, ok := val.(AutoChecker); ok {
		return ac.AutoCheck()
	}
	return nil
}

func ValidateWith(v Validator, val any) error {
	if v != nil {
		if err := v.Validate(val); err != nil {
			return err
		}
	} else {
		if err := Default.Validate(val); err != nil {
			return err
		}
	}
	if ac, ok := val.(AutoChecker); ok {
		return ac.AutoCheck()
	}
	return nil
}
