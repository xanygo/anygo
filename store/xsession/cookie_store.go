//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-22

package xsession

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xcodec"
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

	cipherGetter xsync.OnceDoValue[xcodec.Cipher]
}

func (cs *CookieStore) getCipher() xcodec.Cipher {
	return cs.cipherGetter.Do(cs.initCipher)
}

func (cs *CookieStore) initCipher() xcodec.Cipher {
	cp := cs.Cipher
	if cp == nil {
		var key string
		if bi, ok := debug.ReadBuildInfo(); ok {
			key = bi.Path
		} else {
			key = "7332d" + "af432078" + "b33dca1d26b" + "431ade36"
		}
		cp = &xcodec.AesOFB{
			Key: key,
		}
	}
	return xcodec.Ciphers{
		xcodec.NewCipher(xcodec.GZipCompress, xcodec.GZipDecompress),
		&xcodec.Base64{
			Encoder: base64.RawURLEncoding,
		},
	}
}

func (cs *CookieStore) getCookieName() string {
	if cs.CookieName != "" {
		return cs.CookieName
	}
	return "session"
}

func (cs *CookieStore) Get(ctx context.Context, id string) Session {
	ck, _ := cs.Request.Cookie(cs.getCookieName())
	var cv string
	if ck != nil {
		cv = ck.Value
	}
	bf, _ := cs.getCipher().Decrypt([]byte(cv))
	session := parserCookieValue(bf)
	session.id = id
	session.store = cs
	return session
}

func (cs *CookieStore) Save(ctx context.Context, session Session) error {
	cv, ok := session.(*cookieSession)
	if !ok {
		return fmt.Errorf("not support session type %T", session)
	}
	bf := cv.bytes()
	bf, err := cs.getCipher().Encrypt(bf)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     cs.getCookieName(),
		Value:    string(bf),
		HttpOnly: true,
		Path:     "/",
		Expires:  defaultExpire,
		SameSite: http.SameSiteLaxMode,
	}
	if cs.BeforeSave != nil {
		cs.BeforeSave(cookie)
	}
	http.SetCookie(cs.Writer, cookie)
	return nil
}

func (cs *CookieStore) Delete() {
	cookie := &http.Cookie{
		Name:     cs.getCookieName(),
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
	}
	http.SetCookie(cs.Writer, cookie)
}

var _ Session = (*cookieSession)(nil)

type cookieSession struct {
	id      string
	created int64
	values  xmap.Sync[string, string]
	store   *CookieStore
}

func (c *cookieSession) ID() string {
	return c.id
}

func (c *cookieSession) Set(ctx context.Context, key string, value string) error {
	c.values.Store(key, value)
	return nil
}

func (c *cookieSession) MSet(ctx context.Context, kv map[string]string) error {
	for k, v := range kv {
		c.values.Store(k, v)
	}
	return nil
}

func (c *cookieSession) Get(ctx context.Context, key string) (string, error) {
	v, _ := c.values.Load(key)
	return v, nil
}

func (c *cookieSession) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	result := make(map[string]string)
	for _, key := range keys {
		v, ok := c.values.Load(key)
		if ok {
			result[key] = v
		}
	}
	return result, nil
}

func (c *cookieSession) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		c.values.Delete(key)
	}
	return nil
}

func (c *cookieSession) Created(ctx context.Context) (time.Time, error) {
	return time.Unix(c.created, 0), nil
}

func (c *cookieSession) Save(ctx context.Context) error {
	return c.store.Save(ctx, c)
}

func (c *cookieSession) Clear(ctx context.Context) error {
	c.values.Clear()
	c.store.Delete()
	return nil
}

func (c *cookieSession) bytes() []byte {
	data := map[string]any{
		"c": c.created,
		"v": c.values.ToMap(),
	}
	bf, _ := json.Marshal(data)
	return bf
}

type cookieValueData struct {
	Created int64             `json:"c"`
	Updated int64             `json:"u"`
	Values  map[string]string `json:"v"`
}

func parserCookieValue(bf []byte) *cookieSession {
	var v cookieValueData
	json.Unmarshal(bf, &v)
	val := &cookieSession{
		created: v.Created,
	}
	for k, v := range v.Values {
		val.values.Store(k, v)
	}
	return val
}
