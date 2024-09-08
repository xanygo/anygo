//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-06

package xi18n

import "strings"

type Language string

const (
	LangZh Language = "zh"
	LangEn Language = "en"
)

func ParserAccept(accept string) []Language {
	arr := strings.Split(accept, ",")
	result := make([]Language, 0, len(arr))
	for _, v := range arr {
		v = strings.TrimSpace(v)
		b, _, _ := strings.Cut(v, ";")
		b = strings.TrimSpace(b)
		if b != "" {
			result = append(result, Language(b))
		}
	}
	return result
}
