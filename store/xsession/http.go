//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xsession

import (
	"net/http"
	"time"

	"github.com/xanygo/anygo/xhttp"
	"github.com/xanygo/anygo/xlog"
)

// HTTPHandler Session 处理的中间件
//
// 若期望接口不添加 Session，可以在注册的时候同时注册 meta 信息予以标记，具体如下:
// router.Get("/myProxy/{url:*}  meta|session=no", &myProxy{})
type HTTPHandler struct {
	CookieName  string                                           // 在 cookie 中存储 sessionID 的名字，可选，默认为 sid
	OnSet       func(ck *http.Cookie)                            // 在 cookie 中存储 sessionID 的时候回调，可选
	NewStorage  func(http.ResponseWriter, *http.Request) Storage // 必填，session 数据存储引擎
	NeedSession func(req *http.Request) bool                     // 可选，判断本次请求是否需要Session
}

func (s *HTTPHandler) getCookieName() string {
	if s.CookieName != "" {
		return s.CookieName
	}
	return "sid"
}

var defaultExpire = time.Now().AddDate(100, 0, 0)

func (s *HTTPHandler) setSessionID(w http.ResponseWriter, r *http.Request) (*http.Request, string) {
	name := s.getCookieName()
	var id string
	cookie, err := r.Cookie(name)
	if err == nil && len(cookie.Value) > 5 {
		id = cookie.Value
	} else {
		id = NewID()
		sc := &http.Cookie{
			Name:     name,
			Value:    id,
			HttpOnly: true,
			Expires:  defaultExpire,
			Path:     "/",
		}
		if s.OnSet != nil {
			s.OnSet(sc)
		}

		http.SetCookie(w, sc)
	}
	if xlog.IsMetaContext(r.Context()) {
		xlog.AddMetaAttr(r.Context(), xlog.String("sessionID", id))
	}
	ctx := WithID(r.Context(), id)
	return r.WithContext(ctx), id
}

func (s *HTTPHandler) Next(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.NeedSession != nil && !s.NeedSession(r) {
			h.ServeHTTP(w, r)
			return
		}
		routeInfo := xhttp.ReadRouteInfo(r.Context())
		if value, ok := routeInfo.GetMeta("session"); ok && value == "no" {
			h.ServeHTTP(w, r)
			return
		}

		var sid string
		r, sid = s.setSessionID(w, r)
		store := s.NewStorage(w, r)

		session := store.Get(r.Context(), sid)
		session.Set(r.Context(), "_", "") // 触发更新

		ctx := WithStorage(r.Context(), store)
		ctx = WithSession(ctx, session)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}
