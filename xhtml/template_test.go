//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-07

package xhtml_test

import (
	"bytes"
	"html/template"
	"os"
	"testing"

	"github.com/xanygo/anygo/xhtml"
	"github.com/xanygo/anygo/xt"
)

func TestDump(t *testing.T) {
	bf := &bytes.Buffer{}
	xhtml.Dump(bf, nil)
	xt.NotEmpty(t, bf.String())

	bf.Reset()

	h := xhtml.NewBody()
	xhtml.Dump(bf, h)
	xt.NotEmpty(t, bf.String())
}

func TestFuncMap(t *testing.T) {
	tpl := template.Must(template.New("demo").Funcs(xhtml.FuncMap).Parse(`hello`))
	bf := &bytes.Buffer{}
	xt.NoError(t, tpl.Execute(bf, nil))
	xt.NotEmpty(t, bf.String())
}

func TestWalkParseFS(t *testing.T) {
	tpl := template.New("demo")
	var err error
	tpl, err = xhtml.WalkParseFS(tpl, os.DirFS("./"), ".", "*.html")
	xt.NoError(t, err)
	xt.NotEmpty(t, tpl)
}
