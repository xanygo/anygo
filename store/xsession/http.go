//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xsession

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xlog"
)

var _ Storage = (*CookieStore)(nil)

// CookieStore 在 cookie 中存储 session 信息
type CookieStore struct {
	// Writer 当前请求对应的 writer,必填，session 信息保存时需要
	Writer http.ResponseWriter

	// Request 当前的请求信息，必填，从此处的读取 cookie 内容
	Request *http.Request

	// CookieName 存储数据的 cookie 的 Name,可选，为空时使用 "session"
	CookieName string

	// Cipher cookie value 的压缩，解压缩方法，可选
	Cipher xcodec.Cipher

	// BeforeSave 可选，用于设置 cookie 的 属性
	BeforeSave func(c *http.Cookie)
}

var defaultCipher = xcodec.Ciphers{
	xcodec.NewCipher(xcodec.GZipCompress, xcodec.GZipDecompress),
	&xcodec.Base64{
		Encoder: base64.RawURLEncoding,
	},
}

func (cs *CookieStore) getCipher() xcodec.Cipher {
	if cs.Cipher != nil {
		return cs.Cipher
	}
	return defaultCipher
}

func (cs *CookieStore) getCookieName() string {
	if cs.CookieName != "" {
		return cs.CookieName
	}
	return "session"
}

func (cs *CookieStore) Get(ctx context.Context, id string) (*Session, error) {
	ck, err := cs.Request.Cookie(cs.getCookieName())
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, xerror.NotFound
		}
		return nil, err
	}

	bf, err := cs.getCipher().Decrypt([]byte(ck.Value))
	if err != nil {
		return nil, err
	}
	val, err := ParserValue(bf)
	if err != nil {
		return nil, err
	}
	return val.ToSession(cs), nil
}

func (cs *CookieStore) GetOrCreate(ctx context.Context, id string) *Session {
	se, err := cs.Get(ctx, id)
	if err == nil {
		return se
	}
	return NewValue(id).ToSession(cs)
}

func (cs *CookieStore) Save(ctx context.Context, session *Session) error {
	bf, err := session.Bytes()
	if err != nil {
		return err
	}
	bf, err = cs.getCipher().Encrypt(bf)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     cs.getCookieName(),
		Value:    string(bf),
		HttpOnly: true,
		Path:     "/",
		Expires:  defaultExpire,
	}
	if cs.BeforeSave != nil {
		cs.BeforeSave(cookie)
	}
	http.SetCookie(cs.Writer, cookie)
	return nil
}

type HTTPHandler struct {
	CookieName string                                           // 在 cookie 中存储 sessionID 的名字，可选，默认为 sid
	OnSet      func(ck *http.Cookie)                            // 在 cookie 中存储 sessionID 的时候回调，可选
	NewStorage func(http.ResponseWriter, *http.Request) Storage // 必填，session 数据存储引擎
	AutoSave   bool                                             // 使用自动对 session 保存
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
		var sid string
		r, sid = s.setSessionID(w, r)
		store := s.NewStorage(w, r)

		session := store.GetOrCreate(r.Context(), sid)
		if s.AutoSave {
			defer store.Save(r.Context(), session)
		}
		session.Set("_", "") // 触发更新

		ctx := WithStorage(r.Context(), store)
		ctx = WithSession(ctx, session)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}
