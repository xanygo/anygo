//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-14

package xhttp

import (
	"net/http"
	"strings"

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
