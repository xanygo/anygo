//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml_test

import (
	"fmt"

	"github.com/xanygo/anygo/xhtml"
)

func ExampleTextStringSlice_UL() {
	values := xhtml.TextStringSlice{"1", "2", "3"}
	got, _ := values.UL()
	fmt.Println(string(got))
	// Output:
	// <ul><li>1</li><li>2</li><li>3</li></ul>
}

func ExampleNewHTML() {
	h := xhtml.NewHTML()
	xhtml.Add(h,
		xhtml.WithAny(xhtml.NewHead(), func(a *xhtml.Any) {
			xhtml.Add(a, xhtml.NewTitle(xhtml.TextString("hello")))
		}),
		xhtml.WithAny(xhtml.NewBody(), func(a *xhtml.Any) {
			xhtml.Add(a, xhtml.TextString("Hello World"))
		}),
	)
	got, _ := h.HTML()
	fmt.Println(string(got))
	// Output:
	// <html><head><title>hello</title></head><body>Hello World</body></html>
}
