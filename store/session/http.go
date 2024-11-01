//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package session

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net/http"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xctx"
	"github.com/xanygo/anygo/xerror"
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
	log.Println("session=", string(bf), "err=", err)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     cs.getCookieName(),
		Value:    string(bf),
		HttpOnly: true,
	}
	if cs.BeforeSave != nil {
		cs.BeforeSave(cookie)
	}
	http.SetCookie(cs.Writer, cookie)
	return nil
}

var ctxHTTPStoreKey = xctx.NewKey()

type StorageHTTPHandler struct {
	NewStorage func(http.ResponseWriter, *http.Request) Storage
}

func (hb *StorageHTTPHandler) Next(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store := hb.NewStorage(w, r)
		ctx := context.WithValue(r.Context(), ctxHTTPStoreKey, store)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}

func (hb *StorageHTTPHandler) Store(r *http.Request) Storage {
	return hb.StoreFromContext(r.Context())
}

func (hb *StorageHTTPHandler) StoreFromContext(ctx context.Context) Storage {
	return ctx.Value(ctxHTTPStoreKey).(Storage)
}

func (hb *StorageHTTPHandler) Session(r *http.Request) *Session {
	return hb.Store(r).GetOrCreate(r.Context(), "")
}

func (hb *StorageHTTPHandler) SessionFromContext(ctx context.Context) *Session {
	return hb.StoreFromContext(ctx).GetOrCreate(ctx, "")
}

type CookieStoreHandler struct {
	// CookieName 存储会话信息的 cookie 名称，可选，为空时使用 "session"
	CookieName string

	// Cipher 数据压缩方法，可选
	Cipher xcodec.Cipher

	// BeforeSave 可选，用于设置 cookie 的 属性
	BeforeSave func(c *http.Cookie)
}

func (csh *CookieStoreHandler) Trans() *StorageHTTPHandler {
	return &StorageHTTPHandler{
		NewStorage: csh.newCookieStore,
	}
}

func (csh *CookieStoreHandler) newCookieStore(w http.ResponseWriter, req *http.Request) Storage {
	return &CookieStore{
		Writer:     w,
		Request:    req,
		CookieName: csh.CookieName,
		Cipher:     csh.Cipher,
		BeforeSave: csh.BeforeSave,
	}
}
