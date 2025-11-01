//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-01-01

package xhttp

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/xanygo/anygo/internal"
	"github.com/xanygo/anygo/xlog"
)

var errHTMLTpl = `<!DOCTYPE html>
<html lang="{{ .Lang }}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
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
        <h1>{{ .Code }}</h1>
        <p>{{ .Message }}</p>
       {{ if .GoHomeText }}
         <a href="/">{{ .GoHomeText }}</a>
		{{ end }}
    </div>
</body>
</html>`

var errPage404Zh, errPage404En []byte

var initErrTplOnce sync.Once
var errTpl *template.Template

func doInitErrTpl() {
	errHTMLTpl = strings.ReplaceAll(errHTMLTpl, "\n", "")
	reg1 := regexp.MustCompile(`>\s+<`)
	errHTMLTpl = reg1.ReplaceAllString(errHTMLTpl, `><`)
	reg2 := regexp.MustCompile(`\s{2,}`)
	errHTMLTpl = reg2.ReplaceAllString(errHTMLTpl, ` `)

	errTpl = template.Must(template.New("page").Parse(errHTMLTpl))

	zhData := map[string]any{
		"Lang":       "zh-CN",
		"Title":      "404 - 页面未找到",
		"Code":       "404",
		"Message":    "抱歉，您访问的页面不存在。",
		"GoHomeText": "返回首页",
	}
	bf := &bytes.Buffer{}
	err := errTpl.Execute(bf, zhData)
	if err != nil {
		panic(err)
	}
	errPage404Zh = bytes.Clone(bf.Bytes())
	bf.Reset()

	enData := map[string]string{
		"Lang":       "en",
		"Title":      "404 - Page Not Found",
		"Code":       "404",
		"Message":    "Sorry, the page you are looking for does not exist.",
		"GoHomeText": "Go to Homepage",
	}
	err = errTpl.Execute(bf, enData)
	if err != nil {
		panic(err)
	}
	errPage404En = bytes.Clone(bf.Bytes())
}

func notFound(w http.ResponseWriter, r *http.Request) {
	if internal.HandlerImage404(w, r) {
		return
	}
	if IsAjax(r) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		data := map[string]any{
			"LogID": xlog.FindLogID(r.Context()),
			"Code":  http.StatusNotFound,
			"Msg":   http.StatusText(http.StatusNotFound),
		}
		bf, _ := json.Marshal(data)
		_, _ = w.Write(bf)
		return
	}

	initErrTplOnce.Do(doInitErrTpl)

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

func Error(w http.ResponseWriter, r *http.Request, code int, title string, error string) {
	ctx := r.Context()
	if xlog.IsContext(ctx) {
		xlog.AddAttr(ctx,
			xlog.Int("ErrCode", code),
			xlog.String("ErrTitle", title),
			xlog.String("ErrMsg", error),
		)
	}
	if IsAjax(r) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(code)
		data := map[string]any{
			"LogID": xlog.FindLogID(r.Context()),
			"Code":  code,
			"Msg":   error,
		}
		bf, _ := json.Marshal(data)
		_, _ = w.Write(bf)
		return
	}
	initErrTplOnce.Do(doInitErrTpl)

	if title == "" {
		title = http.StatusText(code)
	}
	data := map[string]string{
		"Lang":       "zh-CN",
		"Title":      title,
		"Code":       strconv.Itoa(code),
		"Message":    error,
		"GoHomeText": "Go to Homepage",
	}
	if code != http.StatusNotFound || r.URL.Path == "/" {
		data["GoHomeText"] = ""
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_ = errTpl.Execute(w, data)
}

// TextError 封装标准库 http.Error，若 context 已是 logContext,则将 code 和 error都记录到日志字段里去
func TextError(w http.ResponseWriter, r *http.Request, error string, code int) {
	ctx := r.Context()
	if xlog.IsContext(ctx) {
		xlog.AddAttr(ctx,
			xlog.Int("ErrCode", code),
			xlog.String("ErrMsg", error),
		)
	}
	http.Error(w, error, code)
}
