//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml_test

import (
	"testing"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/xhtml"
)

func TestTable1(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		tb := &xhtml.Table1{}
		tb.SetHeader(xhtml.NewTh(xhtml.String("name")), xhtml.NewTh(xhtml.String("age")))
		tb.AddRow(xhtml.NewTd(xhtml.String("lilei")), xhtml.NewTd(xhtml.String("18")))
		tb.AddRow(xhtml.NewTd(xhtml.String("hanmeimei")), xhtml.NewTd(xhtml.String("15")))
		tb.SetFooter(xhtml.NewTd(xhtml.String("f1")), xhtml.NewTd(xhtml.String("f2")))

		xhtml.SetID(tb, "#abc")

		got, err := tb.HTML()
		fst.NoError(t, err)
		want := `<table id="#abc">` + "\n" +
			"<thead>\n<tr><th>name</th><th>age</th></tr>\n</thead>\n" +
			"<tbody>\n" +
			"<tr><td>lilei</td><td>18</td></tr>\n" +
			"<tr><td>hanmeimei</td><td>15</td></tr>\n" +
			"</tbody>\n" +
			"<tfoot>\n<tr><td>f1</td><td>f2</td></tr>\n</tfoot>\n" +
			"</table>\n"
		fst.Equal(t, want, string(got))
	})
}
