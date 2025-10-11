//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-04

package xhttp

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/xanygo/anygo/xcodec"
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

//	Router 支持静态地址和通配符、支持中间件的 HTTP Router
//
// # Pattern 格式：
//
//	在使用 Handle、HandleFunc 注册路由时，路由的 Pattern 支持格式 ：(Method\s+)?(Path)(\s+meta|Meta)? 。
//	在使用 Get、GetFUnc 等方法中带有 Method 的时候，Pattern 支持格式： (Path)(\s+meta|Meta)? 。
//
// # Method, Handler 支持的请求方法:
//
// Pattern 中的 Method 可以配置 [0-N] 个，当 Method 为空时，此 handler 可以处理所有的请求类型（路由信息会读取到 Method = "ANY" ）。
// 当为多个时，使用英文逗号连接。注册时的填写的 Method 不区分大小写，传入后会统一转换为大写（如 get -> GET）。
// 如 Handle("/index")，Handle("get /index")，Handle("get,post /index") 。
//
// # Path, 路由地址，支持静态地址和通配符
//
// 下文中 Router 的 Handle、Get 等 API 文档中所述的 Path就是此。
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
//     正则表达式别名：
//     /{id:UUID} 、/{id:UINT}
//     UUID 可匹配 xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx 这个格式的 UUID
//     UINT 可匹配正整数
//     Base62 匹配 [0-9a-zA-Z]+
//     Base36 匹配 [0-9a-z]+
//     Base64URL 匹配 [0-9a-zA-Z\-_]+
//     除此之外，还可以使用 RegisterRegexpAlias 注册自定义的别名
//
//  4. *通配符（简化正则）(* 可以匹配包含 / 的所有字符)
//     /user/*,  /user/*/detail, /user/*/detail/*, /user/*/detail/*.html
//     /user/{s1:*},  /user/{s1:*}/detail,  /user/{s1:*}/detail/{s2:*}
//
// # 路由变量读取：
//
//	在路由地址中的 {name},* 和 {id:[0-9]+} 均为路由变量，是可以使用 http.Request.PathValue("name") 读取对应的值的。
//	对于 {name} 、 {id:[0-9]+}、{age:*} 这种，分别使用 PathValue("name") 和 PathValue("id")、PathValue("age") 就能读取到。
//	对于使用了 /user/*/detail/* 这种方式，分别使用 PathValue("p0")、PathValue("p1")，即变量名 = p + 变量序号( 从 0 开始 )
//
// # Meta 路由元信息：
//
//	在 Handler 中或者中间件中使用 ReadRouteInfo(http.Request.Context()) 可以读取此 Handler 注册的路由信息。
//	如用于监控、鉴权等场景时，可以读取 RouteInfo 信息。
//	在 Pattern 中，除了 Path 等其他元信息可以通过在 Pattern 中添加 (\s+meta|Meta) 段落内容添加。
//	如 Handle("get /index meta|id=1,type=user")。id 字段时固定的字段，还可以添加其他任意 key 。
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
	ctx := ContextWithRequest(req.Context(), req)
	if sr != nil {
		ctx = contextWithRouteInfo(ctx, sr.Info.(RouteInfo))
		req = req.WithContext(ctx)
		sr.ServeHTTP(w, req)
		return
	}
	req = req.WithContext(ctx)
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
		h = r.wrap(NotFoundHandler())
	}
	h.ServeHTTP(w, req)
}

// Handle  注册路由
//
// pattern： 支持格式 (Method\s+)?(Path)(\s+meta|Meta)
//
//	Method: 请求方法，可选，支持一个活多个，如 “GET”，“GET,POST”
//	若 Method 为空则不限定请求方法
//
//	Path: 请求地址，支持静态地址和通配符
//
//	Meta：路由的其他元信息
//	如 meta|id=123 或者 meta|id=123,type=user
func (r *Router) Handle(pattern string, handler http.Handler, mds ...MiddlewareFunc) {
	r.register(pattern, handler, mds...)
}

func (r *Router) register(pattern string, handler http.Handler, mds ...MiddlewareFunc) []RouteInfo {
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

	result := make([]RouteInfo, 0, len(routes))
	for _, route := range routes {
		route.Handler = handler
		info := RouteInfo{
			Method:    route.Method,
			Pattern:   route.Pattern,
			Path:      zroute.CleanPattern(route.Pattern),
			MetaID:    route.Meta.ID,
			MetaOther: route.Meta.Other,
		}
		route.Info = info
		r.subRoute = append(r.subRoute, route)
		result = append(result, info)

		r.AutoLogger().Debug(context.Background(), "Route", route.LogFields()...)
	}
	return result
}

// HandleFunc  注册路由， pattern 支持格式 (Method\s+)?(Path)(\s+meta|Meta)
func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) {
	r.Handle(pattern, handler, mds...)
}

func (r *Router) handleMethod(method string, pattern string, handler http.Handler, mds ...MiddlewareFunc) RouteInfo {
	infos := r.register(method+" "+pattern, handler, mds...)
	return infos[0]
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
	// 单独处理，这样使用 Prefix 注册的逻辑，中间件函数能读取到最终注册的路由信息
	// 否则只能读取到 注册 Prefix 时候的路由
	if rr, ok := h.(*Router); ok {
		old := rr.middlewares
		// 保持 父 Router 注册的 中间件还是在前面
		rr.middlewares = slices.Clone(r.middlewares)
		rr.Use(old...)
		rr.Use(mds...)
		return h
	}

	for i := len(mds) - 1; i >= 0; i-- {
		h = mds[i](h)
	}
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
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

// Head  注册 HEAD 请求路由，pattern 支持格式 (Method\s+)?(Path)(\s+meta|Meta)
func (r *Router) Head(pattern string, handler http.Handler, mds ...MiddlewareFunc) RouteInfo {
	return r.handleMethod(http.MethodHead, pattern, handler, mds...)
}

// HeadFunc  注册 HEAD 请求路由，pattern 支持格式 (Method\s+)?(Path)(\s+meta|Meta)
func (r *Router) HeadFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) RouteInfo {
	return r.Head(pattern, handler, mds...)
}

// Get  注册 GET 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) Get(pattern string, handler http.Handler, mds ...MiddlewareFunc) RouteInfo {
	return r.handleMethod(http.MethodGet, pattern, handler, mds...)
}

// GetFunc  注册 GET 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) GetFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) RouteInfo {
	return r.Get(pattern, handler, mds...)
}

// Post  注册 POST 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) Post(pattern string, handler http.Handler, mds ...MiddlewareFunc) RouteInfo {
	return r.handleMethod(http.MethodPost, pattern, handler, mds...)
}

// PostFunc  注册 POST 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) PostFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) RouteInfo {
	return r.Post(pattern, handler, mds...)
}

// Delete  注册 DELETE 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) Delete(pattern string, handler http.Handler, mds ...MiddlewareFunc) RouteInfo {
	return r.handleMethod(http.MethodDelete, pattern, handler, mds...)
}

// DeleteFunc  注册 DELETE 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) DeleteFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) RouteInfo {
	return r.Delete(pattern, handler, mds...)
}

// Put  注册 PUT 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) Put(pattern string, handler http.Handler, mds ...MiddlewareFunc) RouteInfo {
	return r.handleMethod(http.MethodPut, pattern, handler, mds...)
}

// PutFunc  注册 PUT 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) PutFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) RouteInfo {
	return r.Put(pattern, handler, mds...)
}

// Trace  注册 TRACE 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) Trace(pattern string, handler http.Handler, mds ...MiddlewareFunc) RouteInfo {
	return r.handleMethod(http.MethodTrace, pattern, handler, mds...)
}

// TraceFunc  注册 TRACE 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) TraceFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) RouteInfo {
	return r.Trace(pattern, handler, mds...)
}

// Options  注册 OPTIONS 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) Options(pattern string, handler http.Handler, mds ...MiddlewareFunc) RouteInfo {
	return r.handleMethod(http.MethodOptions, pattern, handler, mds...)
}

// OptionsFunc  注册 OPTIONS 请求路由，pattern 应是一个 (Path)(\s+meta|Meta) 格式的字符串
func (r *Router) OptionsFunc(pattern string, handler http.HandlerFunc, mds ...MiddlewareFunc) RouteInfo {
	return r.Options(pattern, handler, mds...)
}

type RouteInfo struct {
	Method string // 注册的请求方法，如 GET、ANY

	// Pattern 注册的路由地址，如 /user, /user/{id}, /user/*, /user/{category}/{id:[0-9]+}
	Pattern string

	// Path 归一化后的 pattern 地址,,去掉变量的正则只保留变量名，
	// 如 /user, /user/{id}, /user/*， /user/{category}/{id}
	Path string

	// MetaID 注册在 路由 pattern 中的 meta 的 id 值
	MetaID string

	// MetaOther
	MetaOther map[string]string
}

func (ri RouteInfo) Exists() bool {
	return ri.Method != ""
}

func (ri RouteInfo) String() string {
	return xcodec.JSONString(ri)
}

func (ri RouteInfo) GetMeta(key string) (string, bool) {
	if len(ri.MetaOther) == 0 {
		return "", false
	}
	val, ok := ri.MetaOther[key]
	return val, ok
}

type ctxKey uint8

const (
	ctxKeyRouteInfo ctxKey = iota
	ctxKeyRequest
)

func ContextWithRequest(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, ctxKeyRequest, req)
}

// RequestFromContext 从 Context 信息里读取 request 信息
// 默认情况下，Router 已经提前种好了
func RequestFromContext(ctx context.Context) *http.Request {
	req, _ := ctx.Value(ctxKeyRequest).(*http.Request)
	return req
}

func contextWithRouteInfo(ctx context.Context, info RouteInfo) context.Context {
	return context.WithValue(ctx, ctxKeyRouteInfo, info)
}

// ReadRouteInfo 从 http.Request.Context() 信息里读取路由信息
func ReadRouteInfo(ctx context.Context) RouteInfo {
	val, _ := ctx.Value(ctxKeyRouteInfo).(RouteInfo)
	return val
}

func PathJoin(arr ...string) string {
	return zroute.CleanPath(strings.Join(arr, ""))
}

// RegisterRegexpAlias 注册路由正则别买
// 默认已集成 UUID、UINT
func RegisterRegexpAlias(name string, reg string) {
	zroute.RegisterRegexpAlias(name, reg)
}
