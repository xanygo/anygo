//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml_test

import (
	"fmt"

	"github.com/xanygo/anygo/xhtml"
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
