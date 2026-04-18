//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-15

package xhttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xvalidator"
)

func NewBinder(req *http.Request) *Binder {
	return &Binder{
		req: req,
	}
}

type Binder struct {
	req *http.Request

	once sync.Once
	body []byte
	err  error
}

func (b *Binder) readBody() ([]byte, error) {
	b.once.Do(func() {
		if b.req.Body == nil {
			return
		}

		b.body, b.err = io.ReadAll(b.req.Body)
		b.req.Body = io.NopCloser(bytes.NewBuffer(b.body))
	})

	return b.body, b.err
}

func (b *Binder) Bind(obj any) error {
	contentType := b.req.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		return b.BinJSON(obj)
	}
	return fmt.Errorf("not support Content-Type: %s", contentType)
}

func (b *Binder) BinJSON(obj any) error {
	data, err := b.readBody()
	if err != nil {
		return err
	}
	err = xcodec.Decode(xcodec.JSON, data, obj)
	if err != nil {
		return err
	}
	return xvalidator.Validate(obj)
}

func Bind(r *http.Request, obj any) error {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		err = xcodec.Decode(xcodec.JSON, body, obj)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("not support Content-Type: %s", ct)
	}
	return xvalidator.Validate(obj)
}
