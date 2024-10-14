//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-04

package xhttp

import (
	"context"
	"net/http"

	"github.com/xanygo/anygo/xhttp/internal/zroute"
	"github.com/xanygo/anygo/xlog"
)

// MiddlewareFunc HTTP Router(路由) 的 中间件函数类型定义，可用于给 Router 增加切面。
// 如添加额外的日志、可用性统计等独立的功能
type MiddlewareFunc func(http.Handler) http.Handler

var _ http.Handler = (*Router)(nil)

func NewRouter() *Router {
	rt := &Router{}
	return rt
}

//	Router HTTP Router
//
// 路由地址 Path 支持静态地址和通配符(下文中 Router 的 Handle、Get 等 API 文档中所述的 Path就是此)：
//
//  1. 静态路由地址： /user
//
//  2. 单词通配：一个变量匹配一个目录，可以有前缀和后缀
//
//     /user/{name}/{age}, /user/{id}/detail, /user/{id}.html, /user/hello-{id}.html
//
//  3. 正则表达式：
//
//     /user/{category}/{id:[0-9]+}, /user/{id:[0-9]+}.html, /user/hello-{id:[0-9]+}-{age:[0-9]+}.html
//
//  4. *通配符（简化正则）(* 可以匹配包含 / 的所有字符)
//     /user/*,  /user/*/detail, /user/*/detail/*, /user/*/detail/*.html
//     /user/{s1:*},  /user/{s1:*}/detail,  /user/{s1:*}/detail/{s2:*}
//
// 路由变量读取：
//
//	在路由地址中的 {name},* 和 {id:[0-9]+} 均为路由变量，是可以使用 http.Request.PathValue("name") 读取对应的值的。
//	对于 {name} 、 {id:[0-9]+}、{age:*} 这种，分别使用 PathValue("name") 和 PathValue("id")、PathValue("age") 就能读取到。
//	对于使用了 /user/*/detail/* 这种方式，分别使用 PathValue("p0")、PathValue("p1")，即变量名 = p + 变量序号( 从 0 开始 )
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
// pattern： 支持格式 (Method\s+)?(Path)
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

// HandleFunc  注册路由， pattern 支持格式 (Method\s+)?(Path)
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

// Use 给路由注册新的中间件，
// 在使用 Handle、HandleFunc、Get、GetFunc 等注册 Handler 时，会读取已注册到 Router 的中间件列表，用于对 Handler 添加切面。
// 在完成注册 Handler 之后，使用 Use 方法注册的中间件不会影响 已注册的 Handler。
func (r *Router) Use(mds ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, mds...)
}

// Prefix 给地址前缀 prefix 生成一个独立的分组，
// prefix 只能是静态地址，不能包含变量参数，如 /user/
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

// Head  注册 HEAD 请求路由，pattern 支持格式 (Method\s+)?(Path)
func (r *Router) Head(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodHead, pattern, handler, mds...)
}

// HeadFunc  注册 HEAD 请求路由，pattern 支持格式 (Method\s+)?(Path)
func (r *Router) HeadFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Head(pattern, handler, mds...)
}

// Get  注册 GET 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) Get(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodGet, pattern, handler, mds...)
}

// GetFunc  注册 GET 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) GetFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Get(pattern, handler, mds...)
}

// Post  注册 POST 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) Post(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodPost, pattern, handler, mds...)
}

// PostFunc  注册 POST 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) PostFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Post(pattern, handler, mds...)
}

// Delete  注册 DELETE 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) Delete(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodDelete, pattern, handler, mds...)
}

// DeleteFunc  注册 DELETE 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) DeleteFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Delete(pattern, handler, mds...)
}

// Put  注册 PUT 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) Put(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodPut, pattern, handler, mds...)
}

// PutFunc  注册 PUT 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) PutFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Put(pattern, handler, mds...)
}

// Trace  注册 TRACE 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) Trace(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodTrace, pattern, handler, mds...)
}

// TraceFunc  注册 TRACE 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) TraceFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Trace(pattern, handler, mds...)
}

// Options  注册 OPTIONS 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) Options(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.handleMethod(http.MethodOptions, pattern, handler, mds...)
}

// OptionsFunc  注册 OPTIONS 请求路由，pattern 应是一个 Path 格式的字符串
func (r *Router) OptionsFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Options(pattern, handler, mds...)
}
