//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xhttpc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xio"
)

type HandlerFunc func(ctx context.Context, resp *http.Response) error

type handlerCombine []HandlerFunc

func (hfs handlerCombine) Handle(ctx context.Context, resp *http.Response) error {
	for _, f := range hfs {
		if err := f(ctx, resp); err != nil {
			return err
		}
	}
	return nil
}

func Combine(hfs ...HandlerFunc) HandlerFunc {
	return handlerCombine(hfs).Handle
}

func LimitBody(size int64) HandlerFunc {
	return func(ctx context.Context, resp *http.Response) error {
		resp.Body = xio.LimitReaderCloser(resp.Body, size)
		return nil
	}
}

func StatusIn(codes ...int) HandlerFunc {
	mp := make(map[int]bool, len(codes))
	for _, code := range codes {
		mp[code] = true
	}
	return func(ctx context.Context, resp *http.Response) error {
		if !mp[resp.StatusCode] {
			return xerror.NewStatusError(int64(resp.StatusCode))
		}
		return nil
	}
}

func StatusRange(begin int, end int) HandlerFunc {
	return func(ctx context.Context, resp *http.Response) error {
		if resp.StatusCode < begin || resp.StatusCode > end {
			return xerror.NewStatusError(int64(resp.StatusCode))
		}
		return nil
	}
}

func DecodeBody(dc xcodec.Decoder, a any) HandlerFunc {
	return func(ctx context.Context, resp *http.Response) error {
		status := resp.StatusCode
		if status < 200 || status >= 500 {
			return xerror.NewStatusError(int64(status))
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return xcodec.Decode(dc, body, a)
	}
}

func JSONBody(a any) HandlerFunc {
	return DecodeBody(xcodec.JSON, a)
}

func TeeReader(sr *StoredResponse) HandlerFunc {
	start := time.Now()
	if sr.CreateAt == 0 {
		sr.CreateAt = start.Unix()
	}
	return func(ctx context.Context, resp *http.Response) error {
		sr.StatusCode = resp.StatusCode
		sr.Header = resp.Header.Clone()
		sr.Cost = time.Since(start)
		if resp.Request != nil {
			sr.URL = resp.Request.URL.String()
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
		sr.Body = body
		return nil
	}
}

type StoredResponse struct {
	CreateAt   int64         `json:"c"` // 创建时间，unix 时间戳
	StatusCode int           `json:"s"`
	URL        string        `json:"u,omitempty"`
	Header     http.Header   `json:"h,omitempty"`
	Body       []byte        `json:"b,omitempty"`
	Cost       time.Duration `json:"t,omitempty"` // 实际请求耗时
	FromCache  bool          `json:"f,omitempty"` // 是否来自缓存
}

func (sr *StoredResponse) Write(w http.ResponseWriter) {
	for k, vs := range sr.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(sr.Body)))
	w.WriteHeader(sr.StatusCode)
	w.Write(sr.Body)
}

func (sr *StoredResponse) CreateTime() time.Time {
	return time.Unix(sr.CreateAt, 0)
}

// FetchResponse 将 response 返回给传入的 rr
//
// 注意：传入的 rr 应该是已初始化的: rr := &http.Response{}
func FetchResponse(rr *http.Response) HandlerFunc {
	return func(ctx context.Context, resp *http.Response) error {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
		*rr = *resp
		rr.Body = io.NopCloser(bytes.NewBuffer(body))
		return nil
	}
}
