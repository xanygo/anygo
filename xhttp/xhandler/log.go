//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-14

package xhandler

import (
	"context"
	"net/http"
	"time"

	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/xctx"
	"github.com/xanygo/anygo/xhttp"
	"github.com/xanygo/anygo/xlog"
)

// AccessLog 打印访问日志
type AccessLog struct {
	// Logger 必填，打印日志的 logger
	Logger xlog.Logger

	// OnCookies 可选，处理 Cookie
	OnCookies func(cookies []*http.Cookie) []xlog.Attr

	// OnHeaders 可选，处理 Header
	OnHeaders func(h http.Header) []xlog.Attr

	// OnPanic 可选，panic 后，自定义输出
	OnPanic func(w http.ResponseWriter, r *http.Request, re any)

	// OnRequest 可选，ctx 日志字段初始化完成后，在执行后续 ServeHTTP 方法前调用
	OnRequest func(ctx context.Context, r *http.Request)

	// RePanic 当 panic 发生后，是否将 panic 重新抛出
	RePanic bool
}

func (al *AccessLog) Next(handler http.Handler) http.Handler {
	if al.Logger == nil {
		return handler
	}
	handler = al.safely(handler)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := xlog.NewContext(r.Context())
		ctx = al.before(ctx, start, r)
		r = r.WithContext(ctx)
		defer al.after(ctx, start, w, r)
		if al.OnRequest != nil {
			al.OnRequest(ctx, r)
		}
		handler.ServeHTTP(w, r)
	})
}

func (al *AccessLog) safely(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer al.safelyRecover(w, r)
		handler.ServeHTTP(w, r)
	})
}

func (al *AccessLog) safelyRecover(w http.ResponseWriter, r *http.Request) {
	re := recover()
	if re == nil {
		return
	}
	err := safely.NewPanicErr(re, 2)
	safely.RecoveredPECtx(r.Context(), err)
	if al.OnPanic != nil {
		al.OnPanic(w, r, re)
	} else {
		xhttp.WriteTextStatus(w, http.StatusInternalServerError, []byte("Internal Server Error"))
	}
	al.Logger.Error(r.Context(), "panic", xlog.ErrorAttr("panic", err))
	if al.RePanic {
		panic(err)
	}
}

var ctxLogFieldsKey = xctx.NewKey()

func (al *AccessLog) before(ctx context.Context, start time.Time, r *http.Request) context.Context {
	var cookies []xlog.Attr
	if al.OnCookies == nil {
		cookies = al.cookies(r.Cookies())
	} else {
		cookies = al.OnCookies(r.Cookies())
	}
	xlog.WithLogID(ctx, al.logID(r))

	var headers []xlog.Attr
	if al.OnHeaders == nil {
		headers = al.headers(r.Header)
	} else {
		headers = al.OnHeaders(r.Header)
	}

	xlog.AddMetaAttr(ctx,
		xlog.String("method", r.Method),
		xlog.String("uri", r.RequestURI),
		xlog.String("remote", r.RemoteAddr),
	)

	xlog.AddAttr(ctx,
		xlog.Int64("start", start.UnixMilli()),
		xlog.String("host", r.Host),
	)

	// 这两个字段在最后打印打印即可，提前保存以避免被修改
	ctxAttrs := []xlog.Attr{
		xlog.GroupAttrs("cookie", cookies...),
		xlog.GroupAttrs("header", headers...),
	}
	return context.WithValue(ctx, ctxLogFieldsKey, ctxAttrs)
}

func (al *AccessLog) logID(r *http.Request) string {
	if id := r.Header.Get("X-LogID"); id != "" {
		return id
	}
	if id := r.URL.Query().Get("logid"); id != "" {
		return id
	}
	return xlog.NewLogID()
}

func (al *AccessLog) cookies(cookies []*http.Cookie) []xlog.Attr {
	if len(cookies) == 0 {
		return nil
	}
	values := make([]xlog.Attr, len(cookies))
	for idx, cookie := range cookies {
		values[idx] = xlog.String(cookie.Name, cookie.Value)
	}
	return values
}

func (al *AccessLog) headers(h http.Header) []xlog.Attr {
	if len(h) == 0 {
		return nil
	}
	values := make([]xlog.Attr, 0, len(h))
	for key, value := range h {
		values = append(values, xlog.Any(key, value))
	}
	return values
}

func (al *AccessLog) after(ctx context.Context, start time.Time, w http.ResponseWriter, r *http.Request) {
	fields := []xlog.Attr{
		xlog.DurationMS("cost", time.Since(start)),
	}
	if vs, ok := ctx.Value(ctxLogFieldsKey).([]xlog.Attr); ok {
		fields = append(fields, vs...)
	}
	if err := ctx.Err(); err != nil {
		fields = append(fields, xlog.ErrorAttr("after.ctx.err", ctx.Err()))
	}
	al.Logger.Info(ctx, "", fields...)
}
