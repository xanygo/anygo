//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml_test

import (
	"testing"

	"github.com/xanygo/anygo/xhtml"
	"github.com/xanygo/anygo/xt"
)

func TestAttributes(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		attr := &xhtml.WithAttrs{}
		xhtml.SetID(attr, "#abc")
		xhtml.SetName(attr, "hello")
		xhtml.DeleteClass(attr, "c0")

		xhtml.SetClass(attr, "c1", "c2")
		xhtml.SetClass(attr, "c3", "c4")
		xhtml.AddClass(attr, "c5")
		xhtml.DeleteClass(attr, "c4", "c6")

		xhtml.SetValue(attr, `"你好<>"`)

		attrs := attr.FindAttrs()
		bf, err := attrs.HTML()
		xt.NoError(t, err)
		want := `id="#abc" name="hello" class="c3 c5" value="\"你好<>\""`
		xt.Equal(t, want, string(bf))

		wantKeys := []string{"id", "name", "class", "value"}
		xt.Equal(t, wantKeys, attrs.Keys())
	})
}
