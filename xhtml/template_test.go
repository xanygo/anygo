//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-07

package xhtml

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/fsgo/fst"
)

func TestFuncMap(t *testing.T) {
	tpl := template.Must(template.New("demo").Funcs(FuncMap).Parse(`hello`))
	bf := &bytes.Buffer{}
	fst.NoError(t, tpl.Execute(bf, nil))
	fst.NotEmpty(t, bf.String())
}
