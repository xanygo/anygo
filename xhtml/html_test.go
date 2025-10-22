//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml_test

import (
	"testing"

	"github.com/xanygo/anygo/xhtml"
	"github.com/xanygo/anygo/xt"
)

func TestNewDiv(t *testing.T) {
	t.Run("div_empty", func(t *testing.T) {
		div := xhtml.NewDiv()
		got, err := div.HTML()
		xt.NoError(t, err)
		want := `<div></div>`
		xt.Equal(t, want, string(got))
	})

	t.Run("div_p", func(t *testing.T) {
		div := xhtml.NewDiv()
		xhtml.SetID(div, "#abc")
		div.Body = xhtml.ToElements(xhtml.NewP())
		got, err := div.HTML()
		xt.NoError(t, err)
		want := `<div id="#abc"><p></p></div>`
		xt.Equal(t, want, string(got))
	})

	t.Run("div_attrs", func(t *testing.T) {
		div := xhtml.NewDiv()
		xhtml.SetClass(div, "c1", "c2")
		xhtml.SetID(div, "#abc")
		got, err := div.HTML()
		xt.NoError(t, err)
		want := `<div class="c1 c2" id="#abc"></div>`
		xt.Equal(t, want, string(got))
	})
}

func TestBody(t *testing.T) {
	t.Run("with children", func(t *testing.T) {
		body := xhtml.NewBody()
		sa := &xhtml.StyleAttr{}
		sa.MaxWidth("100px").Height("200px")
		xt.NoError(t, sa.SetTo(body))
		div := xhtml.NewDiv()
		div.Body = append(div.Body, xhtml.TextString("hello"))
		body.Body = xhtml.ToElements(div)
		got, err := body.HTML()
		xt.NoError(t, err)
		want := `<body style="max-width:100px; height:200px"><div>hello</div></body>`
		xt.Equal(t, want, string(got))
	})
}

func TestIMG(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		a := xhtml.NewIMG("/a.jpg")
		a.ALT("hello")
		got, err := a.HTML()
		xt.NoError(t, err)
		want := `<img src="/a.jpg" alt="hello"/>`
		xt.Equal(t, want, string(got))
	})

	t.Run("width_height", func(t *testing.T) {
		a := xhtml.NewIMG("/a.jpg")
		xhtml.SetWidth(a, "100px")
		xhtml.SetHeight(a, "110px")
		got, err := a.HTML()
		xt.NoError(t, err)
		want := `<img src="/a.jpg" width="100px" height="110px"/>`
		xt.Equal(t, want, string(got))
	})
}

func TestA(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		a := xhtml.NewA("/gogo")
		xhtml.SetTitle(a, "hello")
		got, err := a.HTML()
		xt.NoError(t, err)
		want := `<a href="/gogo" title="hello"></a>`
		xt.Equal(t, want, string(got))
	})
}

func TestMeta(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		a := xhtml.NewMeta()
		a.Name("robots").Content("all")
		got, err := a.HTML()
		xt.NoError(t, err)
		want := `<meta name="robots" content="all"/>`
		xt.Equal(t, want, string(got))
	})
}

func TestLink(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		a := xhtml.NewLink()
		a.Rel("stylesheet").Type("text/css").Href("/a.css")
		got, err := a.HTML()
		xt.NoError(t, err)
		want := `<link rel="stylesheet" type="text/css" href="/a.css"/>`
		xt.Equal(t, want, string(got))
	})
}

func TestScript(t *testing.T) {
	t.Run("async", func(t *testing.T) {
		a := xhtml.NewScript()
		xhtml.SetAsync(a)
		got, err := a.HTML()
		xt.NoError(t, err)
		want := `<script async></script>`
		xt.Equal(t, want, string(got))
	})
}

func TestInput(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		a := xhtml.NewInput("text", "")
		xhtml.SetValue(a, "hello")
		xhtml.SetOnChange(a, `alter("ok")`)
		got, err := a.HTML()
		xt.NoError(t, err)
		want := `<input type="text" value="hello" onchange="alter(\"ok\")"/>`
		xt.Equal(t, want, string(got))
	})
}
