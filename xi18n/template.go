//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-07

package xi18n

import (
	"fmt"
)

type Render struct{}

func (r Render) BindXI(b *Bundle, languages []Language, namespace string) func(key string, args ...any) string {
	return func(key string, args ...any) string {
		msg := FindMessage(b, languages, namespace, key)
		if msg == nil {
			return "cannot find " + key
		}
		return r.result(msg.Render(args...))
	}
}

func (r Render) result(result string, err error) string {
	if err == nil {
		return result
	}
	return fmt.Sprintf("render i18n message %s", err.Error())
}

func (r Render) BindXTT(b *Bundle, languages []Language, namespace string) func(key string, args ...any) string {
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
		msg := FindMessage(b, languages, namespace, key)
		if msg == nil {
			return r.result(renderMsgSlice(text, args[0:len(args)-1]...))
		}
		return r.result(msg.Render(args[0 : len(args)-1]...))
	}
}

func (r Render) BindIs(languages []Language) func(lang string) bool {
	return func(lang string) bool {
		return len(languages) > 0 && languages[0] == Language(lang)
	}
}

// FuncMap 返回用于注册到 text/template 和 html/template 的 FuncMap。
// 为了简化 template 的使用，应该对每一个支持的语言，创建一个对于的 template 模版，
// 并且注册此 FuncMap。
// 在使用的时候，从 HTTP Header 的 Accept-Language 字段读取支持的语言，然后查找到对应的模版用于渲染。
//
// # 参数说明：
//
//	*Bundle: 本地化资源集
//	[]Language: 优先查找的语言,如初始化用于支持中文的的模版，则此值可以是 []Language{ xi18n.LangZh },
//	 初始化用于支持英文的的模版，则此值可以是 []Language{ xi18n.LangEn }
//
// # 包含模版函数：
//
//  1. xi: 在模版中加载渲染本地化内容
//
//     如  hello {{ xi "index@k1" }} 、 {{ xi "index@k1" arg1 arg2 }}
//     第一个参数 "index@k1" 中 namespace = "index", key="k1"
//     arg1 arg2 等是可选参数，支持 >=0 个
//
// 由于 xi 是通过 namespace + key 读取本地化内容的，所以要求所有的本地化资源都在 Bundle 中有定义
//
// 2. xit: 优先使用预定义本地化信息，并渲染本地化内容
//
//	如 {{ "你好" | xit "index@k1" }} 或者 {{ "你好 {0}" | xit "index@k1" "demo" }}
//	在 “|” 前的内容是预定义的本地化模版信息,本地化信息中的变量使用 {number} 作为占位符，从 0 依次递增
func FuncMap(b *Bundle, languages []Language, namespace string) map[string]any {
	var rd Render
	return map[string]any{
		"xi":    rd.BindXI(b, languages, namespace),
		"xi_is": rd.BindIs(languages),
		"xit":   rd.BindXTT(b, languages, namespace),
	}
}
