//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-05

package xi18n

type Bundle struct {
	// localizes 所有本地化信息
	// map  key- 时语言类型，value 时本地化语言信息
	localizes map[Language]*Localize
}

func (b *Bundle) MustLocalize(lang Language) *Localize {
	if b.localizes == nil {
		b.localizes = make(map[Language]*Localize)
	}
	if b.localizes[lang] == nil {
		b.localizes[lang] = &Localize{}
	}
	return b.localizes[lang]
}

func (b *Bundle) Localize(lang Language) *Localize {
	if b.localizes == nil {
		return nil
	}
	return b.localizes[lang]
}
