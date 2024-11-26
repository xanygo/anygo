//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-14

package xhttp

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/internal"
	"github.com/xanygo/anygo/xhttp/internal/zroute"
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

const errHTMLTpl = `<!DOCTYPE html>
<html lang="{Lang}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{Title}</title>
    <style>
        body {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            text-align: center;
        }
        h1 {
            font-size: 80px;
            color: #FF6347;
            margin: 0;
        }
        .emoji {
            font-size: 60px;
            margin-bottom: 20px;
        }
        p {
            font-size: 18px;
            color: #555;
        }
        a {
            display: inline-block;
            margin-top: 20px;
            padding: 10px 20px;
            background-color: #FF6347;
            color: #fff;
            text-decoration: none;
            border-radius: 5px;
        }
        a:hover {
            background-color: #FF4500;
        }
    </style>
</head>
<body>
    <div>
        <div class="emoji">😞</div>
        <h1>{Code}</h1>
        <p>{Message}</p>
        <a href="/">{GoHomeText}</a>
    </div>
</body>
</html>`

var errPage404Zh, errPage404En []byte

func init() {
	zhData := []string{
		"{Lang}", "zh-CN",
		"{Title}", "404 - 页面未找到",
		"{Code}", "404",
		"{Message}", "抱歉，您访问的页面不存在。",
		"{GoHomeText}", "返回首页",
	}
	errPage404Zh = []byte(strings.NewReplacer(zhData...).Replace(errHTMLTpl))

	enData := []string{
		"{Lang}", "en",
		"{Title}", "404 - Page Not Found",
		"{Code}", "404",
		"{Message}", "Sorry, the page you are looking for does not exist.",
		"{GoHomeText}", "Go to Homepage",
	}
	errPage404En = []byte(strings.NewReplacer(enData...).Replace(errHTMLTpl))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	if internal.HandlerImage404(w, r) {
		return
	}
	w.WriteHeader(http.StatusNotFound)
	accept := r.Header.Get("Accept-Language")
	zhIndex := strings.Index(accept, "zh")
	enIndex := strings.Index(accept, "en")
	if enIndex != -1 && enIndex < zhIndex {
		_, _ = w.Write(errPage404En)
		return
	}
	_, _ = w.Write(errPage404Zh)
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
// 假设 RegisterGroup(r,"/user",&userHandler{}),userHandler{} 中所有实现了 func(http.ResponseWriter, *http.Request)
// 这个函数签名的注册结果如下：
//
//	user.Get          --> GET      /user
//	user.Delete       --> DELETE   /user
//	user.Post         --> POST     /user
//	user.GetByID      --> GET      /userByID
//	user.DeleteByID   --> DELETE   /userByID
//	user.Search       --> GET      /userSearch
//	user.Add          --> GET      /userAdd
//	user.Edit         --> GET      /userEdit
//	user.Save         --> POST     /userSave
func RegisterGroup(r *Router, pattern string, h GroupHandler, mds ...MiddlewareFunc) {
	rt := reflect.TypeOf(h)
	rv := reflect.ValueOf(h)

	infos := h.GroupHandler()

	for i := 0; i < rv.NumMethod(); i++ {
		mt := rt.Method(i)
		mv := rv.Method(i)
		hd, ok := mv.Interface().(func(http.ResponseWriter, *http.Request))
		if !ok {
			continue
		}
		name := mt.Name
		handler := http.HandlerFunc(hd)

		if len(infos) > 0 {
			if info, has := infos[name]; has {
				// Func 为 nil 时跳过，不注册
				if info.Func == nil {
					continue
				}
				fns := xslice.Merge(info.Middleware, mds)
				r.HandleFunc(pattern+info.Pattern, info.Func, fns...)
				continue
			}
		}

		if zroute.IsMethod(name) {
			// 如方法名是  Get、Post、Delete 等
			r.handleMethod(name, pattern, handler, mds...)
		} else if strings.Contains(name, "Save") {
			// 如方法名是  Save、SaveByID
			r.PostFunc(pattern+name, handler, mds...)
		} else {
			// 如方法名是  GetUser、GetUserList、Search 等
			str1, str2 := zroute.SplitCamelCase(name)
			if str2 == "" {
				r.GetFunc(pattern+name, handler, mds...)
			} else {
				if zroute.IsMethod(str1) {
					// 如方法名是  GetByID
					r.handleMethod(str1, pattern+str2, handler, mds...)
				} else {
					r.GetFunc(pattern+name, handler, mds...)
				}
			}
		}
	}
}
