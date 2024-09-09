//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-07

package xi18n

import (
	"fmt"
)

type Render struct{}

func (r Render) BindXI(b *Bundle, languages []Language, ns string) func(key string, args ...any) string {
	return func(key string, args ...any) string {
		msg := FindMessage(b, languages, ns, key)
		if msg == nil {
			return "cannot find " + key
		}
		return r.result(msg.RenderSlice(args...))
	}
}

func (r Render) result(result string, err error) string {
	if err == nil {
		return result
	}
	return fmt.Sprintf("render i18n message %s", err.Error())
}

func (r Render) BindXTT(b *Bundle, languages []Language, ns string) func(key string, args ...any) string {
	useDefault := len(languages) == 0
	if !useDefault {
		if bls := b.Languages(); len(bls) > 0 {
			useDefault = bls[0] == languages[0]
		}
	}
	return func(key string, args ...any) string {
		var ok = len(args) > 0
		var text string
		if ok {
			text, ok = args[len(args)-1].(string)
		}
		if !ok {
			return fmt.Sprintf("key=%q, missing text", key)
		}
		if useDefault {
			return r.result(renderMsgSlice(text, args[0:len(args)-1]...))
		}
		msg := FindMessage(b, languages, ns, key)
		if msg == nil {
			return r.result(renderMsgSlice(text, args[0:len(args)-1]...))
		}
		return r.result(msg.RenderSlice(args[0 : len(args)-1]...))
	}
}

func (r Render) BindIs(languages []Language) func(lang string) bool {
	return func(lang string) bool {
		return len(languages) > 0 && languages[0] == Language(lang)
	}
}

func FuncMap(b *Bundle, languages []Language, ns string) map[string]any {
	var rd Render
	return map[string]any{
		"xi":    rd.BindXI(b, languages, ns),
		"xi_is": rd.BindIs(languages),
		"xit":   rd.BindXTT(b, languages, ns),
	}
}
