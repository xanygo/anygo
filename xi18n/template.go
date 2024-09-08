//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-07

package xi18n

import (
	"fmt"
	"strings"
)

func Render(b *Bundle, languages []Language, ns string) func(key string, args ...any) string {
	return func(key string, args ...any) string {
		ns1, key1, found := strings.Cut(key, nameSpaceKeySep)
		if found {
			ns = ns1
			key = key1
		}
		msg := FindMessage(b, languages, ns, key)
		if msg == nil {
			return "cannot find i18n message " + ns + nameSpaceKeySep + key
		}
		result, err := msg.Render1(args)
		if err == nil {
			return result
		}
		return fmt.Sprintf("render i18n message %q#%q: %s", ns, key, err.Error())
	}
}

func FuncMap(b *Bundle, languages []Language, ns string) map[string]any {
	return map[string]any{
		"it": Render(b, languages, ns),
		"isLang": func(lang string) bool {
			return len(languages) > 0 && languages[0] == Language(lang)
		},
	}
}
