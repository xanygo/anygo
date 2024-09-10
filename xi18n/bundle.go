//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-05

package xi18n

// Bundle 用于存储所有本地化信息的组件，以支持的语言 ( Language )  为 key 存储和查询
//
// 比如在 Bundle 中，会同时存储 zh（中文-Chinese）、en （英文-English）的本地化信息
type Bundle struct {
	// localizes 所有本地化信息
	// map  key- 时语言类型，value 时本地化语言信息
	localizes map[Language]*Localize

	languages []Language
}

// MustLocalize 查找指定语言的本地化配置，若不存在会创建
func (b *Bundle) MustLocalize(lang Language) *Localize {
	if b.localizes == nil {
		b.localizes = make(map[Language]*Localize)
	}
	if b.localizes[lang] == nil {
		b.localizes[lang] = &Localize{}
		b.languages = append(b.languages, lang)
	}
	return b.localizes[lang]
}

// Localize 查找指定语言的本地化配置，若不存在会返回 nil
func (b *Bundle) Localize(lang Language) *Localize {
	if b.localizes == nil {
		return nil
	}
	return b.localizes[lang]
}

// Languages 所有支持的本地化语言，返回的 slice 是排序的，先通过 MustLocalize 注册的会排在前面
func (b *Bundle) Languages() []Language {
	return b.languages
}
