//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-07

package xi18n

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/xanygo/anygo/ds/xmap"
)

// TemplateRender 渲染模版的辅助类
type TemplateRender struct {
	Bundle    *Bundle    // 语言资源包
	Languages []Language // 有 http header accept-language 解析出的首选语言列表
}

// RA 在模版中加载渲染本地化内容
//
//	如  hello {{ .xi.RA "index@k1" }} 、 {{ .xi.RA "index@k1" arg1 arg2 }}
//	第一个参数 "index@k1" 中 namespace = "index", key="k1"
//	arg1 arg2 等是可选参数，支持 >=0 个
//
// 由于 xi 是通过 namespace + key 读取本地化内容的，所以要求所有的本地化资源都在 Bundle 中有定义
func (tr *TemplateRender) RA(key string, args ...any) (string, error) {
	return renderA(tr.Bundle, tr.Languages, key, args...)
}

func renderA(b *Bundle, ls []Language, key string, args ...any) (string, error) {
	msg := FindMessage(b, ls, "", key)
	if msg == nil {
		return "", fmt.Errorf("cannot find %q", key)
	}
	return msg.Render(args...)
}

// RB 渲染本地化内容,若没有对应的语言内容，则使用传入的 text 渲染
//
//	如 {{ .xi.RB "你好" "index@k1" }} 或者 {{ .xi.RB "你好 {0}" "index@k1" "demo" }}
//	在 “|” 前的内容是预定义的本地化模版信息,本地化信息中的变量使用 {number} 作为占位符，从 0 依次递增
func (tr *TemplateRender) RB(text string, key string, args ...any) (string, error) {
	return renderB(tr.Bundle, tr.Languages, text, key, args...)
}

func renderB(b *Bundle, ls []Language, text string, key string, args ...any) (string, error) {
	if len(ls) > 1 {
		ls = ls[:1] // 只查找第一候选语言
	}
	msg := FindMessage(b, ls, "", key)
	if msg != nil {
		return msg.Render(args...)
	}
	return renderMsgSlice(text, args...)
}

// Is 是否首选语言
func (tr *TemplateRender) Is(lang string) bool {
	return len(tr.Languages) > 0 && tr.Languages[0] == Language(lang)
}

func (tr *TemplateRender) BindXI(b *Bundle, languages []Language, namespace string) func(key string, args ...any) (string, error) {
	return func(key string, args ...any) (string, error) {
		msg := FindMessage(b, languages, namespace, key)
		if msg == nil {
			return "", fmt.Errorf("cannot find i18n key %q", key)
		}
		return msg.Render(args...)
	}
}

func (tr *TemplateRender) BindXTT(b *Bundle, languages []Language, namespace string) func(key string, args ...any) (string, error) {
	return func(key string, args ...any) (string, error) {
		var ok = len(args) > 0
		var text string
		if ok {
			text, ok = args[len(args)-1].(string)
		}
		if !ok {
			return "", fmt.Errorf("i18n key=%q, missing text", key)
		}
		if namespace != "" {
			key = namespace + "@" + key
		}
		return renderB(b, languages, text, key, args[:len(args)-1]...)
	}
}

func (tr *TemplateRender) BindIs(languages []Language) func(lang string) bool {
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
	var rd TemplateRender
	return map[string]any{
		"xi":    rd.BindXI(b, languages, namespace),
		"xi_is": rd.BindIs(languages),
		"xit":   rd.BindXTT(b, languages, namespace),
	}
}

var errNoBundleInCtx = errors.New("not found Bundle in context, should ContextWithBundle first")

// RA 使用资源的 key 渲染文本内容,需要提前使用 ContextWithBundle 将 *Bundle 存入 ctx。
// 若是 Bundle 不存在会 panic，key 不存在会返回错误信息。
func RA(ctx context.Context, key string, args ...any) (string, error) {
	rr, ok := ctx.Value(ctxKeyBundle).(*ctxBundle)
	if !ok {
		return "", errNoBundleInCtx
	}
	if rr.namespace != "" {
		key = rr.namespace + "@" + key
	}
	return renderA(rr.bundle, LanguagesFromContext(ctx), key, args...)
}

// RB 使用资源 key 渲染文本内容,若 key 不存在则使用传入的模版( 参数名: text，作为默认模版 ) 渲染
// 需要提前使用 ContextWithBundle 将 *Bundle 和 首选语言 Language 存入 ctx。
// 若是 Bundle 不存在会 panic，key 不存在会返回错误信息。
func RB(ctx context.Context, text string, key string, args ...any) (string, error) {
	rr, ok := ctx.Value(ctxKeyBundle).(*ctxBundle)
	if !ok {
		return "", errNoBundleInCtx
	}
	languages := LanguagesFromContext(ctx)
	if rr.namespace != "" {
		key = rr.namespace + "@" + key
	}
	return renderB(rr.bundle, languages, text, key, args...)
}

func RC(b *Bundle, req *http.Request, key string, args ...any) (string, error) {
	languages := HTTPHandler{}.Languages(req)
	return renderA(b, languages, key, args...)
}

func RD(b *Bundle, req *http.Request, text string, key string, args ...any) (string, error) {
	languages := HTTPHandler{}.Languages(req)
	return renderB(b, languages, text, key, args...)
}

// BuildTemplate 构建支持多语言支持的模版集合
//
//	在模版文件中，可以使用 xi、xit 函数来读取语言资源
//
// 基本原理：
//
//	给 Bundle 中已定义的语言类型(Language,如 zh，en 等)，每个语言各自编译一套模版文件（将 语言资源和此模版绑定），返回一个以 Language 为 key 的 Map。
//	之后可以在渲染时，依据 http.Request 的 Accept-Language （用户浏览器配置的首选语言列表），从 Map 中筛选出对应的模版文件，然后渲染。
func BuildTemplate(defaultLang Language, b *Bundle, build func(t *template.Template) *template.Template) *xmap.Tags[Language, *template.Template, Language] {
	mt := &xmap.Tags[Language, *template.Template, Language]{}
	var langs []Language
	if b != nil {
		langs = b.Languages()
	}
	if len(langs) == 0 {
		langs = []Language{defaultLang}
	}
	for _, lang := range langs {
		prefer := []Language{lang} // 当前语言优先级最高，其他语言作为 backup
		for _, a := range langs {
			if a != lang {
				prefer = append(prefer, a)
			}
		}
		tpl := template.New("i18n_layout").Funcs(FuncMap(b, prefer, ""))
		tpl = build(tpl)
		mt.Set(lang, tpl, lang)
	}
	return mt
}
