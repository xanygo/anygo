//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-06

package xhtml

import (
	"context"
	"html/template"
	"maps"
	"net/http"

	"github.com/xanygo/anygo/xhtml/internal/tplfn"
	"github.com/xanygo/anygo/xsync"
)

type TPLHelper struct {
	Request *http.Request
}

func (t *TPLHelper) Context() context.Context {
	return t.Request.Context()
}

var allTemplateFuncs = &xsync.OnceInit[template.FuncMap]{
	New: func() template.FuncMap {
		values := maps.Clone(templateFuncs)
		maps.Copy(values, tplfn.Funcs)
		return values
	},
}

func FuncMap() template.FuncMap {
	return allTemplateFuncs.Load()
}

var templateFuncs = template.FuncMap{
	"xRender": Render,
}
