//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml_test

import (
	"fmt"

	"github.com/xanygo/anygo/xhtml"
)

func ExampleNewUl() {
	values := []string{"1", "2", "3"}
	ul := xhtml.NewUl(values)
	got, _ := ul.HTML()
	fmt.Println(string(got))
	// Output:
	// <ul><li>1</li><li>2</li><li>3</li></ul>
}

func ExampleNewOl() {
	values := []string{"1", "2", "3"}
	ul := xhtml.NewOl(values)
	style := &xhtml.StyleAttr{}
	_ = style.Width("180px").Height("20px").SetTo(ul)

	got, _ := ul.HTML()
	fmt.Println(string(got))
	// Output:
	// <ol style="width:180px; height:20px"><li>1</li><li>2</li><li>3</li></ol>
}

func ExampleNewHTML() {
	h := xhtml.NewHTML()
	xhtml.Add(h,
		xhtml.WithAny(xhtml.NewHead(), func(a *xhtml.Any) {
			xhtml.Add(a, xhtml.NewTitle(xhtml.Text("hello")))
		}),
		xhtml.WithAny(xhtml.NewBody(), func(a *xhtml.Any) {
			xhtml.Add(a, xhtml.Text("Hello World"))
		}),
	)
	got, _ := h.HTML()
	fmt.Println(string(got))
	// Output:
	// <html><head><title>hello</title></head><body>Hello World</body></html>
}
