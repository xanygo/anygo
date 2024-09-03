//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcfg

// AutoChecker 当配置解析完成后，用于自动校验，
// 这个方法是在 validator 校验完成之后才执行的
type AutoChecker interface {
	AutoCheck() error
}

// Validator 自动规则校验器
//
// 可以使用 github.com/go-playground/validator/v10
// 如下设置所有字段都是必填的：
//
//	type Address struct {
//		Street string `validator:"required"`
//		City   string `validator:"required"`
//		Planet string `validator:"required"`
//		Phone  string `validator:"required"`
//	}
type Validator interface {
	Validate(val any) error
}

var DefaultValidator Validator
