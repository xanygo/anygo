//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml_test

import (
	"embed"
	"fmt"
	"html/template"
	"maps"
	"net/http"

	"github.com/xanygo/anygo/xhtml"
	"github.com/xanygo/anygo/xi18n"
)

func ExampleHTMLBytes_Div() {
	code := xhtml.HTMLBytes("hello")
	got, _ := code.Div()
	fmt.Println(string(got))

	// Output:
	// <div>hello</div>
}

func ExampleTextStringSlice_UL() {
	values := xhtml.TextStringSlice{"1", "2", "3"}
	got, _ := values.UL()
	fmt.Println(string(got))
	// Output:
	// <ul><li>1</li><li>2</li><li>3</li></ul>
}

func ExampleNewHTML() {
	h := xhtml.NewHTML()
	xhtml.AddTo(h,
		xhtml.WithAny(xhtml.NewHead(), func(a *xhtml.Any) {
			xhtml.AddTo(a, xhtml.NewTitle(xhtml.TextString("hello")))
		}),
		xhtml.WithAny(xhtml.NewBody(), func(a *xhtml.Any) {
			xhtml.AddTo(a, xhtml.TextString("Hello World"))
		}),
	)
	got, _ := h.HTML()
	fmt.Println(string(got))
	// Output:
	// <html><head><title>hello</title></head><body>Hello World</body></html>
}

func ExampleNewOption() {
	o1 := xhtml.NewOption("1", xhtml.TextString("class 1"))
	got, _ := o1.HTML()
	fmt.Println(string(got)) // <option value="1">class 1</option>

	o2 := xhtml.NewOption("1", nil)
	got2, _ := o2.HTML()
	fmt.Println(string(got2)) // <option value="1">1</option>

	// Output:
	// <option value="1">class 1</option>
	// <option value="1">1</option>
}

func ExampleTemplateEngine_i18n() {
	// 演示：使用 i18n 渲染模板

	funcmap := template.FuncMap{}
	files := embed.FS{}

	maps.Copy(funcmap, xhtml.FuncMap)
	ib := xi18n.NewBundle(xi18n.LangZh, xi18n.LangEn)
	// 这里略掉了将 language 资源信息加载到 ib 的逻辑

	it := xi18n.BuildTemplate(xi18n.LangZh, ib, func(t *template.Template) *template.Template {
		return template.Must(xhtml.WalkParseFS(t.Funcs(funcmap), files, "files", "*.html"))
	})

	te := &xhtml.TemplateEngine{
		TemplateResolver: func(r *http.Request) xhtml.TemplateExecutor {
			// 需要给 xhttp.Router 注册中间件  xi18n.HTTPHandler
			languages := xi18n.LanguagesFromContext(r.Context())
			return it.Any(xi18n.LangZh, languages...)
		},
	}

	_ = te
}
