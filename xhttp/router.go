//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-04

package xhttp

import (
	"context"
	"github.com/xanygo/anygo/xhttp/internal/zroute"
	"github.com/xanygo/anygo/xlog"
	"net/http"
)

type MiddlewareFunc func(http.Handler) http.Handler

var _ http.Handler = (*Router)(nil)

func NewRouter() *Router {
	rt := &Router{}
	return rt
}

//	Router HTTP Router
//
// 请求地址 Path 支持静态地址和通配符：
//
//  1. 静态路由地址： /user
//
//  2. 单词通配：一个目录最多允许一个变量，可以有前缀和后缀
//     /user/{name}/{age}, /user/{id}/detail,/user/{id}.html,/user/hello-{id}.html
//
//  3. 正则表达式：/user/{category}/{id:[0-9]+}, /user/{id:[0-9]+}.html, /user/hello-{id:[0-9]+}-{age:[0-9]+}.html
//
//  4. *通配符（简化正则）(* 可以匹配包含 / 的所有字符)
//     /user/*,  /user/*/detail, /user/*/detail/*, /user/*/detail/*.html
//     /user/{s1:*},  /user/{s1:*}/detail,  /user/{s1:*}/detail/{s2:*}
type Router struct {
	xlog.WithLogger

	prefix      string
	middlewares []MiddlewareFunc
	notFound    http.Handler
	notFoundRaw http.Handler
	subRoute    []*zroute.Route
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.Method = zroute.CleanMethod(req.Method)
	sr, values := r.findRoute(req)
	for k, v := range values {
		req.SetPathValue(k, v)
	}
	if sr != nil {
		sr.ServeHTTP(w, req)
		return
	}
	r.doNotFound(w, req)
}

func (r *Router) findRoute(req *http.Request) (*zroute.Route, map[string]string) {
	for _, sub := range r.subRoute {
		values, ok := sub.Match(req)
		if ok {
			return sub, values
		}
	}
	return nil, nil
}

func (r *Router) doNotFound(w http.ResponseWriter, req *http.Request) {
	h := r.notFound
	if h == nil {
		h = r.wrap(http.HandlerFunc(NotFound))
	}
	h.ServeHTTP(w, req)
}

// Handle  注册路由
//
// prefix： 支持格式 (Method\s+)?(Path)
//
//	Method: 请求方法，可选，支持一个活多个，如 “GET”，“GET,POST”
//	若 Method 为空则不限定请求方法
//
//	Path: 请求地址，支持静态地址和通配符
func (r *Router) Handle(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	routes, err := zroute.ParserPattern(r.prefix, pattern)
	if err != nil {
		panic(err)
	}

	if handler == nil {
		panic(pattern + ": register with a nil handler")
	}

	r.AutoLogger().Debug(context.Background(), "Handle",
		xlog.String("Pattern", pattern),
		xlog.Int("Routes.cnt", len(routes)),
	)

	handler = r.wrap(handler, mds...)

	for _, route := range routes {
		route.Handler = handler
		r.subRoute = append(r.subRoute, route)

		r.AutoLogger().Debug(context.Background(), "Route", route.LogFields()...)
	}
}

func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Handle(pattern, handler, mds...)
}

func (r *Router) handleMethod(method string, pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.Handle(method+" "+pattern, handler, mds...)
}

func (r *Router) NotFound(handler http.Handler) {
	if handler == nil {
		panic("cannot register a nil handler for NotFound")
	}
	r.notFound = r.wrap(handler)
	r.notFoundRaw = handler
}

func (r *Router) NotFoundFunc(handler http.HandlerFunc) {
	r.NotFound(handler)
}

func (r *Router) wrap(h http.Handler, mds ...MiddlewareFunc) http.Handler {
	for _, mf := range r.middlewares {
		h = mf(h)
	}
	for _, mf := range mds {
		h = mf(h)
	}
	return h
}

func (r *Router) Use(mds ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, mds...)
}

func (r *Router) Prefix(prefix string, mds ...MiddlewareFunc) *Router {
	if prefix == "" {
		panic("prefix must not be empty")
	}
	g := &Router{
		prefix:      zroute.CleanPath(r.prefix + prefix),
		middlewares: mds,
	}
	if g.HasLogger() {
		g.SetLogger(r.Logger())
	}
	if r.notFoundRaw != nil {
		g.NotFound(r.notFoundRaw)
	}
	r.Handle(g.prefix+"*", g)
	return g
}

func (r *Router) Head(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodHead, pattern, handler, mds...)
}

func (r *Router) HeadFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Head(pattern, handler, mds...)
}

func (r *Router) Get(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodGet, pattern, handler, mds...)
}

func (r *Router) GetFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Get(pattern, handler, mds...)
}

func (r *Router) Post(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodPost, pattern, handler, mds...)
}

func (r *Router) PostFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Post(pattern, handler, mds...)
}

func (r *Router) Delete(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodDelete, pattern, handler, mds...)
}

func (r *Router) DeleteFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Delete(pattern, handler, mds...)
}

func (r *Router) Put(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodPut, pattern, handler, mds...)
}

func (r *Router) PutFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Put(pattern, handler, mds...)
}

func (r *Router) Trace(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodTrace, pattern, handler, mds...)
}

func (r *Router) TraceFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Trace(pattern, handler, mds...)
}

func (r *Router) Options(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodOptions, pattern, handler, mds...)
}

func (r *Router) OptionsFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Options(pattern, handler, mds...)
}
