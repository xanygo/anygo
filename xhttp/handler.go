//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-14

package xhttp

import (
	"net/http"
	"reflect"

	"github.com/xanygo/anygo/xhttp/internal/zroute"
	"github.com/xanygo/anygo/xmap"
	"github.com/xanygo/anygo/xslice"
	"github.com/xanygo/anygo/xsync"
)

var notFoundHandler xsync.Value[http.Handler]

func NotFound(w http.ResponseWriter, r *http.Request) {
	NotFoundHandler().ServeHTTP(w, r)
}

func NotFoundHandler() http.Handler {
	h := notFoundHandler.Load()
	if h != nil {
		return h
	}
	return http.HandlerFunc(notFound)
}

func SetNotFoundHandler(h http.Handler) {
	notFoundHandler.Store(h)
}

type GroupHandler interface {
	GroupHandler() map[string]PatternHandler
}

type PatternHandler struct {
	Pattern    string
	Func       http.HandlerFunc
	Middleware []MiddlewareFunc
}

// RegisterGroup  注册一组业务，安装规则将 GroupHandler 中的所有的 http.HandlerFunc 注册到路由中去。
//
//	注册规则如下：
//
//	 1. {HTTPMethod} 指的是 Get、Post、Delete 等 ，不区分大小写，所以 GET、POST、DELETE 也一样。
//	 2. 所有和 {HTTPMethod} 或者 {HTTPMethod}{Xyz} 或者 {HTTPMethod}{Xyz}{Abc} 等驼峰命名的，
//	 注册的路由只支持此种 HTTP 请求，如 Delete 方法只支持 HTTP DELETE 请求。
//	 3. 方法名中包含 Save 的，注册为 POST 请求。
//
// 假设 RegisterGroup(r,"/user/",&userHandler{}),userHandler{} 中所有实现了 func(http.ResponseWriter, *http.Request)
// 这个函数签名的注册结果如下：
//
//	user.Index        --> GET      /user/ 和 /user/Index
//	user.Delete       --> DELETE   /user/
//	user.Post         --> POST     /user/
//	user.GetByID      --> GET      /user/GetByID
//	user.DeleteByID   --> DELETE   /user/DeleteByID
//	user.UpdateStatus --> PUT      /user/UpdateStatus
//	user.Search       --> GET      /user/Search
//	user.Add          --> GET      /user/Add
//	user.Edit         --> GET      /user/Edit
//	user.Save         --> POST     /user/Save
func RegisterGroup(r *Router, prefix string, h GroupHandler, mds ...MiddlewareFunc) {
	rt := reflect.TypeOf(h)
	rv := reflect.ValueOf(h)

	meta := map[string]string{
		"Prefix":       prefix,
		"GroupHandler": rt.String(),
	}

	infos := h.GroupHandler()

	for i := 0; i < rv.NumMethod(); i++ {
		mt := rt.Method(i)
		mv := rv.Method(i)
		hd, ok := mv.Interface().(func(http.ResponseWriter, *http.Request))
		if !ok {
			continue
		}
		name := mt.Name
		meta["MethodName"] = name
		metaStr := " meta|" + xmap.Join(meta, ",")

		handler := http.HandlerFunc(hd)

		if len(infos) > 0 {
			if info, has := infos[name]; has {
				// Func 为 nil 时跳过，不注册
				if info.Func == nil {
					continue
				}
				fns := xslice.Merge(info.Middleware, mds)
				r.HandleFunc(prefix+"/"+info.Pattern+metaStr, info.Func, fns...)
				continue
			}
		}

		method := zroute.GetPrefixMethod(name)
		if name == "Index" {
			r.handleMethod(method, prefix+metaStr, handler, mds...)
		}
		r.handleMethod(method, prefix+"/"+name+metaStr, handler, mds...)
	}
}
