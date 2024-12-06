//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-06

package tplfn

import (
	"bytes"
	"github.com/fsgo/fst"
	"html/template"
	"testing"
)

func TestFuncs(t *testing.T) {
	tpl := template.Must(template.New("demo").Funcs(Funcs).Parse(`hello`))
	bf := &bytes.Buffer{}
	fst.NoError(t, tpl.Execute(bf, nil))
	fst.NotEmpty(t, bf.String())
}
