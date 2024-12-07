//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-06

package xhtml

import (
	"context"
	"html/template"
	"maps"
	"net/http"
	"strings"

	"github.com/xanygo/anygo/xhtml/internal/tplfn"
	"github.com/xanygo/anygo/xsync"
	"github.com/xanygo/anygo/xurl"
)

func NewTPLRequest(req *http.Request) *TPLRequest {
	return &TPLRequest{
		Request: req,
	}
}

type TPLRequest struct {
	Request *http.Request
}

func (t *TPLRequest) Context() context.Context {
	return t.Request.Context()
}

// Query 获取 url 的 query 参数值
func (t *TPLRequest) Query(name string) string {
	query := t.Request.URL.Query()
	return query.Get(name)
}

// BaseLink 基于当前 url，生成新的链接
//
// query：url 中的 query 参数，如 "a=1&b=2&c="，同名参数会将当前链接中的同名参数覆盖，值为空的则将其删除
func (t *TPLRequest) BaseLink(query string) template.URL {
	return template.URL(xurl.NewLink(t.Request.URL, query))
}

func (t *TPLRequest) IfQueryEQ(name string, value string, echo any) any {
	query := t.Request.URL.Query()
	if query.Get(name) == value {
		return echo
	}
	return nil
}

func (t *TPLRequest) IfPathHas(sub string, echo any) any {
	if strings.Contains(t.Request.URL.Path, sub) {
		return echo
	}
	return nil
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
