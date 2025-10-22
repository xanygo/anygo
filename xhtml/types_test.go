//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml_test

import (
	"testing"

	"github.com/xanygo/anygo/xhtml"
	"github.com/xanygo/anygo/xt"
)

func TestStringSlice_Codes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var a xhtml.TextStringSlice
		got, err := a.Elements("li", nil).HTML()
		xt.NoError(t, err)
		want := ``
		xt.Equal(t, want, string(got))
	})

	t.Run("1 value", func(t *testing.T) {
		a := xhtml.TextStringSlice{"123"}
		got, err := a.Elements("li", func(b *xhtml.Any) {
			xhtml.SetClass(b, "red")
		}).HTML()
		xt.NoError(t, err)
		want := `<li class="red">123</li>`
		xt.Equal(t, want, string(got))
	})

	t.Run("2 value", func(t *testing.T) {
		a := xhtml.TextStringSlice{"123", "456"}
		got, err := a.Elements("li", nil).HTML()
		xt.NoError(t, err)
		want := "<li>123</li><li>456</li>"
		xt.Equal(t, want, string(got))
	})
}

func TestStringSlice_HTML(t *testing.T) {
	ss := xhtml.TextStringSlice{"hello", "world"}
	b, err := ss.HTML()
	xt.NoError(t, err)
	want := "<ul><li>hello</li><li>world</li></ul>"
	xt.Equal(t, want, string(b))
}
