//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package session

import (
	"context"
	"errors"
	"net/http"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xerror"
)

var _ Storage = (*CookieStore)(nil)

// CookieStore 在 cookie 中存储 session 信息
type CookieStore struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	CookieName string
	Cipher     xcodec.Cipher

	// BeforeSave 可选，用于设置 cookie 的 属性
	BeforeSave func(c *http.Cookie)
}

func (cs *CookieStore) Get(ctx context.Context, id string) (*Session, error) {
	ck, err := cs.Request.Cookie(cs.CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, xerror.NotFound
		}
		return nil, err
	}

	bf, err := cs.Cipher.Decrypt([]byte(ck.Value))
	if err != nil {
		return nil, err
	}
	val, err := ParserValue(bf)
	if err != nil {
		return nil, err
	}
	return val.ToSession(cs), nil
}

func (cs *CookieStore) GetOrCreate(ctx context.Context, id string) (*Session, error) {
	se, err := cs.Get(ctx, id)
	if err == nil {
		return se, nil
	}
	if !xerror.IsNotFound(err) {
		return nil, err
	}
	val := NewValue(id)
	return val.ToSession(cs), nil
}

func (cs *CookieStore) Save(ctx context.Context, session *Session) error {
	bf, err := session.Bytes()
	if err != nil {
		return err
	}
	bf, err = cs.Cipher.Encrypt(bf)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     cs.CookieName,
		Value:    string(bf),
		HttpOnly: true,
	}
	if cs.BeforeSave != nil {
		cs.BeforeSave(cookie)
	}
	http.SetCookie(cs.Writer, cookie)
	return nil
}
